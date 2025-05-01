package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &iksClusterResource{}
	_ resource.ResourceWithConfigure   = &iksClusterResource{}
	_ resource.ResourceWithImportState = &iksClusterResource{}
)

// orderKubernetesModel maps the resource schema data.
type iksClusterResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Cloudaccount     types.String `tfsdk:"cloudaccount"`
	Name             types.String `tfsdk:"name"`
	K8sversion       types.String `tfsdk:"kubernetes_version"`
	ClusterStatus    types.String `tfsdk:"cluster_status"`
	Network          types.Object `tfsdk:"network"`
	UpgardeAvailable types.Bool   `tfsdk:"upgrade_available"`
	// UpgradableVersions []types.String `tfsdk:"upgrade_k8s_versions_available"`
	Timeouts *timeoutsModel `tfsdk:"timeouts"`

	Storage *models.IKSStorage `tfsdk:"storage"`
}

// NewIKSClusterResource is a helper function to simplify the provider implementation.
func NewIKSClusterResource() resource.Resource {
	return &iksClusterResource{}
}

// orderKubernetes is the resource implementation.
type iksClusterResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *iksClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *iksClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iks_cluster"
}

// Schema defines the schema for the resource.
func (r *iksClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"cloudaccount": schema.StringAttribute{
				Computed: true,
			},
			"kubernetes_version": schema.StringAttribute{
				Required: true,
			},
			"cluster_status": schema.StringAttribute{
				Computed: true,
			},
			"network": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"cluster_cidr": types.StringType,
					"service_cidr": types.StringType,
					"cluster_dns":  types.StringType,
					"enable_lb":    types.BoolType,
				},
				Computed: true,
			},
			"storage": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"size_in_tb": schema.Int64Attribute{
						Required: true,
					},
					"state": schema.StringAttribute{
						Computed: true,
					},
					"storage_provider": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"upgrade_available": schema.BoolAttribute{
				Computed: true,
			},
			// "upgrade_k8s_versions_available": schema.ListAttribute{
			// 	ElementType: types.StringType,
			// 	Computed:    true,
			// },
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"resource_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timeout for cluster resource operations",
						Default:     stringdefault.StaticString(IKSClusterResourceTimeout),
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *iksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan iksClusterResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	createTimeout, err := plan.Timeouts.GetTimeouts(IKSClusterResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse create timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	inArg := itacservices.IKSCreateRequest{
		Name:         plan.Name.ValueString(),
		K8sVersion:   plan.K8sversion.ValueString(),
		InstanceType: "iks-cluster",
		RuntimeName:  "Containerd",
	}
	iksClusterResp, cloudaccount, err := r.client.CreateIKSCluster(ctx, &inArg, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(iksClusterResp.ResourceId)
	plan.ClusterStatus = types.StringValue(iksClusterResp.ClusterState)
	if cloudaccount != nil {
		plan.Cloudaccount = types.StringValue(*cloudaccount)
	} else {
		plan.Cloudaccount = types.StringNull()
	}

	network := models.ClusterNetwork{
		ClusterCIDR: types.StringValue(iksClusterResp.Network.ClusterCIDR),
		ClusterDNS:  types.StringValue(iksClusterResp.Network.ClusterDNS),
		EnableLB:    types.BoolValue(iksClusterResp.Network.EnableLB),
		ServiceCIDR: types.StringValue(iksClusterResp.Network.ServcieCIDR),
	}
	plan.Network, diags = types.ObjectValueFrom(ctx, network.AttributeTypes(), network)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Storage != nil && !plan.Storage.Size.IsNull() {
		inArg := itacservices.IKSStorageCreateRequest{
			Enable: true,
			Size:   fmt.Sprintf("%sTB", strconv.FormatInt(plan.Storage.Size.ValueInt64(), 10)),
		}

		storageResp, _, err := r.client.CreateIKSStorage(ctx, &inArg, plan.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating iks file storage",
				"Could not create iks file storage, unexpected error: "+err.Error(),
			)
			return
		}

		sizeStr := strings.Split(storageResp.Size, "TB")[0]
		sizeNum, _ := strconv.ParseInt(sizeStr, 10, 64)
		currV := models.IKSStorage{
			Size:            types.Int64Value(sizeNum),
			State:           types.StringValue(storageResp.State),
			StorageProvider: types.StringValue(storageResp.Provider),
		}
		plan.Storage = &currV
	}

	plan.UpgardeAvailable = types.BoolValue(iksClusterResp.UpgradeAvailable)

	// for _, k := range iksClusterResp.UpgradableK8sVersions {
	// 	plan.KubernetesCluster.UpgradableVersions = append(plan.KubernetesCluster.UpgradableVersions, types.StringValue(k))
	// }

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *iksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	// Get current state
	var state iksClusterResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	readTimeout, err := state.Timeouts.GetTimeouts(IKSClusterResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	iksClusterResp, cloudaccount, err := r.client.GetIKSClusterByClusterUUID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading state",
			"Could not read state, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	state.ID = types.StringValue(iksClusterResp.ResourceId)
	state.ClusterStatus = types.StringValue(iksClusterResp.ClusterState)
	state.K8sversion = types.StringValue(iksClusterResp.K8sVersion)
	if cloudaccount != nil {
		state.Cloudaccount = types.StringValue(*cloudaccount)
	} else {
		state.Cloudaccount = types.StringNull()
	}

	network := models.ClusterNetwork{
		ClusterCIDR: types.StringValue(iksClusterResp.Network.ClusterCIDR),
		ClusterDNS:  types.StringValue(iksClusterResp.Network.ClusterDNS),
		EnableLB:    types.BoolValue(iksClusterResp.Network.EnableLB),
		ServiceCIDR: types.StringValue(iksClusterResp.Network.ServcieCIDR),
	}
	state.Network, diags = types.ObjectValueFrom(ctx, network.AttributeTypes(), network)

	for _, v := range iksClusterResp.Storages {
		sizeStr := strings.Split(v.Size, "TB")[0]
		size, _ := strconv.ParseInt(sizeStr, 10, 64)

		currV := models.IKSStorage{
			Size:            types.Int64Value(size),
			State:           types.StringValue(v.State),
			StorageProvider: types.StringValue(v.Provider),
		}
		state.Storage = &currV
	}

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// volumes := []models.IKSStorage{}
	// for _, v := range iksClusterResp.Storages {
	// 	currV := models.IKSStorage{
	// 		Size:            types.StringValue(v.Size),
	// 		State:           types.StringValue(v.State),
	// 		StorageProvider: types.StringValue(v.Provider),
	// 	}
	// 	volumes = append(volumes, currV)
	// }
	// state.KubernetesCluster.Storage, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.IKStorageAttributes), volumes)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }

	// ngs := []models.NodeGroup{}
	// for _, n := range iksClusterResp.NodeGroups {
	// 	ng := models.NodeGroup{
	// 		ID:           types.StringValue(n.ID),
	// 		Name:         types.StringValue(n.Name),
	// 		Count:        types.Int64Value(n.Count),
	// 		InstanceType: types.StringValue(n.InstanceType),
	// 		IMIId:        types.StringValue(n.IMIID),
	// 		State:        types.StringValue(n.State),
	// 		UserDataURL:  types.StringValue(n.UserDataURL),
	// 	}
	// 	for _, sshk := range n.SSHKeyNames {
	// 		state.KubernetesCluster.SSHPublicKeyNames = append(state.KubernetesCluster.SSHPublicKeyNames, types.StringValue(sshk.Name))
	// 	}
	// 	ngs = append(ngs, ng)
	// }
	// state.KubernetesCluster.NodeGroups, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.NodeGroupAttributes), ngs)
	// resp.Diagnostics.Append(diags...)
	// if resp.Diagnostics.HasError() {
	// 	return
	// }
	state.UpgardeAvailable = types.BoolValue(iksClusterResp.UpgradeAvailable)

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *iksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state iksClusterResourceModel

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
	updateTimeout, err := state.Timeouts.GetTimeouts(IKSClusterResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse update timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	if !plan.K8sversion.Equal(state.K8sversion) {
		tflog.Info(ctx, "Detected change in iks cluster spec for k8s version, updating cluster",
			map[string]any{"current version ": state.K8sversion.ValueString(), "new version": plan.K8sversion.ValueString()})

		inArg := itacservices.UpgradeClusterRequest{
			ClusterId:  state.ID.ValueString(),
			K8sVersion: plan.K8sversion.ValueString(),
		}
		err := r.client.UpgradeCluster(ctx, &inArg)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating order",
				"Could not create order, unexpected error: "+err.Error(),
			)
			return
		}
	}

	// Get refreshed order value from IDC Service irrespective of whether upgrade was done or skipped
	cluster, cloudaccount, err := r.client.GetIKSClusterByClusterUUID(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IKS Cluster resource",
			"Could not read IKS Cluster resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	currState, err := refreshIKSCLusterResourceModel(ctx, cluster, cloudaccount)
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
	tflog.Info(ctx, "no change detected change in cluster spec, skipping update")
}

func (r *iksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *iksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state iksClusterResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	deleteTimeout, err := state.Timeouts.GetTimeouts(IKSClusterResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse delete timeout: "+err.Error())
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the order from IDC Services
	err = r.client.DeleteIKSCluster(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC IKS Cluster resource",
			"Could not delete IDC IDC IKS Cluster ID "+state.ID.String()+": "+err.Error(),
		)
		return
	}
}

func refreshIKSCLusterResourceModel(ctx context.Context, cluster *itacservices.IKSCluster, cloudaccount *string) (*iksClusterResourceModel, error) {
	state := &iksClusterResourceModel{}
	var diags diag.Diagnostics

	state.ID = types.StringValue(cluster.ResourceId)
	state.K8sversion = types.StringValue(cluster.K8sVersion)
	state.Name = types.StringValue(cluster.Name)
	state.ClusterStatus = types.StringValue(cluster.ClusterState)
	if cloudaccount != nil {
		state.Cloudaccount = types.StringValue(*cloudaccount)
	} else {
		state.Cloudaccount = types.StringNull()
	}

	network := models.ClusterNetwork{
		ClusterCIDR: types.StringValue(cluster.Network.ClusterCIDR),
		ClusterDNS:  types.StringValue(cluster.Network.ClusterDNS),
		EnableLB:    types.BoolValue(cluster.Network.EnableLB),
		ServiceCIDR: types.StringValue(cluster.Network.ServcieCIDR),
	}
	state.Network, diags = types.ObjectValueFrom(ctx, network.AttributeTypes(), network)
	if diags.HasError() {
		return state, fmt.Errorf("error parsing values")
	}

	return state, nil
}
