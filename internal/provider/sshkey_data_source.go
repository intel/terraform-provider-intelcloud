package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewSSHKeysDataSource() datasource.DataSource {
	return &sshkeysDataSource{}
}

type sshkeysDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &sshkeysDataSource{}
	_ datasource.DataSourceWithConfigure = &sshkeysDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type sshkeysDataSourceModel struct {
	SSHKeys []sshkeyModel `tfsdk:"sshkeys"`
}

// coffeesModel maps coffees schema data.
type sshkeyModel struct {
	Metadata resourceMetadata `tfsdk:"metadata"`
	Spec     sshkeySpec       `tfsdk:"spec"`
}

type resourceMetadata struct {
	ResourceId   types.String `tfsdk:"resourceid"`
	Cloudaccount types.String `tfsdk:"cloudaccount"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"createdat"`
}
type sshkeySpec struct {
	SSHPublicKey types.String `tfsdk:"ssh_public_key"`
	OwnerEmail   types.String `tfsdk:"owner_email"`
}

// Configure adds the provider configured client to the data source.
func (d *sshkeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sshkeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshkey"
}

func (d *sshkeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"sshkeys": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"metadata": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"resourceid": schema.StringAttribute{
									Computed: true,
								},
								"cloudaccount": schema.StringAttribute{
									Computed: true,
								},
								"name": schema.StringAttribute{
									Computed: true,
								},
								"createdat": schema.StringAttribute{
									Computed: true,
								},
							},
						},
						"spec": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"ssh_public_key": schema.StringAttribute{
									Computed: true,
								},
								"owner_email": schema.StringAttribute{
									Computed: true,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *sshkeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state sshkeysDataSourceModel
	state.SSHKeys = []sshkeyModel{}

	sshkeyList, err := d.client.GetSSHKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IDC Compute SSHKeys",
			err.Error(),
		)
		return
	}
	for _, key := range sshkeyList.SSHKey {
		sshkeyModel := sshkeyModel{
			Metadata: resourceMetadata{
				Cloudaccount: types.StringValue(key.Metadata.Cloudaccount),
				Name:         types.StringValue(key.Metadata.Name),
			},
			Spec: sshkeySpec{
				SSHPublicKey: types.StringValue(key.Spec.SSHPublicKey),
				OwnerEmail:   types.StringValue(key.Spec.OwnerEmail),
			},
		}
		state.SSHKeys = append(state.SSHKeys, sshkeyModel)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
