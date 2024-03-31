package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &SpaceDataSource{}

// SpaceDataSource defines the data source implementation.
type SpaceDataSource struct {
	client *http.Client
}

// SpaceDataSourceModel describes the data source data model.
type SpaceDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Author       types.String `tfsdk:"author"`
	LastModified types.String `tfsdk:"last_modified"`
	Likes        types.Int64  `tfsdk:"likes"`
	Private      types.Bool   `tfsdk:"private"`
	SDK          types.String `tfsdk:"sdk"`
	Hardware     types.String `tfsdk:"hardware"`
	Storage      types.String `tfsdk:"storage"`
	SleepTime    types.Int64  `tfsdk:"sleep_time"`
}

func (d *SpaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (d *SpaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"author": schema.StringAttribute{
				Computed: true,
			},
			"last_modified": schema.StringAttribute{
				Computed: true,
			},
			"likes": schema.Int64Attribute{
				Computed: true,
			},
			"private": schema.BoolAttribute{
				Computed: true,
			},
			"sdk": schema.StringAttribute{
				Computed: true,
			},
			"hardware": schema.StringAttribute{
				Computed: true,
			},
			"storage": schema.StringAttribute{
				Computed: true,
			},
			"sleep_time": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (d *SpaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *SpaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SpaceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := fmt.Sprintf("https://huggingface.co/api/spaces/%s", data.ID.ValueString())
	log.Printf("[DEBUG] Requesting URL: %s", url)

	httpResp, err := d.client.Get(url)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read space, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	log.Printf("[DEBUG] Response Status Code: %d", httpResp.StatusCode)

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unexpected status code: %d", httpResp.StatusCode))
		return
	}

	var space map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&space)
	if err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode space JSON response, got error: %s", err))
		return
	}

	// Log the space JSON response for debugging
	log.Printf("[DEBUG] Space JSON Response: %+v", space)

	if id, ok := space["id"].(string); ok {
		data.Name = types.StringValue(id)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'id' field is missing or not a string in the space JSON response")
		return
	}

	if author, ok := space["author"].(string); ok {
		data.Author = types.StringValue(author)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'author' field is missing or not a string in the space JSON response")
		return
	}

	if lastModified, ok := space["lastModified"].(string); ok {
		data.LastModified = types.StringValue(lastModified)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'lastModified' field is missing or not a string in the space JSON response")
		return
	}

	if likes, ok := space["likes"].(float64); ok {
		data.Likes = types.Int64Value(int64(likes))
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'likes' field is missing or not a number in the space JSON response")
		return
	}

	if private, ok := space["private"].(bool); ok {
		data.Private = types.BoolValue(private)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'private' field is missing or not a boolean in the space JSON response")
		return
	}

	if sdk, ok := space["sdk"].(string); ok {
		data.SDK = types.StringValue(sdk)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'sdk' field is missing or not a string in the space JSON response")
		return
	}

	// Extract hardware, storage, and sleep time from the space JSON response
	if hardware, ok := space["hardware"].(string); ok {
		data.Hardware = types.StringValue(hardware)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'hardware' field is missing or not a string in the space JSON response")
		return
	}

	if storage, ok := space["storage"].(string); ok {
		data.Storage = types.StringValue(storage)
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'storage' field is missing or not a string in the space JSON response")
		return
	}

	if sleepTime, ok := space["sleepTime"].(float64); ok {
		data.SleepTime = types.Int64Value(int64(sleepTime))
	} else {
		resp.Diagnostics.AddError("Missing or Invalid Field", "The 'sleepTime' field is missing or not a number in the space JSON response")
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func NewSpaceDataSource() datasource.DataSource {
	return &SpaceDataSource{}
}
