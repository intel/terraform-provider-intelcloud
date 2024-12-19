package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type KubernetesClusterModel struct {
	ClusterUUID        types.String   `tfsdk:"uuid"`
	Cloudaccount       types.String   `tfsdk:"cloudaccount"`
	Name               types.String   `tfsdk:"name"`
	AvailabilityZone   types.String   `tfsdk:"availability_zone"`
	K8sversion         types.String   `tfsdk:"kubernetes_version"`
	ClusterStatus      types.String   `tfsdk:"cluster_status"`
	Network            types.Object   `tfsdk:"network"`
	NodeGroups         types.List     `tfsdk:"node_groups"`
	SSHPublicKeyNames  []types.String `tfsdk:"ssh_public_key_names"`
	Storage            types.List     `tfsdk:"storages"`
	LoadBalancer       types.List     `tfsdk:"load_balancers"`
	UpgardeAvailable   types.Bool     `tfsdk:"upgrade_available"`
	UpgradableVersions []types.String `tfsdk:"upgrade_k8s_versions_available"`
}

type IKSClusterModel struct {
	ClusterUUID      types.String `tfsdk:"uuid"`
	Cloudaccount     types.String `tfsdk:"cloudaccount"`
	Name             types.String `tfsdk:"name"`
	AvailabilityZone types.String `tfsdk:"availability_zone"`
	K8sversion       types.String `tfsdk:"kubernetes_version"`
	ClusterStatus    types.String `tfsdk:"cluster_status"`
	Network          types.Object `tfsdk:"network"`
	UpgardeAvailable types.Bool   `tfsdk:"upgrade_available"`

	// UpgradableVersions []types.String `tfsdk:"upgrade_k8s_versions_available"`
}

var UpgradableVersionAttributes = []types.String{}

type IKSStorage struct {
	Size            types.Int64  `tfsdk:"size_in_gb"`
	State           types.String `tfsdk:"state"`
	StorageProvider types.String `tfsdk:"storage_provider"`
}

var IKStorageAttributes = map[string]attr.Type{
	"size_in_gb":       types.StringType,
	"state":            types.StringType,
	"storage_provider": types.StringType,
}

type ClusterNetwork struct {
	ClusterCIDR types.String `tfsdk:"cluster_cidr"`
	ClusterDNS  types.String `tfsdk:"cluster_dns"`
	EnableLB    types.Bool   `tfsdk:"enable_lb"`
	ServiceCIDR types.String `tfsdk:"service_cidr"`
}

func (m ClusterNetwork) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cluster_cidr": types.StringType,
		"cluster_dns":  types.StringType,
		"enable_lb":    types.BoolType,
		"service_cidr": types.StringType,
	}
}

type NodeGroup struct {
	ID                types.String           `tfsdk:"id"`
	Count             types.Int64            `tfsdk:"ng_count"`
	Name              types.String           `tfsdk:"name"`
	InstanceType      types.String           `tfsdk:"instance_type"`
	IMIId             types.String           `tfsdk:"imiid"`
	State             types.String           `tfsdk:"state"`
	UserDataURL       types.String           `tfsdk:"userdata_url"`
	SSHPublicKeyNames []types.String         `tfsdk:"ssh_public_key_names"`
	Interfaces        []NetworkInterfaceSpec `tfsdk:"interfaces"`
}

var NodeGroupAttributes = map[string]attr.Type{
	"id":            types.StringType,
	"ng_count":      types.Int64Type,
	"name":          types.StringType,
	"instance_type": types.StringType,
	"imiid":         types.StringType,
	"state":         types.StringType,
	"userdata_url":  types.StringType,
}

type IKSLoadBalancer struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	VipState types.String `tfsdk:"vip_state"`
	VipIp    types.String `tfsdk:"vip_ip"`
	Port     types.Int64  `tfsdk:"port"`
	PoolPort types.Int64  `tfsdk:"pool_port"`
	VipType  types.String `tfsdk:"vip_type"`
}

var IKSLoadLalancerAttributes = map[string]attr.Type{
	"id":        types.StringType,
	"name":      types.StringType,
	"vip_state": types.StringType,
	"vip_ip":    types.StringType,
	"port":      types.Int64Type,
	"pool_port": types.Int64Type,
	"vip_type":  types.StringType,
}
