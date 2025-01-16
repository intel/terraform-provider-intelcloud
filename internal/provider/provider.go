// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &idcProvider{}
)

// idcProviderModel maps provider schema data to a Go type.
type idcProviderModel struct {
	Region       types.String `tfsdk:"region"`
	Cloudaccount types.String `tfsdk:"cloudaccount"`
	APIToken     types.String `tfsdk:"apitoken"`
	ClientId     types.String `tfsdk:"clientid"`
	ClientSecret types.String `tfsdk:"clientsecret"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &idcProvider{
			version: version,
		}
	}
}

// idcProvider is the provider implementation.
type idcProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *idcProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "intel-cloud"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *idcProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional: true,
			},
			"cloudaccount": schema.StringAttribute{
				Optional: true,
			},
			"apitoken": schema.StringAttribute{
				Optional: true,
			},
			"clientid": schema.StringAttribute{
				Optional: true,
			},
			"clientsecret": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *idcProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	region := os.Getenv("ITAC_REGION")
	cloudaccount := os.Getenv("ITAC_CLOUDACCOUNT")
	clientid := os.Getenv("ITAC_CLIENT_ID")
	clientsecret := os.Getenv("ITAC_CLIENT_SECRET")

	// Retrieve provider data from configuration
	var config idcProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Region.IsNull() {
		region = config.Region.ValueString()
	}

	if !config.Cloudaccount.IsNull() {
		cloudaccount = config.Cloudaccount.ValueString()
	}

	if !config.ClientId.IsNull() {
		clientid = config.ClientId.ValueString()
	}

	if !config.ClientSecret.IsNull() {
		clientsecret = config.ClientSecret.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if region == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("region"),
			"Missing ITAC API Region",
			"The provider cannot create the ITAC API client as there is a missing or empty value for the ITAC API region. "+
				"Set the region value in the configuration or use the ITAC_REGION environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if cloudaccount == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("cloudaccount"),
			"Missing ITAC Cloudaccount",
			"The provider cannot create the ITAC Cloudaccount as there is a missing or empty value for the ITAC Cloudaccount. "+
				"Set the host value in the configuration or use the ITAC_CLOUDACCOUNT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientid == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientid"),
			"Missing ITAC Client Id",
			"The provider cannot create the ITAC Client Id as there is a missing or empty value for the ITAC client id. "+
				"Set the clientid value in the configuration or use the ITAC_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientsecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("clientsecret"),
			"Missing ITAC Client secret",
			"The provider cannot create the ITAC client secret as there is a missing or empty value for the ITAC client secret "+
				"Set the clientsecret value in the configuration or use the ITAC_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	clientTokenEndpoint, serviceEndpoint := discoverITACServiceEndpoint(region)

	// Create a new HashiCups client using the configuration values
	client, err := itacservices.NewClient(ctx, &serviceEndpoint, &clientTokenEndpoint, &cloudaccount, &clientid, &clientsecret, &region)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create ITAC API Client",
			"An unexpected error occurred when creating the ITAC API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"ITAC Client Error: "+err.Error(),
		)
		return
	}

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *idcProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewFilesystemsDataSource,
		NewSSHKeysDataSource,
		NewInstanceDataSource,
		NewInstanceTypesDataSource,
		NewMachineImagesDataSource,
		NewKubernetesDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *idcProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewFilesystemResource,
		NewSSHKeyResource,
		NewComputeInstanceResource,
		NewIKSClusterResource,
		NewIKSNodeGroupResource,
		NewIKSLBResource,
		NewObjectStorageResource,
		NewObjectStorageUserResource,
	}
}

func discoverITACServiceEndpoint(region string) (string, string) {
	switch region {
	case "us-staging-1":
		return "https://client-token.staging.api.idcservice.net", "https://us-staging-1-sdk-api.eglb.intel.com"
	case "us-staging-3":
		return "https://staging-idc-us-3.eglb.intel.com", ""
	case "us-region-1":
		return "https://compute-us-region-1-api.cloud.intel.com", ""
	case "us-region-2":
		return "https://compute-us-region-2-api.cloud.intel.com", ""
	default:
		return "", ""
	}
}
