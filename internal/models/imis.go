package models

import "github.com/hashicorp/terraform-plugin-framework/types"

// model for imis
type ImisModel struct {
	InstanceTypeName string      `tfsdk:"instancetypename"`
	WorkerImi        []WorkerImi `tfsdk:"workerImi"`
}

// model for worker imi
type WorkerImi struct {
	ImiName      types.String `tfsdk:"imiName"`
	Info         types.String `tfsdk:"info"`
	IsDefaultImi types.Bool   `tfsdk:"isDefaultImi"`
}
