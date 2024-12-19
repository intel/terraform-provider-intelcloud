package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewInstanceDataSource() datasource.DataSource {
	return &instanceDataSource{}
}

type instanceDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &instanceDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type instanceDataSourceModel struct {
	Instances []models.InstanceModel `tfsdk:"instances"`
}

// Configure adds the provider configured client to the data source.
func (d *instanceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *instanceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

func (d *instanceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instances": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"resource_id": schema.StringAttribute{
							Computed: true,
						},
						"cloudaccount": schema.StringAttribute{
							Computed: true,
						},
						"spec": schema.SingleNestedAttribute{
							Required: true,
							Attributes: map[string]schema.Attribute{
								"availability_zone": schema.StringAttribute{
									Optional: true,
								},
								"instance_group": schema.StringAttribute{
									Optional: true,
								},
								"instance_type": schema.StringAttribute{
									Required: true,
								},
								"machine_image": schema.StringAttribute{
									Required: true,
								},
								"ssh_public_key_names": schema.ListAttribute{
									ElementType: types.StringType,
									Required:    true,
								},
								"user_data": schema.StringAttribute{
									Optional: true,
								},
							},
						},
						"interfaces": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"address": schema.StringAttribute{
										Computed: true,
									},
									"dns_name": schema.StringAttribute{
										Computed: true,
									},
									"gateway": schema.StringAttribute{
										Computed: true,
									},
									"name": schema.StringAttribute{
										Computed: true,
									},
									"prefix_length": schema.Int64Attribute{
										Computed: true,
									},
									"subnet": schema.StringAttribute{
										Computed: true,
									},
									"vnet": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"access_info": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"username": types.StringType,
							},
							Computed: true,
						},
						"ssh_proxy": schema.ObjectAttribute{
							AttributeTypes: map[string]attr.Type{
								"address": types.StringType,
								"port":    types.Int64Type,
								"user":    types.StringType,
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

func (d *instanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state instanceDataSourceModel
	state.Instances = []models.InstanceModel{}

	instanceList, err := d.client.GetInstances(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IDC Instances",
			err.Error(),
		)
		return
	}
	for _, inst := range instanceList.Instances {
		instModel := models.InstanceModel{
			Cloudaccount: types.StringValue(inst.Metadata.Cloudaccount),
			Name:         types.StringValue(inst.Metadata.Name),
			ResourceId:   types.StringValue(inst.Metadata.ResourceId),
			Spec: models.InstanceSpec{
				InstanceGroup: types.StringValue(inst.Spec.InstanceGroup),
				InstanceType:  types.StringValue(inst.Spec.InstanceType),
				MachineImage:  types.StringValue(inst.Spec.MachineImage),
				UserData:      types.StringValue(inst.Spec.UserData),
			},
			Status: types.StringValue(inst.Status.Phase),
		}

		for _, k := range inst.Spec.SshPublicKeyNames {
			instModel.Spec.SSHPublicKeyNames = append(instModel.Spec.SSHPublicKeyNames, types.StringValue(k))
		}

		infs := []models.NetworkInterface{}
		for _, nic := range inst.Status.Interfaces {
			// currently we ssume a single interface will have a single address
			addr := ""
			if len(nic.Addresses) > 0 {
				addr = nic.Addresses[0]
			}
			inf := models.NetworkInterface{
				Addresses:    types.StringValue(addr),
				DNSName:      types.StringValue(nic.DNSName),
				Gateway:      types.StringValue(nic.Gateway),
				Name:         types.StringValue(nic.Name),
				PrefixLength: types.Int64Value(int64(nic.PrefixLength)),
				Subnet:       types.StringValue(nic.Subnet),
				VNet:         types.StringValue(nic.VNet),
			}
			infs = append(infs, inf)
		}
		infObject, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.ProviderInterfaceAttributes), infs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		instModel.Interfaces = infObject

		accessInfoMap := models.InstanceAccessInfoModel{
			Username: types.StringValue(inst.Status.UserName),
		}

		accessObj, diags := types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		instModel.AccessInfo = accessObj

		sshProxyMap := models.SSHProxyModel{
			ProxyAddress: types.StringValue(inst.Status.SSHProxy.Address),
			ProxyPort:    types.Int64Value(inst.Status.SSHProxy.Port),
			ProxyUser:    types.StringValue(inst.Status.SSHProxy.User),
		}
		sshProxyObj, diags := types.ObjectValueFrom(ctx, sshProxyMap.AttributeTypes(), sshProxyMap)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		instModel.SSHProxy = sshProxyObj

		state.Instances = append(state.Instances, instModel)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
