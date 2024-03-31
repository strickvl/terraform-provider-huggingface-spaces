package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &SpaceResource{}
	_ resource.ResourceWithConfigure   = &SpaceResource{}
	_ resource.ResourceWithImportState = &SpaceResource{}
)

// SpaceResource defines the resource implementation.
type SpaceResource struct {
	client *http.Client
}

// SpaceResourceModel describes the resource data model.
type SpaceResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Private  types.Bool   `tfsdk:"private"`
	SDK      types.String `tfsdk:"sdk"`
	Template types.String `tfsdk:"template"`
}

func (r *SpaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_space"
}

func (r *SpaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"private": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"sdk": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"template": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// ... (Configure, Create, Read, Update, Delete, and ImportState functions)
func (r *SpaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SpaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SpaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := "https://huggingface.co/api/repos/create"

	reqBody := fmt.Sprintf(`{"type": "space", "name": "%s", "private": %t, "sdk": "%s", "template": "%s"}`,
		data.Name.ValueString(),
		data.Private.ValueBool(),
		data.SDK.ValueString(),
		data.Template.ValueString(),
	)

	httpResp, err := r.client.Post(url, "application/json", strings.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create space, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to create space, got status code: %d", httpResp.StatusCode))
		return
	}

	var responseData map[string]interface{}
	err = json.NewDecoder(httpResp.Body).Decode(&responseData)
	if err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode create space response, got error: %s", err))
		return
	}

	log.Printf("[DEBUG] Create Space Response: %+v", responseData)

	spaceName, ok := responseData["name"].(string)
	if !ok {
		resp.Diagnostics.AddError("Invalid Response", "Unable to extract space name from create space response")
		return
	}

	data.ID = types.StringValue(spaceName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SpaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// ... (Retrieve space details using the GET /api/spaces/{space_id} endpoint)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SpaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var state SpaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if the space needs to be renamed
	if state.Name.ValueString() != data.Name.ValueString() {
		url := "https://huggingface.co/api/repos/move"

		fromRepo := state.ID.ValueString()
		toRepo := fmt.Sprintf("%s/%s", strings.Split(state.ID.ValueString(), "/")[0], data.Name.ValueString())

		reqBody := fmt.Sprintf(`{"fromRepo": "%s", "toRepo": "%s", "type": "space"}`, fromRepo, toRepo)
		log.Printf("[DEBUG] Rename Space Request Body: %s", reqBody)

		httpResp, err := r.client.Post(url, "application/json", strings.NewReader(reqBody))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rename space, got error: %s", err))
			return
		}
		defer httpResp.Body.Close()

		log.Printf("[DEBUG] Rename Space Response Status Code: %d", httpResp.StatusCode)

		respBody, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
			return
		}
		log.Printf("[DEBUG] Rename Space Response Body: %s", string(respBody))

		if httpResp.StatusCode != http.StatusOK {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to rename space, got status code: %d", httpResp.StatusCode))
			return
		}

		state.ID = types.StringValue(toRepo)
		state.Name = data.Name
	}

	// Check if the space visibility needs to be updated
	if state.Private != data.Private {
		url := fmt.Sprintf("https://huggingface.co/api/spaces/%s/settings", data.ID.ValueString())

		reqBody := fmt.Sprintf(`{"private": %t}`, data.Private.ValueBool())
		log.Printf("[DEBUG] Update Space Visibility Request Body: %s", reqBody)

		httpReq, err := http.NewRequest(http.MethodPut, url, strings.NewReader(reqBody))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space visibility, got error: %s", err))
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		httpResp, err := r.client.Do(httpReq)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space visibility, got error: %s", err))
			return
		}
		defer httpResp.Body.Close()

		log.Printf("[DEBUG] Update Space Visibility Response Status Code: %d", httpResp.StatusCode)

		respBody, err := ioutil.ReadAll(httpResp.Body)
		if err != nil {
			resp.Diagnostics.AddError("API Response Error", fmt.Sprintf("Unable to read response body, got error: %s", err))
			return
		}
		log.Printf("[DEBUG] Update Space Visibility Response Body: %s", string(respBody))

		if httpResp.StatusCode != http.StatusOK {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update space visibility, got status code: %d", httpResp.StatusCode))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SpaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SpaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	url := "https://huggingface.co/api/repos/delete"

	reqBody := fmt.Sprintf(`{"type": "space", "name": "%s"}`, data.Name.ValueString())

	httpReq, err := http.NewRequest(http.MethodDelete, url, strings.NewReader(reqBody))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete space, got error: %s", err))
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := r.client.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete space, got error: %s", err))
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete space, got status code: %d", httpResp.StatusCode))
		return
	}
}

func (r *SpaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
