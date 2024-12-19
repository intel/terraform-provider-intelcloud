package provider

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &iksLBResource{}
	_ resource.ResourceWithConfigure = &iksLBResource{}
)

// orderIKSNodeGroupModel maps the resource schema data.
type iksLBResourceModel struct {
	ClusterUUID   types.String             `tfsdk:"cluster_uuid"`
	LoadBalancers []models.IKSLoadBalancer `tfsdk:"load_balancers"`
}

// NewIKSLB is a helper function to simplify the provider implementation.
func NewIKSLBResource() resource.Resource {
	return &iksLBResource{}
}

// orderIKSNodeGroup is the resource implementation.
type iksLBResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *iksLBResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*itacservices.IDCServicesClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *idcservices.IDCServicesClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *iksLBResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iks_lb"
}

// Schema defines the schema for the resource.
func (r *iksLBResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cluster_uuid": schema.StringAttribute{
				Required: true,
			},
			"load_balancers": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Required: true,
						},
						"vip_state": schema.StringAttribute{
							Computed: true,
						},
						"vip_ip": schema.StringAttribute{
							Computed: true,
						},
						"port": schema.Int64Attribute{
							Required: true,
						},
						"pool_port": schema.Int64Attribute{
							Computed: true,
						},
						"vip_type": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *iksLBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iksLBResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for idx := range plan.LoadBalancers {
		inArg := itacservices.IKSLoadBalancerRequest{
			Name:    plan.LoadBalancers[idx].Name.ValueString(),
			Port:    int(plan.LoadBalancers[idx].Port.ValueInt64()),
			VIPType: plan.LoadBalancers[idx].VipType.ValueString(),
		}

		ilbResp, _, err := r.client.CreateIKSLoadBalancer(ctx, &inArg, plan.ClusterUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating iks load balancer",
				"Could not create iks load balancer, unexpected error: "+err.Error(),
			)
			return
		}

		plan.LoadBalancers[idx].ID = types.StringValue(strconv.FormatInt(ilbResp.ID, 64))
		plan.LoadBalancers[idx].PoolPort = types.Int64Value(int64(ilbResp.PoolPort))
		plan.LoadBalancers[idx].VipState = types.StringValue(ilbResp.VIPState)
		plan.LoadBalancers[idx].VipIp = types.StringValue(ilbResp.VIPIP)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *iksLBResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state iksLBResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for idx, lb := range state.LoadBalancers {
		vipIdNum, _ := strconv.ParseInt(lb.ID.ValueString(), 10, 64)
		refreshedState, err := r.client.GetIKSLoadBalancerByID(ctx, state.ClusterUUID.ValueString(), vipIdNum)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading IDC Compute IKS Load Balancer resource",
				"Could not read IDC Compute IKS Load Balancer resource ID "+state.ClusterUUID.ValueString()+": "+err.Error(),
			)
			return
		}
		state.LoadBalancers[idx].PoolPort = types.Int64Value(int64(refreshedState.PoolPort))
		state.LoadBalancers[idx].VipIp = types.StringValue(refreshedState.VIPIP)
		state.LoadBalancers[idx].VipState = types.StringValue(refreshedState.VIPState)
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *iksLBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *iksLBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
