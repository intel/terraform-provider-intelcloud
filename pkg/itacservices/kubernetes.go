package itacservices

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"terraform-provider-intelcloud/pkg/itacservices/common"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	retry "github.com/sethvargo/go-retry"
)

var (
	getAllK8sClustersURL       = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters"
	createK8sClusterURL        = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters"
	getIksClusterByClusterUUID = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}"
	deleteIksCluster           = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}"

	createK8sNodeGroupURL = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/nodegroups"
	getK8sNodeGroupURL    = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/nodegroups/{{.NodeGroupUUID}}"
	updateNodeGroupURL    = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/nodegroups/{{.NodeGroupUUID}}"

	getK8sKubeconfigURL  = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/kubeconfig"
	upgradeK8sClusterURL = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/upgrade"

	createK8sFileStorageURL = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/storage"
	createIKSLBURL          = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/vips"
	getIKSLBURLByCluster    = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/vips"
	getIKSLBURLByID         = "{{.Host}}/v1/cloudaccounts/{{.Cloudaccount}}/iks/clusters/{{.ClusterUUID}}/vips/{{.VipID}}"
)

type IKSClusters struct {
	Clusters []IKSCluster `json:"clusters"`
}

type IKSCluster struct {
	ResourceId            string         `json:"uuid"`
	Name                  string         `json:"name"`
	Description           string         `json:"description"`
	CreatedAt             string         `json:"createddate"`
	ClusterState          string         `json:"clusterstate"`
	K8sVersion            string         `json:"k8sversion"`
	UpgradeAvailable      bool           `json:"upgradeavailable"`
	UpgradableK8sVersions []string       `json:"upgradek8sversionavailable"`
	Network               ClusterNetwork `json:"network"`
	NodeGroups            []NodeGroup    `json:"nodegroups"`
	StorageEnabled        bool           `json:"storageenabled"`
	Storages              []K8sStorage   `json:"storages"`
	VIPs                  []IKSVIP       `json:"vips"`
}

type IKSVIP struct {
	Id       int64  `json:"vipid"`
	Name     string `json:"name"`
	State    string `json:"vipstate"`
	IP       string `json:"vipIp"`
	Port     int64  `json:"port"`
	PoolPort int64  `json:"poolport"`
	Type     string `json:"viptype"`
}

type ClusterNetwork struct {
	EnableLB    bool   `json:"enableloadbalancer"`
	ServcieCIDR string `json:"servicecidr"`
	ClusterCIDR string `json:"clustercidr"`
	ClusterDNS  string `json:"clusterdns"`
}

type NodeGroup struct {
	ClusterID            string `json:"clusteruuid"`
	ID                   string `json:"nodegroupuuid"`
	Name                 string `json:"name"`
	Count                int64  `json:"count"`
	InstanceType         string `json:"instancetypeid"`
	State                string `json:"nodegroupstate"`
	SSHKeyNames          []SKey `json:"sshkeyname"`
	NetworkInterfaceName string `json:"networkinterfacename"`
	IMIID                string `json:"imiid"`
	UserDataURL          string `json:"userdataurl"`
	Interfaces           []Vnet `json:"vnets"`
}

type SKey struct {
	Name string `json:"sshkey"`
}

type K8sStorage struct {
	Provider string `json:"storageprovider"`
	Size     string `json:"size"`
	State    string `json:"state"`
}

type IKSNodeGroupCreateRequest struct {
	Count          int64  `json:"count"`
	Name           string `json:"name"`
	ProductType    string `json:"instanceType"`
	InstanceTypeId string `json:"instancetypeid"`
	SSHKeyNames    []SKey `json:"sshkeyname"`
	UserDataURL    string `json:"userdataurl"`
	Vnets          []Vnet `json:"vnets"`
}

type Vnet struct {
	AvailabilityZoneName     string `json:"availabilityzonename"`
	NetworkInterfaceVnetName string `json:"networkinterfacevnetname"`
}

type IKSCreateRequest struct {
	Name         string `json:"name"`
	Count        int64  `json:"count"`
	K8sVersion   string `json:"k8sversionname"`
	InstanceType string `json:"instanceType"`
	RuntimeName  string `json:"runtimename"`
}

type IKSStorageCreateRequest struct {
	Enable bool   `json:"enablestorage"`
	Size   string `json:"storagesize"`
}

type IKSLoadBalancerRequest struct {
	Name    string `json:"name"`
	Port    int    `json:"port"`
	VIPType string `json:"viptype"`
}

type IKSLoadBalancer struct {
	ID       int64  `json:"vipid"`
	Name     string `json:"name"`
	Port     int    `json:"port"`
	VIPType  string `json:"viptype"`
	VIPState string `json:"vipstate"`
	VIPIP    string `json:"vipip"`
	PoolPort int    `json:"poolport"`
}

type IKSLBsByCluster struct {
	Items []IKSLoadBalancer `json:"response"`
}

type KubeconfigResponse struct {
	Config string `json:"kubeconfig"`
}

type UpgradeClusterRequest struct {
	ClusterId  string `json:"clusteruuid"`
	K8sVersion string `json:"k8sversionname"`
}

type UpgradeClusterPayload struct {
	K8sVersion string `json:"k8sversionname"`
}

type UpdateNodeGroupRequest struct {
	ClusterId   string `json:"clusteruuid"`
	NodeGroupId string `json:"nodegroupuuid"`
	Count       int64  `json:"count"`
}

type UpdateNodeGroupPayload struct {
	Count int64 `json:"count"`
}

func (client *IDCServicesClient) GetKubernetesClusters(ctx context.Context) (*IKSClusters, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getAllK8sClustersURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	tflog.Debug(ctx, "iks read api", map[string]any{"retcode": retcode, "retval": string(retval)})
	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks clusters")
	}

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	clusters := IKSClusters{}
	if err := json.Unmarshal(retval, &clusters); err != nil {
		tflog.Debug(ctx, "iks read api", map[string]any{"err": err})
		return nil, nil, fmt.Errorf("error parsing iks cluster response")
	}

	return &clusters, client.Cloudaccount, nil
}

func (client *IDCServicesClient) CreateIKSCluster(ctx context.Context, in *IKSCreateRequest, async bool) (*IKSCluster, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createK8sClusterURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "iks create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks create response")
	}
	tflog.Debug(ctx, "iks create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	cluster := &IKSCluster{}
	if err := json.Unmarshal(retval, cluster); err != nil {
		return nil, nil, fmt.Errorf("error parsing instance response")
	}

	if async {
		cluster, _, err = client.GetIKSClusterByClusterUUID(ctx, cluster.ResourceId)
		if err != nil {
			return cluster, nil, fmt.Errorf("error reading iks cluster state")
		}
	} else {
		backoffTimer := retry.NewConstant(5 * time.Second)
		backoffTimer = retry.WithMaxDuration(3000*time.Second, backoffTimer)

		if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
			cluster, _, err = client.GetIKSClusterByClusterUUID(ctx, cluster.ResourceId)
			if err != nil {
				return fmt.Errorf("error reading instance state")
			}
			if cluster.ClusterState == "Active" {
				return nil
			} else if cluster.ClusterState == "Failed" {
				return fmt.Errorf("instance state failed")
			} else {
				return retry.RetryableError(fmt.Errorf("iks cluster state not ready, retry again"))
			}
		}); err != nil {
			return nil, nil, fmt.Errorf("iks cluster state not ready after maximum retries")
		}
	}

	return cluster, client.Cloudaccount, nil
}

func (client *IDCServicesClient) GetIKSClusterByClusterUUID(ctx context.Context, clusterUUID string) (*IKSCluster, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getIksClusterByClusterUUID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading sshkey by resource id")
	}
	tflog.Debug(ctx, "iks get cluster by UUID api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	cluster := IKSCluster{}
	if err := json.Unmarshal(retval, &cluster); err != nil {
		return nil, nil, fmt.Errorf("error parsing iks cluster get response")
	}
	return &cluster, client.Cloudaccount, nil
}

func (client *IDCServicesClient) DeleteIKSCluster(ctx context.Context, clusterUUID string) error {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(deleteIksCluster, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	tflog.Debug(ctx, "iks cluster delete api", map[string]any{"parsedurl": parsedURL})
	retcode, retval, err := common.MakeDeleteAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return fmt.Errorf("error deleting sshkey by resource id")
	}

	if retcode != http.StatusOK {
		return common.MapHttpError(retcode, retval)
	}

	tflog.Debug(ctx, "iks cluster delete api", map[string]any{"retcode": retcode})

	return nil
}

func (client *IDCServicesClient) CreateIKSNodeGroup(ctx context.Context, in *IKSNodeGroupCreateRequest, clusterUUID string, async bool) (*NodeGroup, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createK8sNodeGroupURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "iks node group create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks node group create response")
	}
	tflog.Debug(ctx, "iks node group create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	ng := &NodeGroup{}
	if err := json.Unmarshal(retval, ng); err != nil {
		return nil, nil, fmt.Errorf("error parsing node group response")
	}

	backoffTimer := retry.NewConstant(common.DefaultRetryInterval)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		ng, err = client.GetIKSNodeGroupByID(ctx, clusterUUID, ng.ID)
		if err != nil {
			return fmt.Errorf("error reading node group state")
		}
		tflog.Debug(ctx, "iks node group create api response", map[string]any{"nodegroupuuid": ng.ID, "state": ng.State})
		if ng.State == "Active" {
			return nil
		} else if ng.State == "Failed" {
			return fmt.Errorf("node group state failed")
		}
		return retry.RetryableError(fmt.Errorf("iks node group state not ready, retry again"))
	}); err != nil {
		return nil, nil, fmt.Errorf("iks node group state not ready after maximum retries")
	}
	return ng, client.Cloudaccount, nil
}

func (client *IDCServicesClient) GetIKSNodeGroupByID(ctx context.Context, clusterId, ngId string) (*NodeGroup, error) {
	params := struct {
		Host          string
		Cloudaccount  string
		ClusterUUID   string
		NodeGroupUUID string
	}{
		Host:          *client.Host,
		Cloudaccount:  *client.Cloudaccount,
		ClusterUUID:   clusterId,
		NodeGroupUUID: ngId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getK8sNodeGroupURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading node group resource by id")
	}
	tflog.Debug(ctx, "iks node group read response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	nodeGroup := NodeGroup{}
	if err := json.Unmarshal(retval, &nodeGroup); err != nil {
		return nil, fmt.Errorf("error parsing iks cluster get response")
	}
	return &nodeGroup, nil
}

func (client *IDCServicesClient) CreateIKSStorage(ctx context.Context, in *IKSStorageCreateRequest, clusterUUID string) (*K8sStorage, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createK8sFileStorageURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "iks file storage create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks file storage create response")
	}
	tflog.Debug(ctx, "iks file storage create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	storage := &K8sStorage{}
	if err := json.Unmarshal(retval, storage); err != nil {
		return nil, nil, fmt.Errorf("error parsing node group response")
	}

	backoffTimer := retry.NewConstant(common.DefaultRetryInterval)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		iksCluster, _, err := client.GetIKSClusterByClusterUUID(ctx, clusterUUID)
		if err != nil {
			return fmt.Errorf("error reading node group state")
		}
		for _, v := range iksCluster.Storages {
			if strings.EqualFold(v.Size, storage.Size) {
				if v.State == "Active" {
					storage.Provider = v.Provider
					storage.State = v.State
					storage.Size = v.Size
					return nil
				} else if v.State == "Failed" {
					return fmt.Errorf("file storage state failed")
				}
			} else {
				return retry.RetryableError(fmt.Errorf("iks file storage state not ready, retry again"))
			}
		}
		return retry.RetryableError(fmt.Errorf("iks file storage state not ready, retry again"))
	}); err != nil {
		return nil, nil, fmt.Errorf("iks node group state not ready after maximum retries")
	}

	return storage, client.Cloudaccount, nil
}

func (client *IDCServicesClient) CreateIKSLoadBalancer(ctx context.Context, in *IKSLoadBalancerRequest, clusterUUID string) (*IKSLoadBalancer, *string, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(createIKSLBURL, params)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(in, "", "    ")
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing input arguments")
	}

	tflog.Debug(ctx, "iks load balancer create api request", map[string]any{"url": parsedURL, "inArgs": string(inArgs)})
	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)

	if err != nil {
		return nil, nil, fmt.Errorf("error reading iks load balancer create response")
	}
	tflog.Debug(ctx, "iks load balancer create api response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, nil, common.MapHttpError(retcode, retval)
	}

	iksLB := &IKSLoadBalancer{}
	if err := json.Unmarshal(retval, iksLB); err != nil {
		return nil, nil, fmt.Errorf("error parsing load balancer response")
	}

	backoffTimer := retry.NewConstant(5 * time.Second)
	backoffTimer = retry.WithMaxDuration(3000*time.Second, backoffTimer)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		iksLB, err = client.GetIKSLoadBalancerByID(ctx, clusterUUID, iksLB.ID)
		if err != nil {
			return fmt.Errorf("error reading node group state")
		}
		if iksLB.VIPState == "Active" {
			return nil
		} else {
			return retry.RetryableError(fmt.Errorf("iks load balancer state not ready, retry again"))
		}
	}); err != nil {
		return nil, nil, fmt.Errorf("iks node group state not ready after maximum retries")
	}

	return iksLB, client.Cloudaccount, nil
}

func (client *IDCServicesClient) GetIKSLoadBalancerByID(ctx context.Context, clusterUUID string, vipId int64) (*IKSLoadBalancer, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
		VipID        int64
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
		VipID:        vipId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getIKSLBURLByID, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading load balancer resource by id")
	}
	tflog.Debug(ctx, "iks load balancer read response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	iksLB := IKSLoadBalancer{}
	if err := json.Unmarshal(retval, &iksLB); err != nil {
		return nil, fmt.Errorf("error parsing iks load balancer get response")
	}
	return &iksLB, nil
}

func (client *IDCServicesClient) GetIKSLoadBalancerByClusterUUID(ctx context.Context, clusterUUID string) (*IKSLBsByCluster, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterUUID,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getIKSLBURLByCluster, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error reading load balancer resource by cluster")
	}
	tflog.Debug(ctx, "iks load balancer read response", map[string]any{"retcode": retcode, "retval": string(retval)})

	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	resp := IKSLBsByCluster{}
	if err := json.Unmarshal(retval, &resp); err != nil {
		return nil, fmt.Errorf("error parsing iks load balancer get response")
	}
	return &resp, nil
}

func (client *IDCServicesClient) DeleteIKSNodeGroup(ctx context.Context, clusterId, ngId string) error {
	params := struct {
		Host          string
		Cloudaccount  string
		ClusterUUID   string
		NodeGroupUUID string
	}{
		Host:          *client.Host,
		Cloudaccount:  *client.Cloudaccount,
		ClusterUUID:   clusterId,
		NodeGroupUUID: ngId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getK8sNodeGroupURL, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	tflog.Debug(ctx, "iks node group delete api", map[string]any{"parsedurl": parsedURL})
	retcode, retval, err := common.MakeDeleteAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return fmt.Errorf("error deleting iks node group by resource id")
	}
	tflog.Debug(ctx, "iks node group delete api", map[string]any{"retcode": retcode})
	if retcode != http.StatusOK {
		return common.MapHttpError(retcode, retval)
	}

	return nil
}

func (client *IDCServicesClient) GetClusterKubeconfig(ctx context.Context, clusterId string) (*string, error) {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  clusterId,
	}

	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(getK8sKubeconfigURL, params)
	if err != nil {
		return nil, fmt.Errorf("error parsing the url")
	}

	retcode, retval, err := common.MakeGetAPICall(ctx, parsedURL, *client.Apitoken, nil)
	if err != nil {
		return nil, fmt.Errorf("error calling get kubeconfig api")
	}
	tflog.Debug(ctx, "iks get kubeconfig", map[string]any{"retcode": retcode})
	if retcode != http.StatusOK {
		return nil, common.MapHttpError(retcode, retval)
	}

	resp := KubeconfigResponse{}
	if err := json.Unmarshal(retval, &resp); err != nil {
		return nil, fmt.Errorf("error parsing iks kubeconfig get response")
	}

	return &resp.Config, nil
}

func (client *IDCServicesClient) UpgradeCluster(ctx context.Context, in *UpgradeClusterRequest) error {
	params := struct {
		Host         string
		Cloudaccount string
		ClusterUUID  string
	}{
		Host:         *client.Host,
		Cloudaccount: *client.Cloudaccount,
		ClusterUUID:  in.ClusterId,
	}

	inArg := UpgradeClusterPayload{
		K8sVersion: in.K8sVersion,
	}
	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(upgradeK8sClusterURL, params)
	if err != nil {
		return fmt.Errorf("error parsing the url")
	}

	inArgs, err := json.MarshalIndent(inArg, "", "    ")
	if err != nil {
		return fmt.Errorf("error parsing input arguments")
	}

	retcode, retval, err := common.MakePOSTAPICall(ctx, parsedURL, *client.Apitoken, inArgs)
	if err != nil {
		return fmt.Errorf("error calling upgrade cluster api")
	}
	if retcode != http.StatusOK {
		return common.MapHttpError(retcode, retval)
	}
	tflog.Debug(ctx, "iks upgrade cluster", map[string]any{"retcode": retcode, "retval": retval})

	cluster := &IKSCluster{}
	if err := json.Unmarshal(retval, cluster); err != nil {
		return fmt.Errorf("error parsing instance response")
	}

	backoffTimer := retry.NewConstant(5 * time.Second)
	backoffTimer = retry.WithMaxDuration(1800*time.Second, backoffTimer)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		cluster, _, err = client.GetIKSClusterByClusterUUID(ctx, in.ClusterId)
		if err != nil {
			return fmt.Errorf("error reading instance state after upgrade")
		}
		if cluster.ClusterState == "Active" {
			return nil
		} else if cluster.ClusterState == "Failed" {
			return fmt.Errorf("instance state failed")
		} else {
			return retry.RetryableError(fmt.Errorf("iks cluster state not ready, retry again"))
		}
	}); err != nil {
		return fmt.Errorf("iks cluster state not ready after maximum retries")
	}

	return nil
}

func (client *IDCServicesClient) UpdateNodeGroup(ctx context.Context, in *UpdateNodeGroupRequest) error {
	params := struct {
		Host          string
		Cloudaccount  string
		ClusterUUID   string
		NodeGroupUUID string
	}{
		Host:          *client.Host,
		Cloudaccount:  *client.Cloudaccount,
		ClusterUUID:   in.ClusterId,
		NodeGroupUUID: in.NodeGroupId,
	}

	inArg := UpdateNodeGroupPayload{
		Count: in.Count,
	}
	// Parse the template string with the provided data
	parsedURL, err := common.ParseString(updateNodeGroupURL, params)
	if err != nil {
		return fmt.Errorf("error parsing the url: %v", err)
	}

	inArgs, err := json.MarshalIndent(inArg, "", "    ")
	if err != nil {
		return fmt.Errorf("error parsing input arguments")
	}

	retcode, retval, err := common.MakePutAPICall(ctx, parsedURL, *client.Apitoken, inArgs)
	if err != nil {
		return fmt.Errorf("error calling upgrade cluster api")
	}
	if retcode != http.StatusOK {
		return common.MapHttpError(retcode, retval)
	}
	tflog.Debug(ctx, "iks update nodegroup", map[string]any{"retcode": retcode, "retval": retval})

	nodeGroup := &NodeGroup{}
	if err := json.Unmarshal(retval, nodeGroup); err != nil {
		return fmt.Errorf("error parsing instance response")
	}

	backoffTimer := retry.NewConstant(5 * time.Second)

	if err := retry.Do(ctx, backoffTimer, func(_ context.Context) error {
		nodeGroup, err = client.GetIKSNodeGroupByID(ctx, in.ClusterId, in.NodeGroupId)
		if err != nil {
			return fmt.Errorf("error reading nodegroup state after update")
		}
		if nodeGroup.State == "Active" {
			return nil
		} else if nodeGroup.State == "Failed" {
			return fmt.Errorf("nodegroup state failed")
		} else {
			return retry.RetryableError(fmt.Errorf("iks nodegroup state not ready, retry again"))
		}
	}); err != nil {
		return fmt.Errorf("iks nodegroup state not ready after maximum retries")
	}

	return nil
}
