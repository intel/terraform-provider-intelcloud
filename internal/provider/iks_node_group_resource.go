package provider

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	Timeouts          *timeoutsModel `tfsdk:"timeouts"`
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
				Optional: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"userdata_url": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"ssh_public_key_names": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"vnets": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"availabilityzonename": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"networkinterfacevnetname": schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"resource_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timeout for nodegroup resource operations",
						Default:     stringdefault.StaticString(IKSNodegroupResourceName),
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

	// use timeouts if requested by the user
	createTimeout, err := plan.Timeouts.GetTimeouts(IKSNodegroupResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse create timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

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

	if !plan.IMIId.IsNull() && !plan.IMIId.IsUnknown() {
		inArg.WorkerImiId = plan.IMIId.ValueString()
	}

	nodeGroupResp, _, err := r.client.CreateIKSNodeGroup(ctx, &inArg, plan.ClusterUUID.ValueString(), false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating iks node group",
			"Could not create iks node group, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ClusterUUID = types.StringValue(nodeGroupResp.ClusterID)
	plan.Name = types.StringValue(nodeGroupResp.Name)
	plan.ID = types.StringValue(nodeGroupResp.ID)
	plan.IMIId = types.StringValue(nodeGroupResp.IMIID)
	plan.State = types.StringValue(nodeGroupResp.State)
	plan.Count = types.Int64Value(nodeGroupResp.Count)
	plan.NodeType = types.StringValue(nodeGroupResp.InstanceType)
	plan.UserDataURL = types.StringValue(nodeGroupResp.UserDataURL)
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

	plan.SSHPublicKeyNames = []types.String{}
	for _, k := range nodeGroupResp.SSHKeyNames {
		plan.SSHPublicKeyNames = append(plan.SSHPublicKeyNames, types.StringValue(k.Name))
	}

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

	// use timeouts if requested by the user
	readTimeout, err := state.Timeouts.GetTimeouts(IKSNodegroupResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// Get refreshed order value from IDC Service
	ngState, err := r.client.GetIKSNodeGroupByID(ctx, state.ClusterUUID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Compute IKS Node Group resource",
			"Could not read IDC Compute IKS Node Group resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ClusterUUID = types.StringValue(ngState.ClusterID)
	state.Name = types.StringValue(ngState.Name)
	state.ID = types.StringValue(ngState.ID)
	state.IMIId = types.StringValue(ngState.IMIID)
	state.State = types.StringValue(ngState.State)
	state.Count = types.Int64Value(ngState.Count)
	state.NodeType = types.StringValue(ngState.InstanceType)
	state.UserDataURL = types.StringValue(ngState.UserDataURL)

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

	state.SSHPublicKeyNames = []types.String{}
	for _, k := range ngState.SSHKeyNames {
		state.SSHPublicKeyNames = append(state.SSHPublicKeyNames, types.StringValue(k.Name))
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *iksNodeGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state iksNodeGroupResourceModel

	// Retrieve the desired configuration from the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve the current state
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	updateTimeout, err := state.Timeouts.GetTimeouts(IKSNodegroupResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse update timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if !plan.Count.Equal(state.Count) {
		tflog.Info(ctx, "Detected change in iks node group spec for count, updating node group",
			map[string]any{"current count ": state.Count.ValueInt64(), "new count": plan.Count.ValueInt64()})
		inArg := itacservices.UpdateNodeGroupRequest{
			ClusterId:   state.ClusterUUID.ValueString(),
			NodeGroupId: state.ID.ValueString(),
			Count:       plan.Count.ValueInt64(),
		}
		err := r.client.UpdateNodeGroup(ctx, &inArg)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating node group order",
				"Could not update nodegroup, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Get refreshed order value from IDC Service irrespective of whether upgrade was done or skipped
	nodeGroup, err := r.client.GetIKSNodeGroupByID(ctx, state.ClusterUUID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IKS nodegroup resource",
			"Could not read IKS nodegroup resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	currState, err := refreshIKSNodegroupResourceModel(ctx, nodeGroup)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IKS cluster resource",
			"Could not read IKS cluster resource ID "+plan.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	// set timeout again for consistency
	currState.Timeouts = plan.Timeouts

	// Set refreshed state
	diags = resp.State.Set(ctx, currState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *iksNodeGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	// Expect import ID in the format: cluster_id:id
	ids := strings.Split(req.ID, ":")
	if len(ids) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import format",
			"Expected import ID in the format 'cluster_id:id'. Example: abc123:def456",
		)
		return
	}

	clusterID := ids[0]
	nodegroupId := ids[1]

	// Set both attributes in state
	resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), clusterID)
	resp.State.SetAttribute(ctx, path.Root("id"), nodegroupId)
	//resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

	// use timeouts if requested by the user
	deleteTimeout, err := state.Timeouts.GetTimeouts(IKSNodegroupResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse delete timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the order from IDC Services
	err = r.client.DeleteIKSNodeGroup(ctx, state.ClusterUUID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC IKS node group resource",
			"Could not delete IDC KS node group resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func refreshIKSNodegroupResourceModel(ctx context.Context, nodegroup *itacservices.NodeGroup) (*iksNodeGroupResourceModel, error) {
	state := &iksNodeGroupResourceModel{}

	state.ID = types.StringValue(nodegroup.ID)
	state.ClusterUUID = types.StringValue(nodegroup.ClusterID)
	state.Name = types.StringValue(nodegroup.Name)
	state.Count = types.Int64Value(nodegroup.Count)
	state.State = types.StringValue(nodegroup.State)
	state.IMIId = types.StringValue(nodegroup.IMIID)
	state.UserDataURL = types.StringValue(nodegroup.UserDataURL)
	state.NodeType = types.StringValue(nodegroup.InstanceType)
	state.SSHPublicKeyNames = []types.String{}
	for _, k := range nodegroup.SSHKeyNames {
		state.SSHPublicKeyNames = append(state.SSHPublicKeyNames, types.StringValue(k.Name))
	}
	vnets := []models.NetworkInterfaceSpec{}
	for _, iface := range nodegroup.Interfaces {
		v := models.NetworkInterfaceSpec{
			AvailabilityZoneName:     types.StringValue(iface.AvailabilityZoneName),
			NetworkInterfaceVnetName: types.StringValue(iface.NetworkInterfaceVnetName),
		}
		vnets = append(vnets, v)
	}
	vnetObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.VnetAttributes), vnets)
	if diags.HasError() {
		return state, fmt.Errorf("error parsing values")
	}
	state.Vnets = vnetObj
	if diags.HasError() {
		return state, fmt.Errorf("error parsing values")
	}

	return state, nil
}
