package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &objectStorageResource{}
	_ resource.ResourceWithConfigure   = &objectStorageResource{}
	_ resource.ResourceWithImportState = &objectStorageResource{}
)

// objectstorageResourceModel maps the resource schema data.
type objectStorageResourceModel struct {
	ID              types.String   `tfsdk:"id"`
	Cloudaccount    types.String   `tfsdk:"cloudaccount"`
	Name            types.String   `tfsdk:"name"`
	Versioned       types.Bool     `tfsdk:"versioned"`
	Size            types.String   `tfsdk:"size"`
	Status          types.String   `tfsdk:"status"`
	PrivateEndpoint types.String   `tfsdk:"private_endpoint"`
	SecurityGroups  types.List     `tfsdk:"security_groups"`
	Timeouts        *timeoutsModel `tfsdk:"timeouts"`
}

// NewObjectStorageResource is a helper function to simplify the provider implementation.
func NewObjectStorageResource() resource.Resource {
	return &objectStorageResource{}
}

// orderResource is the resource implementation.
type objectStorageResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *objectStorageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *objectStorageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket"
}

// Schema defines the schema for the resource.
func (r *objectStorageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"versioned": schema.BoolAttribute{
				Required: true,
			},
			"size": schema.StringAttribute{
				Computed: true,
			},
			"private_endpoint": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"security_groups": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"gateway": schema.StringAttribute{
							Computed: true,
						},
						"prefix_length": schema.Int64Attribute{
							Computed: true,
						},
						"subnet": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"resource_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Timeout for objectstorage resource operations",
						Default:     stringdefault.StaticString(ObjectstorageResourceTimeout),
					},
				},
			},
		},
	}

}

// Create creates the resource and sets the initial Terraform state.
func (r *objectStorageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan objectStorageResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	createTimeout, err := plan.Timeouts.GetTimeouts(ObjectStorageResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse create timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	inArg := itacservices.ObjectBucketCreateRequest{
		Metadata: struct {
			Name string "json:\"name\""
		}{
			Name: plan.Name.ValueString(),
		},
		Spec: struct {
			Versioned    bool   "json:\"versioned\""
			InstanceType string "json:\"instanceType\""
		}{
			Versioned:    plan.Versioned.ValueBool(),
			InstanceType: "storage-object",
		},
	}
	tflog.Info(ctx, "making a call to IDC Service for create bucket")
	bucket, err := r.client.CreateObjectStorageBucket(ctx, &inArg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Cloudaccount = types.StringValue(bucket.Metadata.Cloudaccount)
	plan.ID = types.StringValue(bucket.Metadata.ResourceId)
	plan.Status = types.StringValue(mapObjectBucketStatus(bucket.Status.Phase))
	plan.PrivateEndpoint = types.StringValue(bucket.Status.Cluster.AccessEndpoint)
	plan.Size = types.StringValue(bucket.Spec.Request.Size)

	secGroups := []models.NetworkSecurityGroup{}
	for _, sg := range bucket.Status.SecurityGroups.NetworkFilterAllow {
		newSg := models.NetworkSecurityGroup{
			Gateway:      types.StringValue(sg.Gateway),
			PrefixLength: types.Int64Value(int64(sg.PrefixLength)),
			Subnet:       types.StringValue(sg.Subnet),
		}
		secGroups = append(secGroups, newSg)
	}
	plan.SecurityGroups, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.NetworkSecurityGroupAttributes), secGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure timeout block is preserved
	if plan.Timeouts != nil {
		plan.Timeouts = &timeoutsModel{
			ResourceTimeout: plan.Timeouts.ResourceTimeout,
		}
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *objectStorageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state objectStorageResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// use timeouts if requested by the user
	readTimeout, err := state.Timeouts.GetTimeouts(ObjectStorageResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse read timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	// Get refreshed order value from IDC Service
	bucket, err := r.client.GetObjectBucketByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Object Bucket resource",
			"Could not read IDC Object Bucket resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(bucket.Metadata.ResourceId)
	state.Cloudaccount = types.StringValue(bucket.Metadata.Cloudaccount)
	state.Name = types.StringValue(bucket.Metadata.Name)
	state.Size = types.StringValue(bucket.Spec.Request.Size)
	state.Versioned = types.BoolValue(bucket.Spec.Versioned)

	state.Status = types.StringValue(mapObjectBucketStatus(bucket.Status.Phase))
	state.PrivateEndpoint = types.StringValue(bucket.Status.Cluster.AccessEndpoint)

	secGroups := []models.NetworkSecurityGroup{}
	for _, sg := range bucket.Status.SecurityGroups.NetworkFilterAllow {
		newSg := models.NetworkSecurityGroup{
			Gateway:      types.StringValue(sg.Gateway),
			PrefixLength: types.Int64Value(int64(sg.PrefixLength)),
			Subnet:       types.StringValue(sg.Subnet),
		}
		secGroups = append(secGroups, newSg)
	}
	state.SecurityGroups, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.NetworkSecurityGroupAttributes), secGroups)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
func (r *objectStorageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *objectStorageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *objectStorageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state objectStorageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, err := state.Timeouts.GetTimeouts(ObjectStorageResourceName)
	if err != nil {
		resp.Diagnostics.AddError("Invalid timeout", "Could not parse delete timeout: "+err.Error())
	}
	// Use the timeout context
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	// Delete the order from IDC Services
	err = r.client.DeleteBucketByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC Object Storage Bucket resource",
			"Could not delete IDC Object Storage Bucket resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func mapObjectBucketStatus(fsStatus string) string {
	switch fsStatus {
	case "BucketReady":
		return "ready"
	case "BucketFailed":
		return "failed"
	case "BucketProvisioning":
		return "provisioning"
	case "BucketDeleting":
		return "deleting"
	case "BucketDeleted":
		return "deleted"
	default:
		return "unspecified"
	}
}
