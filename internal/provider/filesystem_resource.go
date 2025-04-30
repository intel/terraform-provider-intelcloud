package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &filesystemResource{}
	_ resource.ResourceWithConfigure   = &filesystemResource{}
	_ resource.ResourceWithImportState = &filesystemResource{}
)

// filesystemModel maps the resource schema data.
type filesystemResourceModel struct {
	ID               types.String           `tfsdk:"id"`
	Cloudaccount     types.String           `tfsdk:"cloudaccount"`
	Name             types.String           `tfsdk:"name"`
	AvailabilityZone types.String           `tfsdk:"availability_zone"`
	Spec             *models.FilesystemSpec `tfsdk:"spec"`
	Status           types.String           `tfsdk:"status"`
	ClusterInfo      types.Object           `tfsdk:"cluster_info"`
	AccessInfo       types.Object           `tfsdk:"access_info"`
	Timeouts         *timeoutsModel         `tfsdk:"timeouts"`
}

// NewFilesystemResource is a helper function to simplify the provider implementation.
func NewFilesystemResource() resource.Resource {
	return &filesystemResource{}
}

// orderResource is the resource implementation.
type filesystemResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *filesystemResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Metadata returns the resource type name.
func (r *filesystemResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_filesystem"
}

// Schema defines the schema for the resource.
func (r *filesystemResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"cloudaccount": schema.StringAttribute{
				Computed: true,
			},
			"availability_zone": schema.StringAttribute{
				Computed: true,
			},
			"spec": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"size_in_tb": schema.Int64Attribute{
						Required: true,
					},
					"access_mode": schema.StringAttribute{
						Computed: true,
						Default:  stringdefault.StaticString("ReadWrite"),
					},
					"encrypted": schema.BoolAttribute{
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"storage_class": schema.StringAttribute{
						Computed: true,
					},
					"filesystem_type": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"cluster_info": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"cluster_address": types.StringType,
					"cluster_version": types.StringType,
				},
				Computed: true,
			},
			"access_info": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"namespace":       types.StringType,
					"filesystem_name": types.StringType,
					"username":        types.StringType,
					"password":        types.StringType,
				},
				Computed: true,
				Optional: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"resource_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timeout for resource operation, supports 1s, 2m, 3h etc.",
						Default:     stringdefault.StaticString(FileSystemResourceTimeout),
					},
				},
			},
		},
	}

}

// Create creates the resource and sets the initial Terraform state.
func (r *filesystemResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan filesystemResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	createTimeout, err := plan.Timeouts.GetTimeouts(FilesystemResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse create timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	inArg := itacservices.FilesystemCreateRequest{
		Metadata: struct {
			Name string "json:\"name\""
		}{
			Name: plan.Name.ValueString(),
		},
		Spec: struct {
			Request struct {
				Size string "json:\"storage\""
			} "json:\"request\""
			StorageClass     string "json:\"storageClass\""
			AccessMode       string "json:\"accessModes\""
			FilesystemType   string "json:\"filesystemType\""
			InstanceType     string "json:\"instanceType\""
			Encrypted        bool   "json:\"Encrypted\""
			AvailabilityZone string "json:\"availabilityZone\""
		}{
			Request: struct {
				Size string "json:\"storage\""
			}{
				Size: fmt.Sprintf("%dTB", plan.Spec.Size.ValueInt64()),
			},
			FilesystemType:   "ComputeGeneral",
			InstanceType:     "storage-file", // hard-coded for now
			AvailabilityZone: fmt.Sprintf("%sa", *r.client.Region),
			StorageClass:     "GeneralPurpose",
			AccessMode:       plan.Spec.AccessMode.ValueString(),
			Encrypted:        plan.Spec.Encrypted.ValueBool(),
		},
	}
	tflog.Info(ctx, "making a call to IDC Service for create filesystem")
	fsResp, err := r.client.CreateFilesystem(ctx, &inArg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.AvailabilityZone = types.StringValue(fsResp.Spec.AvailabilityZone)
	plan.Cloudaccount = types.StringValue(fsResp.Metadata.Cloudaccount)
	plan.ID = types.StringValue(fsResp.Metadata.ResourceId)
	plan.Status = types.StringValue(mapFilesystemStatus(fsResp.Status.Phase))

	plan.Spec.FilesystemType = types.StringValue(fsResp.Spec.FilesystemType)
	plan.Spec.StorageClass = types.StringValue(fsResp.Spec.StorageClass)
	sizeStr := strings.Split(fsResp.Spec.Request.Size, "TB")[0]
	size, _ := strconv.ParseInt(sizeStr, 10, 64)
	plan.Spec.Size = types.Int64Value(size)
	clusterInfoMap := models.FilesystemClusteModel{
		ClusterAddress: types.StringValue(fsResp.Status.Mount.ClusterAddr),
		ClusterVersion: types.StringValue(fsResp.Status.Mount.ClusterVersion),
	}

	accessInfoMap := models.FilesystemAccessModel{
		Namespace:  types.StringValue(fsResp.Status.Mount.Namespace),
		Filesystem: types.StringValue(fsResp.Status.Mount.FilesystemName),
		Username:   types.StringValue(fsResp.Status.Mount.UserName),
		Password:   types.StringValue(fsResp.Status.Mount.Password),
	}

	plan.ClusterInfo, diags = types.ObjectValueFrom(ctx, clusterInfoMap.AttributeTypes(), clusterInfoMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.AccessInfo, diags = types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *filesystemResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var orig filesystemResourceModel

	diags := req.State.Get(ctx, &orig)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	readTimeout, err := orig.Timeouts.GetTimeouts(FilesystemResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// Get refreshed order value from IDC Service
	filesystem, err := r.client.GetFilesystemByResourceId(ctx, orig.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Filesystem resource",
			"Could not read IDC Filesystem resource ID "+orig.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state, err := refreshFilesystemResourceModel(ctx, filesystem)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Filesystem resource",
			"Could not read IDC Filesystem resource ID "+orig.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *filesystemResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state filesystemResourceModel

	// Retrieve the desired configuration from the plan
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve the current state
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Detect changes in the "spec" field
	if !plan.Spec.Size.Equal(state.Spec.Size) {
		tflog.Info(ctx, "Detected change in filesystem spec, updating resource")

		// use timeouts if requested by the user
		updateTimeout, err := plan.Timeouts.GetTimeouts(FilesystemResourceName)
		if err != nil {
			resp.Diagnostics.AddError("Invalid timeout", "Could not parse update timeout: "+err.Error())
		}
		// Use the timeout context
		ctx, cancel := context.WithTimeout(ctx, updateTimeout)
		defer cancel()

		inArg := itacservices.FilesystemUpdateRequest{
			Metadata: struct {
				Name string "json:\"name\""
			}{
				Name: plan.Name.ValueString(),
			},
			Payload: itacservices.FileSystemUpdatePayload{
				Spec: struct {
					Request struct {
						Size string "json:\"storage\""
					} "json:\"request\""
				}{
					Request: struct {
						Size string "json:\"storage\""
					}{
						Size: fmt.Sprintf("%dTB", plan.Spec.Size.ValueInt64()),
					},
				},
			},
		}

		tflog.Info(ctx, "making a call to IDC Service for update filesystem", map[string]any{"Payload": inArg})
		err = r.client.UpdateFilesystem(ctx, &inArg)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error creating order",
				"Could not create order, unexpected error: "+err.Error(),
			)
			return
		}

		// Get refreshed order value from IDC Service
		filesystem, err := r.client.GetFilesystemByResourceId(ctx, state.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading IDC Filesystem resource",
				"Could not read IDC Filesystem resource ID "+state.ID.ValueString()+": "+err.Error(),
			)
			return
		}

		currState, err := refreshFilesystemResourceModel(ctx, filesystem)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading IDC Filesystem resource",
				"Could not read IDC Filesystem resource ID "+plan.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		currState.Spec.Size = plan.Spec.Size

		// Set refreshed state
		diags = resp.State.Set(ctx, currState)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	tflog.Info(ctx, "no change detected change in filesystem spec, skipping update")
}

func (r *filesystemResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *filesystemResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state filesystemResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	deleteTimeout, err := state.Timeouts.GetTimeouts(FilesystemResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse delete timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the order from IDC Services
	err = r.client.DeleteFilesystemByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC Instance resource",
			"Could not delete IDC Instance resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func mapFilesystemStatus(fsStatus string) string {
	switch fsStatus {
	case "FSReady":
		return "ready"
	case "FSFailed":
		return "failed"
	case "FSProvisioning":
		return "provisioning"
	case "FSDeleting":
		return "deleting"
	case "FSDeleted":
		return "deleted"
	default:
		return "unspecified"
	}
}

func refreshFilesystemResourceModel(ctx context.Context, filesystem *itacservices.Filesystem) (*filesystemResourceModel, error) {

	state := &filesystemResourceModel{}
	var diags diag.Diagnostics

	sizeStr := strings.Split(filesystem.Spec.Request.Size, "TB")[0]
	size, _ := strconv.ParseInt(sizeStr, 10, 64)

	state.ID = types.StringValue(filesystem.Metadata.ResourceId)
	state.Cloudaccount = types.StringValue(filesystem.Metadata.Cloudaccount)
	state.Name = types.StringValue(filesystem.Metadata.Name)

	state.AvailabilityZone = types.StringValue(filesystem.Spec.AvailabilityZone)
	state.Spec = &models.FilesystemSpec{
		Size:           types.Int64Value(size),
		AccessMode:     types.StringValue(filesystem.Spec.AccessMode),
		Encrypted:      types.BoolValue(filesystem.Spec.Encrypted),
		FilesystemType: types.StringValue(filesystem.Spec.FilesystemType),
		StorageClass:   types.StringValue(filesystem.Spec.StorageClass),
	}

	state.Status = types.StringValue(mapFilesystemStatus(filesystem.Status.Phase))

	clusterInfoMap := models.FilesystemClusteModel{
		ClusterAddress: types.StringValue(filesystem.Status.Mount.ClusterAddr),
		ClusterVersion: types.StringValue(filesystem.Status.Mount.ClusterVersion),
	}

	accessInfoMap := models.FilesystemAccessModel{
		Namespace:  types.StringValue(filesystem.Status.Mount.Namespace),
		Filesystem: types.StringValue(filesystem.Status.Mount.FilesystemName),
		Username:   types.StringValue(filesystem.Status.Mount.UserName),
		Password:   types.StringValue(filesystem.Status.Mount.Password),
	}

	state.ClusterInfo, diags = types.ObjectValueFrom(ctx, clusterInfoMap.AttributeTypes(), clusterInfoMap)
	if diags.HasError() {
		return state, fmt.Errorf("error parsing values")
	}

	state.AccessInfo, diags = types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
	if diags.HasError() {
		return state, fmt.Errorf("error parsing values")
	}
	return state, nil
}
