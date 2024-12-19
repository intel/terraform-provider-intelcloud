package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type ResourceMetadata struct {
	ResourceId   types.String `tfsdk:"resourceid"`
	Cloudaccount types.String `tfsdk:"cloudaccount"`
	Name         types.String `tfsdk:"name"`
	CreatedAt    types.String `tfsdk:"createdat"`
}
