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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &computeInstanceResource{}
	_ resource.ResourceWithConfigure = &computeInstanceResource{}
)

// orderFilesystemModel maps the resource schema data.
type computeInstanceResourceModel struct {
	ID               types.String         `tfsdk:"id"`
	Cloudaccount     types.String         `tfsdk:"cloudaccount"`
	Name             types.String         `tfsdk:"name"`
	AvailabilityZone types.String         `tfsdk:"availability_zone"`
	Spec             *models.InstanceSpec `tfsdk:"spec"`
	Status           types.String         `tfsdk:"status"`
	Interfaces       types.List           `tfsdk:"interfaces"`
	SSHProxy         types.Object         `tfsdk:"ssh_proxy"`
	AccessInfo       types.Object         `tfsdk:"access_info"`
}

// NewOrderFilesystem is a helper function to simplify the provider implementation.
func NewComputeInstanceResource() resource.Resource {
	return &computeInstanceResource{}
}

// computeInstanceResource is the resource implementation.
type computeInstanceResource struct {
	client *itacservices.IDCServicesClient
}

// Configure adds the provider configured client to the resource.
func (r *computeInstanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *computeInstanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_instance"
}

// Schema defines the schema for the resource.
func (r *computeInstanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
					"instance_group": schema.StringAttribute{
						Optional: true,
					},
					"instance_type": schema.StringAttribute{
						Required: true,
					},
					"machine_image": schema.StringAttribute{
						Required: true,
					},
					"ssh_public_key_names": schema.ListAttribute{
						ElementType: types.StringType,
						Required:    true,
					},
					"user_data": schema.StringAttribute{
						Optional: true,
					},
				},
			},
			"interfaces": schema.ListNestedAttribute{
				Optional: true,
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"dns_name": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"gateway": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"prefix_length": schema.Int64Attribute{
							Computed: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.UseStateForUnknown(),
							},
						},
						"subnet": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"vnet": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"access_info": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"username": types.StringType,
				},
				Computed: true,
			},
			"ssh_proxy": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"address": types.StringType,
					"port":    types.Int64Type,
					"user":    types.StringType,
				},
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *computeInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan computeInstanceResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "making a call to IDC Service to createVnetIfNotExist")
	vnetResp, err := r.client.CreateVNetIfNotFound(ctx)
	if err != nil || vnetResp == nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	sshKeys := []string{}
	for _, k := range plan.Spec.SSHPublicKeyNames {
		sshKeys = append(sshKeys, k.ValueString())
	}

	inArg := itacservices.InstanceCreateRequest{
		Metadata: struct {
			Name string "json:\"name\""
		}{
			Name: plan.Name.ValueString(),
		},
		Spec: struct {
			AvailabilityZone string "json:\"availabilityZone\""
			InstanceGroup    string "json:\"instanceGroup,omitempty\""
			InstanceType     string "json:\"instanceType\""
			Interfaces       []struct {
				Name string "json:\"name\""
				VNet string "json:\"vNet\""
			} "json:\"interfaces\""
			MachineImage      string   "json:\"machineImage\""
			SshPublicKeyNames []string "json:\"sshPublicKeyNames\""
			UserData          string   "json:\"userData,omitempty\""
		}{
			AvailabilityZone: fmt.Sprintf("%sa", *r.client.Region),
			InstanceGroup:    plan.Spec.InstanceGroup.ValueString(),
			Interfaces: []struct {
				Name string "json:\"name\""
				VNet string "json:\"vNet\""
			}{
				{
					Name: "eth0",
					VNet: fmt.Sprintf("%sa-default", *r.client.Region),
				},
			},
			InstanceType:      plan.Spec.InstanceType.ValueString(),
			MachineImage:      plan.Spec.MachineImage.ValueString(),
			UserData:          plan.Spec.UserData.ValueString(),
			SshPublicKeyNames: sshKeys,
		},
	}

	tflog.Info(ctx, "making a call to IDC Service for create instance")
	instResp, err := r.client.CreateInstance(ctx, &inArg, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating order",
			"Could not create order, unexpected error: "+err.Error(),
		)
		return
	}

	// Map response body to schema and populate Computed attribute values
	plan.ID = types.StringValue(instResp.Metadata.ResourceId)
	plan.Cloudaccount = types.StringValue(instResp.Metadata.Cloudaccount)
	plan.Status = types.StringValue(instResp.Status.Phase)
	plan.AvailabilityZone = types.StringValue(instResp.Spec.AvailabilityZone)

	accessInfoMap := models.InstanceAccessInfoModel{
		Username: types.StringValue(instResp.Status.UserName),
	}

	plan.AccessInfo, diags = types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshProxyMap := models.SSHProxyModel{
		ProxyAddress: types.StringValue(instResp.Status.SSHProxy.Address),
		ProxyPort:    types.Int64Value(instResp.Status.SSHProxy.Port),
		ProxyUser:    types.StringValue(instResp.Status.SSHProxy.User),
	}
	plan.SSHProxy, diags = types.ObjectValueFrom(ctx, sshProxyMap.AttributeTypes(), sshProxyMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	infs := []models.NetworkInterface{}
	for _, nic := range instResp.Status.Interfaces {
		// currently we ssume a single interface will have a single address
		addr := ""
		if len(nic.Addresses) > 0 {
			addr = nic.Addresses[0]
		}
		inf := models.NetworkInterface{
			Addresses:    types.StringValue(addr),
			DNSName:      types.StringValue(nic.DNSName),
			Gateway:      types.StringValue(nic.Gateway),
			Name:         types.StringValue(nic.Name),
			PrefixLength: types.Int64Value(int64(nic.PrefixLength)),
			Subnet:       types.StringValue(nic.Subnet),
			VNet:         types.StringValue(nic.VNet),
		}
		infs = append(infs, inf)
	}
	plan.Interfaces, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.ProviderInterfaceAttributes), infs)
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
func (r *computeInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state computeInstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from IDC Service
	instance, err := r.client.GetInstanceByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading IDC Compute Instance resource",
			"Could not read IDC Compute Instance resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "instance read request response", map[string]any{"instance": instance, "resourceId": state.ID.ValueString()})

	// state = orderInstanceModel{}
	// state.Instance = models.InstanceModel{}

	state.Cloudaccount = types.StringValue(instance.Metadata.Cloudaccount)
	state.ID = types.StringValue(instance.Metadata.ResourceId)
	state.Name = types.StringValue(instance.Metadata.Name)
	state.AvailabilityZone = types.StringValue(instance.Spec.AvailabilityZone)
	state.Spec = &models.InstanceSpec{
		InstanceGroup: types.StringValue(instance.Spec.InstanceGroup),
		InstanceType:  types.StringValue(instance.Spec.InstanceType),
		MachineImage:  types.StringValue(instance.Spec.MachineImage),
		UserData:      types.StringValue(instance.Spec.UserData),
	}

	for _, k := range instance.Spec.SshPublicKeyNames {
		state.Spec.SSHPublicKeyNames = append(state.Spec.SSHPublicKeyNames, types.StringValue(k))
	}

	state.Status = types.StringValue(instance.Status.Phase)

	accessInfoMap := models.InstanceAccessInfoModel{
		Username: types.StringValue(instance.Status.UserName),
	}

	state.AccessInfo, diags = types.ObjectValueFrom(ctx, accessInfoMap.AttributeTypes(), accessInfoMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshProxyMap := models.SSHProxyModel{
		ProxyAddress: types.StringValue(instance.Status.SSHProxy.Address),
		ProxyPort:    types.Int64Value(instance.Status.SSHProxy.Port),
		ProxyUser:    types.StringValue(instance.Status.SSHProxy.User),
	}
	state.SSHProxy, diags = types.ObjectValueFrom(ctx, sshProxyMap.AttributeTypes(), sshProxyMap)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	infs := []models.NetworkInterface{}
	for _, nic := range instance.Status.Interfaces {
		inf := models.NetworkInterface{
			Addresses:    types.StringValue(nic.Addresses[0]),
			DNSName:      types.StringValue(nic.DNSName),
			Gateway:      types.StringValue(nic.Gateway),
			Name:         types.StringValue(nic.Name),
			PrefixLength: types.Int64Value(int64(nic.PrefixLength)),
			Subnet:       types.StringValue(nic.Subnet),
			VNet:         types.StringValue(nic.VNet),
		}
		// for _, addr := range nic.Addresses {
		// 	inf.Addresses = append(inf.Addresses, types.StringValue(addr))
		// }
		infs = append(infs, inf)
	}
	state.Interfaces, diags = types.ListValueFrom(ctx, types.ObjectType{}.WithAttributeTypes(models.ProviderInterfaceAttributes), infs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "instance read request state ready", map[string]any{"status": state.Status.ValueString(), "resourceId": state.ID.ValueString()})

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *computeInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *computeInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *computeInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state computeInstanceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the order from IDC Services
	err := r.client.DeleteInstanceByResourceId(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting IDC Filesystem resource",
			"Could not delete IDC Filesystem resource ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
}
