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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &iksLBResource{}
	_ resource.ResourceWithConfigure   = &iksLBResource{}
	_ resource.ResourceWithImportState = &iksLBResource{}
)

// orderIKSNodeGroupModel maps the resource schema data.
type iksLoadBalancerResourceModel struct {
	ClusterUUID  types.String           `tfsdk:"cluster_uuid"`
	LoadBalancer models.IKSLoadBalancer `tfsdk:"load_balancers"`
	Timeouts     *timeoutsModel         `tfsdk:"timeouts"`
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
		},
		Blocks: map[string]schema.Block{
			"load_balancers": schema.SingleNestedBlock{
				Description: "List of load balancers to be provisioned.",
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required:    true,
						Description: "Name of the load balancer.",
					},
					"schema": schema.StringAttribute{
						Required:    true,
						Description: "Schema under which the load balancer is created.",
					},
				},
				Blocks: map[string]schema.Block{
					"listeners": schema.ListNestedBlock{
						Description: "List of listener configurations.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"port": schema.Int64Attribute{
									Required:    true,
									Description: "Listener port.",
								},
								"protocol": schema.StringAttribute{
									Required:    true,
									Description: "Listener protocol (e.g., LBProtocolTCP).",
								},
							},
							Blocks: map[string]schema.Block{
								"pool": schema.SingleNestedBlock{
									Description: "Pool configuration for the listener.",
									Attributes: map[string]schema.Attribute{
										"port": schema.Int64Attribute{
											Required:    true,
											Description: "Pool port.",
										},
										"monitor": schema.StringAttribute{
											Required:    true,
											Description: "Health monitor type (e.g., https).",
										},
										"load_balancing_mode": schema.StringAttribute{
											Required:    true,
											Description: "Load balancing mode (e.g., roundRobin).",
										},
										"node_group_id": schema.StringAttribute{
											Required:    true,
											Description: "ID of the associated node group.",
										},
									},
								},
								"security": schema.SingleNestedBlock{
									Description: "Security configuration for the load balancer listener.",
									Attributes: map[string]schema.Attribute{
										"source_ips": schema.ListAttribute{
											ElementType: types.StringType,
											Required:    true,
											Description: "List of allowed source IPs.",
										},
									},
								},
							},
						},
					},
					"security": schema.SingleNestedBlock{
						Description: "Security configuration for the load balancer.",
						Attributes: map[string]schema.Attribute{
							"source_ips": schema.ListAttribute{
								ElementType: types.StringType,
								Required:    true,
								Description: "List of allowed source IPs.",
							},
						},
					},
				},
			},
			"timeouts": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"resource_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timeout for loadbalancer resource operations",
						Default:     stringdefault.StaticString(IKSLoadBalancerResourceName),
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *iksLBResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iksLoadBalancerResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	createTimeout, err := plan.Timeouts.GetTimeouts(IKSLoadBalancerResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse create timeout for loadbalancer: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	inArg := itacservices.IKSLoadbalancerCreateRequest{
		Metadata: itacservices.IKSLoadBalancerCreateMetadata{
			Name:        plan.LoadBalancer.Name.ValueString(),
			ClusterUUID: plan.ClusterUUID.ValueString(),
		},
	}
	for _, listener := range plan.LoadBalancer.Listeners {
		inArg.Spec.Listeners = append(inArg.Spec.Listeners, itacservices.IKSLoadBalancerListener{
			Port:     listener.Port.ValueInt64(),
			Protocol: listener.Protocol.ValueString(),
			Pool: itacservices.IKSLoadBalancerPool{
				Port:              listener.Pool.Port.ValueInt64(),
				Monitor:           listener.Pool.Monitor.ValueString(),
				LoadBalancingMode: listener.Pool.LoadBalancingMode.ValueString(),
				NodeGroupID:       listener.Pool.NodeGroupId.ValueString(),
			},
			Security: itacservices.IKSLoadBalancerSecurity{
				SourceIps: convertTFStringsToGoStrings(listener.Security.SourceIps),
			},
		})
		inArg.Spec.Security.SourceIps = convertTFStringsToGoStrings(plan.LoadBalancer.Security.SourceIps)
		inArg.Spec.Schema = plan.LoadBalancer.Schema.ValueString()

		ilbResp, _, err := r.client.CreateIKSLoadBalancer(ctx, &inArg, plan.ClusterUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating iks load balancer",
				"Could not create iks load balancer, unexpected error: "+err.Error(),
			)
			return
		}

		plan.LoadBalancer.ID = types.StringValue(ilbResp.Metadata.ResourceID)
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
	var state *iksLoadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	readTimeout, err := state.Timeouts.GetTimeouts(IKSLoadBalancerResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout for loadbalancer: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	currState, err := r.refreshIKSLoadBalancerResourceModel(ctx, state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IKS Load Balancer resource",
			"Could not read IKS Load Balancer for cluster ID "+state.ClusterUUID.String()+": "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, currState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *iksLBResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state iksLoadBalancerResourceModel

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
	updateTimeout, err := state.Timeouts.GetTimeouts(IKSLoadBalancerResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse update timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	lb := plan.LoadBalancer
	// Check if the load balancer already exists in the state
	lbExists, lbID := r.checkLBExistsAndGetID(ctx, plan.ClusterUUID.ValueString(), lb.Name.ValueString())
	if lbExists {
		// Update existing load balancer
		tflog.Info(ctx, "Updating existing IKS Load Balancer", map[string]any{"Name": lb.Name.ValueString()})
		inArg := itacservices.IKSLoadBalancerUpdateRequest{
			Metadata: itacservices.IKSLoadBalancerUpdateMetadata{
				ResourceId: lb.ID.ValueString(),
			},
			Spec: itacservices.IKSLoadBalancerUpdateSpec{
				Security: itacservices.IKSLoadBalancerSecurity{
					SourceIps: convertTFStringsToGoStrings(lb.Security.SourceIps),
				},
				Listeners: []itacservices.IKSLoadBalancerListener{},
			},
		}
		for _, listener := range lb.Listeners {
			inArg.Spec.Listeners = append(inArg.Spec.Listeners, itacservices.IKSLoadBalancerListener{
				Port:     listener.Port.ValueInt64(),
				Protocol: listener.Protocol.ValueString(),
				Pool: itacservices.IKSLoadBalancerPool{
					Port:              listener.Pool.Port.ValueInt64(),
					Monitor:           listener.Pool.Monitor.ValueString(),
					LoadBalancingMode: listener.Pool.LoadBalancingMode.ValueString(),
					NodeGroupID:       listener.Pool.NodeGroupId.ValueString(),
				},
				Security: itacservices.IKSLoadBalancerSecurity{
					SourceIps: convertTFStringsToGoStrings(listener.Security.SourceIps),
				},
			})
		}
		// Call the update API
		err := r.client.UpdateIKSLoadBalancer(ctx, &inArg, state.ClusterUUID.ValueString(), lbID)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating IKS Load Balancer",
				"Could not update IKS Load Balancer with ID "+lb.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		tflog.Info(ctx, "Successfully updated IKS Load Balancer", map[string]any{"ID": lb.ID.ValueString()})
	}

	// Get refreshed order value from IDC Service irrespective of whether upgrade was done or skipped
	plan.LoadBalancer.ID = types.StringValue(lbID)
	currState, err := r.refreshIKSLoadBalancerResourceModel(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IKS Loadbalancer resource",
			"Could not read IKS Loadbalancer for cluster ID "+plan.ClusterUUID.String()+": "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, currState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, "no change detected change in cluster spec, skipping update")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *iksLBResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state *iksLoadBalancerResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	deleteTimeout, err := state.Timeouts.GetTimeouts(IKSLoadBalancerResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout for loadbalancer: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Check if the load balancer exists before attempting to delete
	tflog.Info(ctx, "Deleting IKS Load Balancer", map[string]any{"ID": state.LoadBalancer.ID.ValueString()})
	// Call the delete API
	err = r.client.DeleteIKSLoadBalancer(ctx, state.ClusterUUID.ValueString(), state.LoadBalancer.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting IKS Load Balancer",
			"Could not delete IKS Load Balancer with ID "+state.LoadBalancer.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	tflog.Info(ctx, "Successfully deleted IKS Load Balancer", map[string]any{"ID": state.LoadBalancer.ID.ValueString()})

}

func (r *iksLBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	clusterUUID := ids[0]
	lbId := ids[1]

	// Basic validation
	if clusterUUID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be the cluster UUID, got empty string.",
		)
		return
	}

	// Fetch LBs for this cluster
	lb, err := r.client.GetIKSLoadBalancerByID(ctx, clusterUUID, lbId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to import IKS Load Balancer",
			fmt.Sprintf("Error retrieving load balancers for cluster %s %s", clusterUUID, err.Error()),
		)
		return
	}

	// Convert API response to state model
	var listeners []models.IKSLoadBalancerListenerModel
	for _, l := range lb.Spec.Listeners {
		var sourceIps []types.String
		for _, ip := range l.Security.SourceIps {
			sourceIps = append(sourceIps, types.StringValue(ip))
		}
		listeners = append(listeners, models.IKSLoadBalancerListenerModel{
			Port:     types.Int64Value(int64(l.Port)),
			Protocol: types.StringValue(string(l.Protocol)),
			Security: models.IKSLoadBalancerSecurityModel{
				SourceIps: sourceIps,
			},
			Pool: models.IKSLoadBalancerPoolModel{
				Port:              types.Int64Value(int64(l.Pool.Port)),
				Monitor:           types.StringValue(string(l.Pool.Monitor)),
				LoadBalancingMode: types.StringValue(string(l.Pool.LoadBalancingMode)),
				NodeGroupId:       types.StringValue(l.Pool.InstanceSelectors["nodegroupName"]),
			},
		})
	}
	var securitySourceIps []types.String
	for _, ip := range lb.Spec.Security.SourceIps {
		securitySourceIps = append(securitySourceIps, types.StringValue(ip))
	}
	lbModel := models.IKSLoadBalancer{
		ID:   types.StringValue(lb.Metadata.ResourceID),
		Name: types.StringValue(lb.Metadata.Name),
		Security: models.IKSLoadBalancerSecurityModel{
			SourceIps: securitySourceIps,
		},
		Schema:    types.StringValue(string(lb.Spec.Schema)),
		Listeners: listeners,
	}

	// Set the full state
	resp.State.Set(ctx, &iksLoadBalancerResourceModel{
		ClusterUUID:  types.StringValue(clusterUUID),
		LoadBalancer: lbModel,
	})

	// Set the import state
	resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), clusterUUID)
}

func (r *iksLBResource) refreshIKSLoadBalancerResourceModel(ctx context.Context, plan *iksLoadBalancerResourceModel) (*iksLoadBalancerResourceModel, error) {
	state := &iksLoadBalancerResourceModel{}
	state.ClusterUUID = plan.ClusterUUID
	lb := plan.LoadBalancer
	loadbalancer, err := r.client.GetIKSLoadBalancerByID(ctx, plan.ClusterUUID.ValueString(), lb.ID.ValueString())
	if err != nil {
		return state, fmt.Errorf("error fetching IKS Load Balancer with ID %s: %w", lb.ID.ValueString(), err)
	}
	var securitySourceIps []types.String
	for _, ip := range loadbalancer.Spec.Security.SourceIps {
		securitySourceIps = append(securitySourceIps, types.StringValue(ip))
	}
	var listeners []models.IKSLoadBalancerListenerModel
	for _, listener := range loadbalancer.Spec.Listeners {
		var sourceIps []types.String
		for _, ip := range listener.Security.SourceIps {
			sourceIps = append(sourceIps, types.StringValue(ip))
		}
		listeners = append(listeners, models.IKSLoadBalancerListenerModel{
			Port:     types.Int64Value(int64(listener.Port)),
			Protocol: types.StringValue(string(listener.Protocol)),
			Security: models.IKSLoadBalancerSecurityModel{
				SourceIps: sourceIps,
			},
			Pool: models.IKSLoadBalancerPoolModel{
				Port:              types.Int64Value(int64(listener.Pool.Port)),
				Monitor:           types.StringValue(string(listener.Pool.Monitor)),
				LoadBalancingMode: types.StringValue(string(listener.Pool.LoadBalancingMode)),
				NodeGroupId:       types.StringValue(listener.Pool.InstanceSelectors["nodegroupName"]),
			},
		})
	}
	state.LoadBalancer = models.IKSLoadBalancer{
		ID:   types.StringValue(loadbalancer.Metadata.ResourceID),
		Name: types.StringValue(loadbalancer.Metadata.Name),
		Security: models.IKSLoadBalancerSecurityModel{
			SourceIps: securitySourceIps,
		},
		Schema:    types.StringValue(string(loadbalancer.Spec.Schema)),
		Listeners: listeners,
	}

	// set timeout again for consistency
	state.Timeouts = plan.Timeouts

	return state, nil
}

func (r *iksLBResource) checkLBExistsAndGetID(ctx context.Context, clusteruuid, lbName string) (bool, string) {
	lb, err := r.client.GetIKSLoadBalancerByClusterUUID(ctx, clusteruuid)
	if err != nil {
		tflog.Error(ctx, "Error checking if IKS Load Balancer exists", map[string]any{"error": err.Error()})
		return false, ""
	}

	if len(lb.Items) > 0 {
		for _, item := range lb.Items {
			if item.Metadata.Name == lbName && item.Metadata.Name != "public-apiserver" {
				return true, item.Metadata.ResourceID
			}
		}
	}

	return false, ""
}
