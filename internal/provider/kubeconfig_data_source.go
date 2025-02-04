package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewKubeconfigDataSource() datasource.DataSource {
	return &kubeconfigDataSource{}
}

type kubeconfigDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &kubeconfigDataSource{}
	_ datasource.DataSourceWithConfigure = &kubeconfigDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type kubeconfigDataSourceModel struct {
	ClusterUUID types.String `tfsdk:"cluster_uuid"`
	Kubeconfig  types.String `tfsdk:"kubeconfig"`
}

// Configure adds the provider configured client to the data source.
func (d *kubeconfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *kubeconfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubeconfig"
}

func (d *kubeconfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_uuid": schema.StringAttribute{
				Required: true,
			},
			"kubeconfig": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *kubeconfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state kubeconfigDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	kubeconfigStr, err := d.client.GetClusterKubeconfig(ctx, state.ClusterUUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve ITAC IKS kubeconfig",
			err.Error(),
		)
		return
	}

	state.Kubeconfig = types.StringValue(*kubeconfigStr)

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
