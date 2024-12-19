package models

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// InstanceModel maps IDC Compute Instance schema data.
type InstanceModel struct {
	ResourceId   types.String `tfsdk:"resource_id"`
	Cloudaccount types.String `tfsdk:"cloudaccount"`
	Name         types.String `tfsdk:"name"`
	Spec         InstanceSpec `tfsdk:"spec"`
	Status       types.String `tfsdk:"status"`
	Interfaces   types.List   `tfsdk:"interfaces"`
	SSHProxy     types.Object `tfsdk:"ssh_proxy"`
	AccessInfo   types.Object `tfsdk:"access_info"`
}

type InstanceSpec struct {
	InstanceGroup     types.String   `tfsdk:"instance_group"`
	InstanceType      types.String   `tfsdk:"instance_type"`
	MachineImage      types.String   `tfsdk:"machine_image"`
	SSHPublicKeyNames []types.String `tfsdk:"ssh_public_key_names"`
	UserData          types.String   `tfsdk:"user_data"`
}

type NetworkInterfaceSpec struct {
	Name types.String `tfsdk:"name"`
	VNet types.String `tfsdk:"vnet"`
}

type NetworkInterface struct {
	Addresses    types.String `tfsdk:"address"`
	DNSName      types.String `tfsdk:"dns_name"`
	Gateway      types.String `tfsdk:"gateway"`
	Name         types.String `tfsdk:"name"`
	PrefixLength types.Int64  `tfsdk:"prefix_length"`
	Subnet       types.String `tfsdk:"subnet"`
	VNet         types.String `tfsdk:"vnet"`
}

var ProviderInterfaceAttributes = map[string]attr.Type{
	"address":       types.StringType,
	"dns_name":      types.StringType,
	"gateway":       types.StringType,
	"name":          types.StringType,
	"prefix_length": types.Int64Type,
	"subnet":        types.StringType,
	"vnet":          types.StringType,
}

func (m NetworkInterface) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{}
}

type InstanceAccessInfoModel struct {
	Username types.String `tfsdk:"username"`
}

func (m InstanceAccessInfoModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"username": types.StringType,
	}
}

type SSHProxyModel struct {
	ProxyAddress types.String `tfsdk:"address"`
	ProxyPort    types.Int64  `tfsdk:"port"`
	ProxyUser    types.String `tfsdk:"user"`
}

func (m SSHProxyModel) AttributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"address": types.StringType,
		"port":    types.Int64Type,
		"user":    types.StringType,
	}
}
