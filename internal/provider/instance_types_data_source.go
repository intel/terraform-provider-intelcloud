package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewInstanceTypesDataSource() datasource.DataSource {
	return &instanceTypesDataSource{}
}

type instanceTypesDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &instanceTypesDataSource{}
	_ datasource.DataSourceWithConfigure = &instanceTypesDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type instanceTypesDataSourceModel struct {
	InstanceTypes []models.InstanceType `tfsdk:"instance_types"`
}

// Configure adds the provider configured client to the data source.
func (d *instanceTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *instanceTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance_types"
}

func (d *instanceTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"instance_types": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"instance_category": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *instanceTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state instanceTypesDataSourceModel

	instTypes, err := d.client.GetInstanceTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IDC Instance types",
			err.Error(),
		)
		return
	}

	for _, t := range instTypes.Items {
		ifInst := models.InstanceType{
			Name:             types.StringValue(t.Metadata.Name),
			Description:      types.StringValue(t.Spec.Description),
			InstanceCategory: types.StringValue(t.Spec.InstanceCategory),
		}
		state.InstanceTypes = append(state.InstanceTypes, ifInst)
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
