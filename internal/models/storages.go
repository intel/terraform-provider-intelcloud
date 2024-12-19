package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FilesystemCreateRequest struct {
	Name        types.String   `tfsdk:"name"`
	Description types.String   `tfsdk:"description"`
	Spec        FilesystemSpec `tfsdk:"spec"`
}

type FilesystemModel struct {
	ResourceId       types.String   `tfsdk:"resource_id"`
	Cloudaccount     types.String   `tfsdk:"cloudaccount"`
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	Spec             FilesystemSpec `tfsdk:"spec"`
	Status           types.String   `tfsdk:"status"`
	ClusterInfo      types.Object   `tfsdk:"cluster_info"`
	AccessInfo       types.Object   `tfsdk:"access_info"`
}

type FilesystemSpec struct {
	Size           types.Int64  `tfsdk:"size_in_tb"`
	AccessMode     types.String `tfsdk:"access_mode"`
	Encrypted      types.Bool   `tfsdk:"encrypted"`
	FilesystemType types.String `tfsdk:"filesystem_type"`
	StorageClass   types.String `tfsdk:"storage_class"`
}

type ObjectStoreSpec struct {
	Versioned types.Bool `tfsdk:"versioned"`
}

type FilesystemClusteModel struct {
	ClusterAddress types.String `tfsdk:"cluster_address"`
	ClusterVersion types.String `tfsdk:"cluster_version"`
}

type FilesystemAccessModel struct {
	Namespace  types.String `tfsdk:"namespace"`
	Filesystem types.String `tfsdk:"filesystem_name"`
	Username   types.String `tfsdk:"username"`
	Password   types.String `tfsdk:"password"`
}

func (m FilesystemAccessModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"namespace":       types.StringType,
		"filesystem_name": types.StringType,
		"username":        types.StringType,
		"password":        types.StringType,
	}
}

func (m FilesystemClusteModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"cluster_address": types.StringType,
		"cluster_version": types.StringType,
	}
}

type NetworkSecurityGroup struct {
	Gateway      types.String `tfsdk:"gateway"`
	PrefixLength types.Int64  `tfsdk:"prefix_length"`
	Subnet       types.String `tfsdk:"subnet"`
}

var NetworkSecurityGroupAttributes = map[string]attr.Type{
	"gateway":       types.StringType,
	"prefix_length": types.Int64Type,
	"subnet":        types.StringType,
}

type ObjectUserAccessModel struct {
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}

func (m ObjectUserAccessModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"access_key": types.StringType,
		"secret_key": types.StringType,
	}
}
