package provider

import (
	"context"
	"fmt"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &objectStorageUserResource{}
	_ resource.ResourceWithConfigure   = &objectStorageUserResource{}
	_ resource.ResourceWithImportState = &objectStorageUserResource{}
)

// objectStorageUserResourceModel maps the resource schema data.
type objectStorageUserResourceModel struct {
	ID            types.String     `tfsdk:"id"`
	BucketId      types.String     `tfsdk:"bucket_id"`
	Cloudaccount  types.String     `tfsdk:"cloudaccount"`
	Name          types.String     `tfsdk:"name"`
	Status        types.String     `tfsdk:"status"`
	AllowActions  []types.String   `tfsdk:"allow_actions"`
	AllowPolicies ObjectUserPolicy `tfsdk:"allow_policies"`
	AccessInfo    types.Object     `tfsdk:"access_info"`
}

type ObjectUserPolicy struct {
	PathPrefix types.String   `tfsdk:"path_prefix"`
	Policies   []types.String `tfsdk:"policies"`
}

// NewObjectStorageResource is a helper function to simplify the provider implementation.
func NewObjectStorageUserResource() resource.Resource {
	return &objectStorageUserResource{}
}

// orderResource is the resource implementation.
type objectStorageUserResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *objectStorageUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*itacservices.IDCServicesClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *idcservices.IDCServicesClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Metadata returns the resource type name.
func (r *objectStorageUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_object_storage_bucket_user"
}

// Schema defines the schema for the resource.
func (r *objectStorageUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"bucket_id": schema.StringAttribute{
				Required: true,
			},
			"cloudaccount": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"allow_actions": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"allow_policies": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"path_prefix": schema.StringAttribute{
						Required: true,
					},
					"policies": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
					},
				},
			},
			"access_info": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"access_key": types.StringType,
					"secret_key": types.StringType,
				},
				Computed: true,
			},
		},
	}

}

// Create creates the resource and sets the initial Terraform state.
func (r *objectStorageUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan objectStorageUserResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actions := []string{}
	for _, a := range plan.AllowActions {
		actions = append(actions, a.ValueString())
	}

	perms := []string{}
	for _, p := range plan.AllowPolicies.Policies {
		perms = append(perms, p.ValueString())
	}

	bucketPolicy := []itacservices.BucketPolicy{
		{
			BucketId:    plan.BucketId.ValueString(),
			Actions:     actions,
			Permissions: perms,
			Prefix:      plan.AllowPolicies.PathPrefix.ValueString(),
		},
	}

	inArg := itacservices.ObjectUserCreateRequest{}
	inArg.Metadata.Name = plan.Name.ValueString()

	inArg.Spec = append(inArg.Spec, bucketPolicy...)

	tflog.Info(ctx, "making a call to IDC Service for create bucket")
	user, err := r.client.CreateObjectStorageUser(ctx, &inArg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Cloudaccount = types.StringValue(user.Metadata.Cloudaccount)
	plan.ID = types.StringValue(user.Metadata.UserId)
	plan.Status = types.StringValue(mapObjectUserStatus(user.Status.Phase))
	plan.Name = types.StringValue(user.Metadata.Name)

	creds := models.ObjectUserAccessModel{
		AccessKey: types.StringValue(user.Status.Principal.Credentials.AccessKey),
		SecretKey: types.StringValue(user.Status.Principal.Credentials.SecretKey),
	}

	plan.AccessInfo, diags = types.ObjectValueFrom(ctx, creds.AttributeTypes(), creds)
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
func (r *objectStorageUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state objectStorageUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from IDC Service
	user, err := r.client.GetObjectUserByUserId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Object Bucket user resource",
			"Could not read IDC Object Bucket user ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(user.Metadata.UserId)
	state.Cloudaccount = types.StringValue(user.Metadata.Cloudaccount)
	state.Name = types.StringValue(user.Metadata.Name)
	state.Status = types.StringValue(mapObjectUserStatus(user.Status.Phase))

	creds := models.ObjectUserAccessModel{
		AccessKey: types.StringValue(user.Status.Principal.Credentials.AccessKey),
		SecretKey: types.StringValue(user.Status.Principal.Credentials.SecretKey),
	}

	state.AccessInfo, diags = types.ObjectValueFrom(ctx, creds.AttributeTypes(), creds)
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
func (r *objectStorageUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *objectStorageUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *objectStorageUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state objectStorageUserResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the order from IDC Services
	err := r.client.DeleteObjectUserByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC Object Storage Bucket user resource",
			"Could not delete IDC Object Storage Bucket user resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}

func mapObjectUserStatus(status string) string {
	switch status {
	case "ObjectUserReady":
		return "ready"
	case "ObjectUserFailed":
		return "failed"
	case "BucketProvisioning":
		return "provisioning"
	case "ObjectUserDeleting":
		return "deleting"
	case "ObjectUserDeleted":
		return "deleted"
	default:
		return "unspecified"
	}
}
