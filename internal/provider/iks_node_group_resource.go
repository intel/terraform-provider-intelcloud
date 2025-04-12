package provider

import (
	"context"
	"fmt"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &iksNodeGroupResource{}
	_ resource.ResourceWithConfigure = &iksNodeGroupResource{}
)

// iksNodeGroupResourceModel maps the resource schema data.
type iksNodeGroupResourceModel struct {
	ClusterUUID       types.String   `tfsdk:"cluster_uuid"`
	ID                types.String   `tfsdk:"id"`
	Count             types.Int64    `tfsdk:"node_count"`
	Name              types.String   `tfsdk:"name"`
	NodeType          types.String   `tfsdk:"node_type"`
	IMIId             types.String   `tfsdk:"imiid"`
	State             types.String   `tfsdk:"state"`
	UserDataURL       types.String   `tfsdk:"userdata_url"`
	SSHPublicKeyNames []types.String `tfsdk:"ssh_public_key_names"`
	Vnets             types.List     `tfsdk:"vnets"`
}

// NewOrderKubernetes is a helper function to simplify the provider implementation.
func NewIKSNodeGroupResource() resource.Resource {
	return &iksNodeGroupResource{}
}

// orderIKSNodeGroup is the resource implementation.
type iksNodeGroupResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *iksNodeGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the resource type name.
func (r *iksNodeGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iks_node_group"
}

// Schema defines the schema for the resource.
func (r *iksNodeGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"cluster_uuid": schema.StringAttribute{
				Required: true,
			},
			"node_count": schema.Int64Attribute{
				Required: true,
			},
			"node_type": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"imiid": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"userdata_url": schema.StringAttribute{
				Optional: true,
			},
			"ssh_public_key_names": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"vnets": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availabilityzonename": schema.StringAttribute{
							Computed: true,
						},
						"networkinterfacevnetname": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *iksNodeGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iksNodeGroupResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inArg := itacservices.IKSNodeGroupCreateRequest{
		Name:           plan.Name.ValueString(),
		Count:          plan.Count.ValueInt64(),
		ProductType:    "iks-cluster",
		InstanceTypeId: plan.NodeType.ValueString(),
		UserDataURL:    plan.UserDataURL.ValueString(),
	}

	for _, k := range plan.SSHPublicKeyNames {
		inArg.SSHKeyNames = append(inArg.SSHKeyNames, itacservices.SKey{Name: k.ValueString()})
	}

	tflog.Info(ctx, "making a call to IDC Service to createVnetIfNotExist")
	vnetResp, err := r.client.CreateVNetIfNotFound(ctx, *r.client.Region)
	if err != nil || vnetResp == nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	inArg.Vnets = append(inArg.Vnets,
		itacservices.Vnet{
			AvailabilityZoneName:     vnetResp.Spec.AvailabilityZone,
			NetworkInterfaceVnetName: vnetResp.Metadata.Name,
		})

	nodeGroupResp, _, err := r.client.CreateIKSNodeGroup(ctx, &inArg, plan.ClusterUUID.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating iks node group",
			"Could not create iks node group, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(nodeGroupResp.ID)
	plan.IMIId = types.StringValue(nodeGroupResp.IMIID)
	plan.State = types.StringValue(nodeGroupResp.State)
	vnets := []models.NetworkInterfaceSpec{}
	for _, iface := range nodeGroupResp.Interfaces {
		v := models.NetworkInterfaceSpec{
			AvailabilityZoneName:     types.StringValue(iface.AvailabilityZoneName),
			NetworkInterfaceVnetName: types.StringValue(iface.NetworkInterfaceVnetName),
		}
		vnets = append(vnets, v)
	}

	vnetObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.VnetAttributes), vnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Vnets = vnetObj

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *iksNodeGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state iksNodeGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from IDC Service
	ngState, _, err := r.client.GetIKSNodeGroupByID(ctx, state.ClusterUUID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Compute IKS Node Group resource",
			"Could not read IDC Compute IKS Node Group resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.IMIId = types.StringValue(ngState.IMIID)
	state.State = types.StringValue(ngState.State)

	vnets := []models.NetworkInterfaceSpec{}
	for _, i := range ngState.Interfaces {
		v := models.NetworkInterfaceSpec{
			AvailabilityZoneName:     types.StringValue(i.AvailabilityZoneName),
			NetworkInterfaceVnetName: types.StringValue(i.NetworkInterfaceVnetName),
		}
		vnets = append(vnets, v)
	}

	vnetObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.VnetAttributes), vnets)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Vnets = vnetObj

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *iksNodeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *iksNodeGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state iksNodeGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the order from IDC Services
	err := r.client.DeleteIKSNodeGroup(ctx, state.ClusterUUID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC IKS node group resource",
			"Could not delete IDC KS node group resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}
