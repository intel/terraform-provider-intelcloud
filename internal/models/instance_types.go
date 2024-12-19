package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type InstanceType struct {
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	InstanceCategory types.String `tfsdk:"instance_category"`
}
