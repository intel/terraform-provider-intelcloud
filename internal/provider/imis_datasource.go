package provider

import (
	"context"
	"fmt"
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
	Filters []KVFilter         `tfsdk:"filters"`
	Result  *models.ImisModel  `tfsdk:"result"`
	Imis    []models.ImisModel `tfsdk:"items"`
}

// Configure adds the provider configured client to the data source.
func (d *imisDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
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

func (d *imisDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
					"workerImi": schema.ListNestedAttribute{
						Computed: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"imiName": schema.StringAttribute{
									Computed: true,
								},
								"info": schema.StringAttribute{
									Computed: true,
								},
								"isDefaultImi": schema.BoolAttribute{
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
						"workerImi": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"imiName": schema.StringAttribute{
										Computed: true,
									},
									"info": schema.StringAttribute{
										Computed: true,
									},
									"isDefaultImi": schema.BoolAttribute{
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

func (d *imisDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_imis"
}

func (d *imisDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state imisDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceImis, err := d.client.GetImis(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ITAC Imis'",
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
	filteredImages := filterImis(allImis, state.Filters)

	state.Imis = append(state.Imis, filteredImages...)
	state.Result = &filteredImages[0]

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func filterImis(allImis []models.ImisModel, filters []KVFilter) []models.ImisModel {
	filteredImages := allImis

	for _, filter := range filters {
		switch filter.Key {
		case "imis-name":
			filteredImages = filterImisByName(filteredImages, filter.Values)
		case "instance-type":
			filteredImages = filterByInstanceType(filteredImages, filter.Values)
		default:
			return allImis
		}
	}
	return filteredImages
}

func filterImisByName(allImis []models.ImisModel, values []string) []models.ImisModel {
	filteredImis := []models.ImisModel{}
	for _, v := range values {
		for _, imi := range allImis {
			for _, workerImi := range imi.WorkerImi {
				if strings.Contains(workerImi.ImiName.ValueString(), v) {
					filteredImis = append(filteredImis, imi)
				}
			}
		}
	}
	return filteredImis
}

func filterByInstanceType(allImis []models.ImisModel, values []string) []models.ImisModel {
	filteredImages := []models.ImisModel{}
	for _, v := range values {
		for _, imi := range allImis {
			if strings.Contains(imi.InstanceTypeName, v) {
				filteredImages = append(filteredImages, imi)
			}
		}
	}
	return filteredImages
}
