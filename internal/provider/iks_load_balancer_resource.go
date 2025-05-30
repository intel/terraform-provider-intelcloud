package provider

import (
	"context"
	"fmt"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &iksLBResource{}
	_ resource.ResourceWithConfigure   = &iksLBResource{}
	_ resource.ResourceWithImportState = &iksLBResource{}
)

// orderIKSNodeGroupModel maps the resource schema data.
type iksLBResourceModel struct {
	ClusterUUID   types.String             `tfsdk:"cluster_uuid"`
	LoadBalancers []models.IKSLoadBalancer `tfsdk:"load_balancers"`
	Timeouts      *timeoutsModel           `tfsdk:"timeouts"`
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
			"load_balancers": schema.ListNestedBlock{
				Description: "List of load balancers to be provisioned.",
				NestedObject: schema.NestedBlockObject{
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
	var plan iksLBResourceModel

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

	for idx := range plan.LoadBalancers {
		inArg := itacservices.IKSLoadbalancerCreateRequest{
			Metadata: itacservices.IKSLoadBalancerCreateMetadata{
				Name:        plan.LoadBalancers[idx].Name.ValueString(),
				ClusterUUID: plan.ClusterUUID.ValueString(),
			},
		}
		for _, listener := range plan.LoadBalancers[idx].Listeners {
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
		inArg.Spec.Security.SourceIps = convertTFStringsToGoStrings(plan.LoadBalancers[idx].Security.SourceIps)

		ilbResp, _, err := r.client.CreateIKSLoadBalancer(ctx, &inArg, plan.ClusterUUID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating iks load balancer",
				"Could not create iks load balancer, unexpected error: "+err.Error(),
			)
			return
		}

		plan.LoadBalancers[idx].ID = types.StringValue(ilbResp.Metadata.ResourceID)
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
		//lbId, _ := strconv.ParseInt(lb.ID.ValueString(), 10, 64)
		refreshedState, err := r.client.GetIKSLoadBalancerByID(ctx, state.ClusterUUID.ValueString(), lb.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading IDC Compute IKS Load Balancer resource",
				"Could not read IDC Compute IKS Load Balancer resource ID "+state.ClusterUUID.ValueString()+": "+err.Error(),
			)
			return
		}
		state.LoadBalancers[idx].ID = types.StringValue(refreshedState.Metadata.ResourceID)

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
	// To be implemented, currently API access is disabled
}

func (r *iksLBResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	// Expect import ID in the format: cluster_id:id
	clusterUUID := req.ID

	// Basic validation
	if clusterUUID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Expected import ID to be the cluster UUID, got empty string.",
		)
		return
	}

	// Fetch LBs for this cluster
	lbs, err := r.client.GetIKSLoadBalancerByClusterUUID(ctx, clusterUUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to import IKS Load Balancer",
			fmt.Sprintf("Error retrieving load balancers for cluster %s: %s", clusterUUID, err.Error()),
		)
		return
	}

	// Convert API response to state model
	var lbModels []models.IKSLoadBalancer
	for _, lb := range lbs.Items {
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
					NodeGroupId:       types.StringValue(l.Pool.NodeGroupID),
				},
			})
		}
		var securitySourceIps []types.String
		for _, ip := range lb.Spec.Security.SourceIps {
			securitySourceIps = append(securitySourceIps, types.StringValue(ip))
		}
		lbModels = append(lbModels, models.IKSLoadBalancer{
			ID:   types.StringValue(lb.Metadata.ResourceID),
			Name: types.StringValue(lb.Metadata.Name),
			Security: models.IKSLoadBalancerSecurityModel{
				SourceIps: securitySourceIps,
			},
			Schema:    types.StringValue(string(lb.Spec.Schema)),
			Listeners: listeners,
		})
	}

	// Set the full state
	resp.State.Set(ctx, &iksLBResourceModel{
		ClusterUUID:   types.StringValue(clusterUUID),
		LoadBalancers: lbModels,
	})

	// Set the import state
	resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), clusterUUID)
}
