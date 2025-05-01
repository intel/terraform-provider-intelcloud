package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewFilesystemsDataSource() datasource.DataSource {
	return &filesystemsDataSource{}
}

type filesystemsDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &filesystemsDataSource{}
	_ datasource.DataSourceWithConfigure = &filesystemsDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type filesystemsDataSourceModel struct {
	Filesystems []models.FilesystemModel `tfsdk:"filesystems"`
}

// Configure adds the provider configured client to the data source.
func (d *filesystemsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *filesystemsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filesystems"
}

func (d *filesystemsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"filesystems": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"resource_id": schema.StringAttribute{
							Computed: true,
						},
						"cloudaccount": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Optional: true,
						},
						"availability_zone": schema.StringAttribute{
							Computed: true,
						},
						"spec": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"size_in_tb": schema.Int64Attribute{
									Computed: true,
								},
								"access_mode": schema.StringAttribute{
									Computed: true,
								},
								"filesystem_type": schema.StringAttribute{
									Computed: true,
								},
								"storage_class": schema.StringAttribute{
									Computed: true,
								},
								"encrypted": schema.BoolAttribute{
									Computed: true,
								},
							},
						},
						"cluster_info": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"cluster_address": types.StringType,
								"cluster_version": types.StringType,
							},
							Computed: true,
						},
						"access_info": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"namespace":       types.StringType,
								"filesystem_name": types.StringType,
								"username":        types.StringType,
								"password":        types.StringType,
							},
							Computed: true,
						},
						"status": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *filesystemsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state filesystemsDataSourceModel
	state.Filesystems = []models.FilesystemModel{}

	var diags diag.Diagnostics
	fsList, err := d.client.GetFilesystems(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IDC Filesystems",
			err.Error(),
		)
		return
	}

	for _, fs := range fsList.FilesystemList {
		sizeStr := strings.Split(fs.Spec.Request.Size, "GB")[0]
		size, _ := strconv.ParseInt(sizeStr, 10, 64)
		fsModel := models.FilesystemModel{
			Cloudaccount:     types.StringValue(fs.Metadata.Cloudaccount),
			Name:             types.StringValue(fs.Metadata.Name),
			Description:      types.StringValue(fs.Metadata.Description),
			ResourceId:       types.StringValue(fs.Metadata.ResourceId),
			AvailabilityZone: types.StringValue(fs.Spec.AvailabilityZone),
			Spec: models.FilesystemSpec{
				Size:           types.Int64Value(size),
				AccessMode:     types.StringValue(fs.Spec.AccessMode),
				FilesystemType: types.StringValue(fs.Spec.FilesystemType),
				StorageClass:   types.StringValue(fs.Spec.StorageClass),
				Encrypted:      types.BoolValue(fs.Spec.Encrypted),
			},

			Status: types.StringValue(fs.Status.Phase),
		}
		fsModel.Status = types.StringValue(mapFilesystemStatus(fs.Status.Phase))

		clusterInfoMap := models.FilesystemClusteModel{
			ClusterAddress: types.StringValue(fs.Status.Mount.ClusterAddr),
			ClusterVersion: types.StringValue(fs.Status.Mount.ClusterVersion),
		}

		accessInfoMap := models.FilesystemAccessModel{
			Namespace:  types.StringValue(fs.Status.Mount.Namespace),
			Filesystem: types.StringValue(fs.Status.Mount.FilesystemName),
			Username:   types.StringValue(fs.Status.Mount.UserName),
			Password:   types.StringValue(fs.Status.Mount.Password),
		}

		fsModel.ClusterInfo, diags = types.ObjectValueFrom(ctx, clusterInfoMap.AttributeTypes(), clusterInfoMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		fsModel.AccessInfo, diags = types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.Filesystems = append(state.Filesystems, fsModel)
	}

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
