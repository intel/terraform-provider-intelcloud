package provider

import (
	"context"
	"fmt"
	"strconv"
	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewKubernetesDataSource() datasource.DataSource {
	return &kubernetesDataSource{}
}

type kubernetesDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &kubernetesDataSource{}
	_ datasource.DataSourceWithConfigure = &kubernetesDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type kubernetesDataSourceModel struct {
	Clusters []models.KubernetesClusterModel `tfsdk:"clusters"`
}

// Configure adds the provider configured client to the data source.
func (d *kubernetesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *kubernetesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iks_clusters"
}

func (d *kubernetesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"clusters": schema.ListNestedAttribute{
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
						"availability_zone": schema.StringAttribute{
							Optional: true,
						},
						"kubernetes_version": schema.StringAttribute{
							Computed: true,
						},
						"cluster_status": schema.StringAttribute{
							Computed: true,
						},
						"load_balancers": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed: true,
									},
									"vip_state": schema.StringAttribute{
										Computed: true,
									},
									"vip_ip": schema.StringAttribute{
										Computed: true,
									},
									"port": schema.Int64Attribute{
										Computed: true,
									},
									"pool_port": schema.Int64Attribute{
										Computed: true,
									},
									"vip_type": schema.StringAttribute{
										Computed: true,
									},
								},
							},
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
						"node_groups": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed: true,
									},
									"count": schema.Int64Attribute{
										Computed: true,
									},
									"name": schema.StringAttribute{
										Computed: true,
									},
									"instance_type": schema.StringAttribute{
										Computed: true,
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
								},
							},
						},
						"ssh_public_key_names": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"storages": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"size": schema.StringAttribute{
										Computed: true,
									},
									"state": schema.StringAttribute{
										Computed: true,
									},
									"storage_provider": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
						"upgrade_available": schema.BoolAttribute{
							Computed: true,
						},
						"upgrade_k8s_versions_available": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}
func (d *kubernetesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var state kubernetesDataSourceModel
	state.Clusters = []models.KubernetesClusterModel{}

	var diags diag.Diagnostics
	iksClusters, cloudaccount, err := d.client.GetKubernetesClusters(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read IDC Kubernetes Clusters",
			err.Error(),
		)
		return
	}

	for _, cl := range iksClusters.Clusters {
		iksModel := models.KubernetesClusterModel{
			ClusterUUID:      types.StringValue(cl.ResourceId),
			Name:             types.StringValue(cl.Name),
			AvailabilityZone: types.StringNull(),
			Cloudaccount:     types.StringValue(*cloudaccount),
			K8sversion:       types.StringValue(cl.K8sVersion),
			ClusterStatus:    types.StringValue(cl.ClusterState),
			UpgardeAvailable: types.BoolValue(cl.UpgradeAvailable),
		}
		network := models.ClusterNetwork{
			ClusterCIDR: types.StringValue(cl.Network.ClusterCIDR),
			ServiceCIDR: types.StringValue(cl.Network.ServcieCIDR),
			ClusterDNS:  types.StringValue(cl.Network.ClusterDNS),
			EnableLB:    types.BoolValue(cl.Network.EnableLB),
		}
		iksModel.Network, diags = types.ObjectValueFrom(ctx, network.AttributeTypes(), network)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// for _, k := range cl.UpgradableK8sVersions {
		// 	iksModel.UpgradableVersions = append(iksModel.UpgradableVersions, types.StringValue(k))
		// }

		// Map NodeGroups
		ngs := []models.NodeGroup{}
		for _, n := range cl.NodeGroups {
			ng := models.NodeGroup{
				ID:           types.StringValue(n.ID),
				Name:         types.StringValue(n.Name),
				Count:        types.Int64Value(n.Count),
				InstanceType: types.StringValue(n.InstanceType),
				IMIId:        types.StringValue(n.IMIID),
				State:        types.StringValue(n.State),
				UserDataURL:  types.StringValue(n.UserDataURL),
			}
			for _, sshk := range n.SSHKeyNames {
				iksModel.SSHPublicKeyNames = append(iksModel.SSHPublicKeyNames, types.StringValue(sshk.Name))
			}
			ngs = append(ngs, ng)
		}

		ngObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.NodeGroupAttributes), ngs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		iksModel.NodeGroups = ngObj

		// Map LoadBalancer/VIPs
		vips := []models.IKSLoadBalancer{}
		for _, v := range cl.VIPs {
			vip := models.IKSLoadBalancer{
				Name:     types.StringValue(v.Name),
				VipState: types.StringValue(v.State),
				VipIp:    types.StringValue(v.IP),
				Port:     types.Int64Value(v.Port),
				PoolPort: types.Int64Value(v.PoolPort),
				VipType:  types.StringValue(v.Type),
			}
			vips = append(vips, vip)
		}
		lbObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.IKSLoadLalancerAttributes), vips)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		iksModel.LoadBalancer = lbObj

		// Map Storages
		vols := []models.IKSStorage{}
		for _, s := range cl.Storages {
			sizeNum, _ := strconv.ParseInt(s.Size, 10, 64)
			vol := models.IKSStorage{
				Size:            types.Int64Value(sizeNum),
				StorageProvider: types.StringValue(s.Provider),
				State:           types.StringValue(s.State),
			}
			vols = append(vols, vol)
		}
		storageObj, diags := types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.IKStorageAttributes), vols)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		iksModel.Storage = storageObj

		state.Clusters = append(state.Clusters, iksModel)
	}
	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
