package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type MachineImage struct {
	Name             types.String   `tfsdk:"name"`
	Description      types.String   `tfsdk:"description"`
	InstanceCategory []types.String `tfsdk:"instance_category"`
	InstanceTypes    []types.String `tfsdk:"instance_types"`
}
