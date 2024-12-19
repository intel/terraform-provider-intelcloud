package provider

import (
	"context"
	"fmt"
	"strings"

	"terraform-provider-intelcloud/internal/models"
	"terraform-provider-intelcloud/pkg/itacservices"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewMachineImagesDataSource() datasource.DataSource {
	return &machineImagesDataSource{}
}

type machineImagesDataSource struct {
	client *itacservices.IDCServicesClient
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &machineImagesDataSource{}
	_ datasource.DataSourceWithConfigure = &machineImagesDataSource{}
)

// storagesDataSourceModel maps the data source schema data.
type machineImagesDataSourceModel struct {
	MostRecent bool                  `tfsdk:"most_recent"`
	Filters    []KVFilter            `tfsdk:"filters"`
	Result     *models.MachineImage  `tfsdk:"result"`
	Images     []models.MachineImage `tfsdk:"items"`
}

type KVFilter struct {
	Key    string   `tfsdk:"name"`
	Values []string `tfsdk:"values"`
}

// Configure adds the provider configured client to the data source.
func (d *machineImagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *machineImagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_images"
}

func (d *machineImagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"most_recent": schema.BoolAttribute{
				Required: true,
			},
			"filters": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true,
						},
						"values": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
			"result": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed: true,
					},
					"description": schema.StringAttribute{
						Computed: true,
					},
					"instance_category": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
					"instance_types": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"items": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"instance_category": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"instance_types": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *machineImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state machineImagesDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	osImgs, err := d.client.GetMachineImages(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read ITAC Machine Images",
			err.Error(),
		)
		return
	}

	allImages := []models.MachineImage{}

	for _, img := range osImgs.Items {
		if img.Hidden {
			continue
		}
		tfImg := models.MachineImage{
			Name:        types.StringValue(img.Metadata.Name),
			Description: types.StringValue(img.Spec.Description),
		}
		for _, i := range img.Spec.InstanceCategories {
			tfImg.InstanceCategory = append(tfImg.InstanceCategory, types.StringValue(i))
		}
		for _, t := range img.Spec.InstanceTypes {
			tfImg.InstanceTypes = append(tfImg.InstanceCategory, types.StringValue(t))
		}
		allImages = append(allImages, tfImg)
	}
	filteredImages := filterImages(allImages, state.Filters)

	state.Images = append(state.Images, filteredImages...)
	state.Result = &filteredImages[0]

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func filterImages(allImages []models.MachineImage, filters []KVFilter) []models.MachineImage {
	filteredImages := allImages

	for _, filter := range filters {
		switch filter.Key {
		case "name":
			filteredImages = filterByName(filteredImages, filter.Values)
		case "machine-type":
			filteredImages = filterByMachineType(filteredImages, filter.Values)
		default:
			return allImages
		}
	}
	return filteredImages
}

func filterByName(allImages []models.MachineImage, values []string) []models.MachineImage {
	filteredImages := []models.MachineImage{}
	for _, v := range values {
		for _, img := range allImages {
			if strings.Contains(img.Name.ValueString(), v) {
				filteredImages = append(filteredImages, img)
			}
		}
	}
	return filteredImages
}

func filterByMachineType(allImages []models.MachineImage, values []string) []models.MachineImage {
	filteredImages := []models.MachineImage{}
	for _, v := range values {
		for _, img := range allImages {
			for _, inst := range img.InstanceTypes {
				if strings.Contains(inst.ValueString(), v) {
					filteredImages = append(filteredImages, img)
				}
			}
		}
	}
	return filteredImages
}
