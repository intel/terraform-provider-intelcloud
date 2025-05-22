package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// model for imis
type ImisModel struct {
	InstanceTypeName string      `tfsdk:"instancetypename"`
	WorkerImi        []WorkerImi `tfsdk:"workerimi"`
}

// model for worker imi
type WorkerImi struct {
	ImiName      types.String `tfsdk:"iminame"`
	Info         types.String `tfsdk:"info"`
	IsDefaultImi types.Bool   `tfsdk:"isdefaultimi"`
}
