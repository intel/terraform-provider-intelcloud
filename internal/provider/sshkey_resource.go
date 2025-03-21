package provider

import (
	"context"
	"fmt"
	"time"

	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &sshKeyResource{}
	_ resource.ResourceWithConfigure = &sshKeyResource{}
)

// orderSSHKeyModel maps the resource schema data.
type sshKeyResourceModel struct {
	Metadata resourceMetadata `tfsdk:"metadata"`
	Spec     sshkeySpec       `tfsdk:"spec"`
}

// NewOrderFilesystem is a helper function to simplify the provider implementation.
func NewSSHKeyResource() resource.Resource {
	return &sshKeyResource{}
}

// orderResource is the resource implementation.
type sshKeyResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *sshKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *sshKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sshkey"
}

// Schema defines the schema for the resource.
func (r *sshKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"metadata": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"resourceid": schema.StringAttribute{
						Computed: true,
					},
					"cloudaccount": schema.StringAttribute{
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required: true,
					},
					"createdat": schema.StringAttribute{
						Computed: true,
					},
				},
			},
			"spec": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"ssh_public_key": schema.StringAttribute{
						Required: true,
					},
					"owner_email": schema.StringAttribute{
						Computed: true,
						Optional: true,
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *sshKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan sshKeyResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	inArg := itacservices.SSHKeyCreateRequest{
		Metadata: struct {
			Name string "json:\"name\""
		}{
			Name: plan.Metadata.Name.ValueString(),
		},
		Spec: struct {
			SSHPublicKey string "json:\"sshPublicKey\""
		}{
			SSHPublicKey: plan.Spec.SSHPublicKey.ValueString(),
		},
	}
	tflog.Info(ctx, "making a call to IDC Service for create sshkey")
	sshkeyCreateResp, err := r.client.CreateSSHkey(ctx, &inArg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.Metadata.CreatedAt = types.StringValue(time.Now().Format(time.RFC850))
	plan.Metadata.ResourceId = types.StringValue(sshkeyCreateResp.Metadata.ResourceId)
	plan.Metadata.Cloudaccount = types.StringValue(sshkeyCreateResp.Metadata.Cloudaccount)
	plan.Spec.OwnerEmail = types.StringValue(sshkeyCreateResp.Spec.OwnerEmail)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Read refreshes the Terraform state with the latest data.
func (r *sshKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state sshKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from IDC Service
	sshkey, err := r.client.GetSSHKeyByResourceId(ctx, state.Metadata.ResourceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC SSHKey resource",
			"Could not read IDC SSHKey resource ID "+state.Metadata.ResourceId.ValueString()+": "+err.Error(),
		)
		return
	}

	state.Metadata = resourceMetadata{
		ResourceId:   types.StringValue(sshkey.Metadata.ResourceId),
		Cloudaccount: types.StringValue(sshkey.Metadata.Cloudaccount),
		Name:         types.StringValue(sshkey.Metadata.Name),
	}
	state.Spec = sshkeySpec{
		SSHPublicKey: types.StringValue(sshkey.Spec.SSHPublicKey),
		OwnerEmail:   types.StringValue(sshkey.Spec.OwnerEmail),
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *sshKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *sshKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state sshKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the order from IDC Services
	err := r.client.DeleteSSHKeyByResourceId(ctx, state.Metadata.ResourceId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC SSHKey resource",
			"Could not delete IDC SSHKey resource ID "+state.Metadata.ResourceId.ValueString()+": "+err.Error(),
		)
		return
	}
}
