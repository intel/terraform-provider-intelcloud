package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewImisDataSource() datasource.DataSource {
	return &imisDataSource{}
}

type imisDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &imisDataSource{}
	_ datasource.DataSourceWithConfigure = &imisDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type imisDataSourceModel struct {
	Latest      types.Bool         `tfsdk:"latest"`
	ClusterUUID types.String       `tfsdk:"clusteruuid"`
	Filters     []KVFilter         `tfsdk:"filters"`
	Result      *models.ImisModel  `tfsdk:"result"`
	Imis        []models.ImisModel `tfsdk:"items"`
}

// Configure adds the provider configured client to the data source.
func (d *imisDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		fmt.Println("[DEBUG] ProviderData is nil")
		return
	}
	client, ok := req.ProviderData.(*itacservices.IDCServicesClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *itacservices.IDCServicesClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *imisDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_imis"
}

func (d *imisDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"clusteruuid": schema.StringAttribute{
				Required:    true,
				Description: "The UUID of the cluster.",
			},
			"latest": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, only the latest IMI will be returned.",
				Computed:    true,
			},
			"filters": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{ // maps to KVFilter.Key
							Required: true,
						},
						"values": schema.ListAttribute{ // maps to KVFilter.Values
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
			"result": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"instancetypename": schema.StringAttribute{
						Computed: true,
					},
					"workerimi": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"iminame": schema.StringAttribute{
									Computed: true,
								},
								"info": schema.StringAttribute{
									Computed: true,
								},
								"isdefaultimi": schema.BoolAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"instancetypename": schema.StringAttribute{
							Computed: true,
						},
						"workerimi": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"iminame": schema.StringAttribute{
										Computed: true,
									},
									"info": schema.StringAttribute{
										Computed: true,
									},
									"isdefaultimi": schema.BoolAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *imisDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state imisDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.ClusterUUID.IsUnknown() || state.ClusterUUID.IsNull() {
		resp.Diagnostics.AddError("Missing clusteruuid", "The 'clusteruuid' field is required.")
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError("client is nil", "The client is not configured. Please check your provider configuration.")
		return
	}

	instanceImis, err := d.client.GetImis(ctx, state.ClusterUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read ITAC Imis: %s", state.ClusterUUID.ValueString()),
			err.Error(),
		)
		return
	}

	allImis := []models.ImisModel{}

	for _, imi := range instanceImis.InstanceTypes {
		tfImi := models.ImisModel{
			InstanceTypeName: imi.Name,
			WorkerImi:        []models.WorkerImi{},
		}
		for _, workerImis := range imi.WorkerImi {
			tfImi.WorkerImi = append(tfImi.WorkerImi, models.WorkerImi{
				ImiName:      types.StringValue(workerImis.ImiName),
				Info:         types.StringValue(workerImis.Info),
				IsDefaultImi: types.BoolValue(workerImis.IsDefaultImi),
			})
		}
		allImis = append(allImis, tfImi)
	}
	filteredImages := filterImis(allImis, state.Filters, state.Latest.ValueBool())

	state.Imis = append(state.Imis, filteredImages...)
	if len(filteredImages) > 0 {
		state.Imis = filteredImages
		state.Result = &filteredImages[0]
	} else {
		state.Imis = []models.ImisModel{}
		state.Result = nil
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func filterImis(allImis []models.ImisModel, filters []KVFilter, latest bool) []models.ImisModel {
	filteredImages := allImis

	for _, filter := range filters {
		switch filter.Key {
		case "instance-type":
			filteredImages = filterByInstanceType(filteredImages, filter.Values, latest)
		default:
			return allImis
		}
	}
	return filteredImages
}

func filterByInstanceType(allImis []models.ImisModel, values []string, latest bool) []models.ImisModel {
	filteredImages := []models.ImisModel{}
	for _, v := range values {
		for _, imi := range allImis {
			if strings.Contains(imi.InstanceTypeName, v) {
				filteredImages = append(filteredImages, imi)
			}
		}
	}
	if latest && len(filteredImages) > 0 {
		getLatestImi(filteredImages)
	}
	return filteredImages
}

// getLatestImi sorts the WorkerImi slice in each ImisModel based on ImiName and returns latest IMI
func getLatestImi(imisList []models.ImisModel) {
	for i := range imisList {
		sort.Slice(imisList[i].WorkerImi, func(a, b int) bool {
			return imisList[i].WorkerImi[a].ImiName.ValueString() < imisList[i].WorkerImi[b].ImiName.ValueString()
		})
	}
	imisList[0].WorkerImi = imisList[0].WorkerImi[:1]
}
