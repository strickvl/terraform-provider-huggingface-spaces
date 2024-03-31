package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Private   types.Bool   `tfsdk:"private"`
	SDK       types.String `tfsdk:"sdk"`
	Template  types.String `tfsdk:"template"`
	Secrets   types.Map    `tfsdk:"secrets"`
	Variables types.Map    `tfsdk:"variables"`
	Hardware  types.String `tfsdk:"hardware"`
	Storage   types.String `tfsdk:"storage"`
	SleepTime types.Int64  `tfsdk:"sleep_time"`
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
			"secrets": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"hardware": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"storage": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"sleep_time": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

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

	reqBody := fmt.Sprintf(`{"type": "space", "name": "%s", "private": %t, "sdk": "%s", "template": "%s", "hardware": "%s", "storage": "%s", "sleepTime": %d}`,
		data.Name.ValueString(),
		data.Private.ValueBool(),
		data.SDK.ValueString(),
		data.Template.ValueString(),
		data.Hardware.ValueString(),
		data.Storage.ValueString(),
		data.SleepTime.ValueInt64(),
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

	// Add secrets
	if !data.Secrets.IsNull() && !data.Secrets.IsUnknown() {
		secretsMap := data.Secrets.Elements()
		for key, value := range secretsMap {
			secretURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/secrets", data.ID.ValueString())
			secretReqBody := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value.(types.String).ValueString())
			secretResp, err := r.client.Post(secretURL, "application/json", strings.NewReader(secretReqBody))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add secret, got error: %s", err))
				return
			}
			defer secretResp.Body.Close()

			if secretResp.StatusCode != http.StatusOK {
				resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add secret, got status code: %d", secretResp.StatusCode))
				return
			}
		}
	}

	// Add variables
	if !data.Variables.IsNull() && !data.Variables.IsUnknown() {
		variablesMap := data.Variables.Elements()
		for key, value := range variablesMap {
			variableURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/variables", data.ID.ValueString())
			variableReqBody := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value.(types.String).ValueString())
			variableResp, err := r.client.Post(variableURL, "application/json", strings.NewReader(variableReqBody))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add variable, got error: %s", err))
				return
			}
			defer variableResp.Body.Close()

			if variableResp.StatusCode != http.StatusOK {
				resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add variable, got status code: %d", variableResp.StatusCode))
				return
			}
		}
	}

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

	// Update secrets
	if !data.Secrets.IsNull() && !data.Secrets.IsUnknown() {
		// Delete existing secrets
		secretsURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/secrets", data.ID.ValueString())
		secretsResp, err := r.client.Get(secretsURL)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to retrieve secrets, got error: %s", err))
			return
		}
		defer secretsResp.Body.Close()

		if secretsResp.StatusCode == http.StatusOK {
			var existingSecrets map[string]interface{}
			err = json.NewDecoder(secretsResp.Body).Decode(&existingSecrets)
			if err != nil {
				resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode secrets response, got error: %s", err))
				return
			}

			for key := range existingSecrets {
				deleteSecretURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/secrets", data.ID.ValueString())
				deleteSecretReqBody := fmt.Sprintf(`{"key": "%s"}`, key)
				deleteSecretReq, err := http.NewRequest(http.MethodDelete, deleteSecretURL, strings.NewReader(deleteSecretReqBody))
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete secret, got error: %s", err))
					return
				}
				deleteSecretReq.Header.Set("Content-Type", "application/json")

				deleteSecretResp, err := r.client.Do(deleteSecretReq)
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete secret, got error: %s", err))
					return
				}
				defer deleteSecretResp.Body.Close()

				if deleteSecretResp.StatusCode != http.StatusOK {
					resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete secret, got status code: %d", deleteSecretResp.StatusCode))
					return
				}
			}
		}

		// Add new secrets
		secretsMap := data.Secrets.Elements()
		stateSecretsMap := make(map[string]attr.Value)
		for key, value := range secretsMap {
			secretURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/secrets", data.ID.ValueString())
			secretReqBody := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value.(types.String).ValueString())
			secretResp, err := r.client.Post(secretURL, "application/json", strings.NewReader(secretReqBody))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add secret, got error: %s", err))
				return
			}
			defer secretResp.Body.Close()

			if secretResp.StatusCode != http.StatusOK {
				resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add secret, got status code: %d", secretResp.StatusCode))
				return
			}
			stateSecretsMap[key] = value
		}
		state.Secrets, _ = types.MapValue(types.StringType, stateSecretsMap)
	}

	// Update variables
	if !data.Variables.IsNull() && !data.Variables.IsUnknown() {
		// Delete existing variables
		variablesURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/variables", data.ID.ValueString())
		variablesResp, err := r.client.Get(variablesURL)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to retrieve variables, got error: %s", err))
			return
		}
		defer variablesResp.Body.Close()

		if variablesResp.StatusCode == http.StatusOK {
			var existingVariables map[string]interface{}
			err = json.NewDecoder(variablesResp.Body).Decode(&existingVariables)
			if err != nil {
				resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode variables response, got error: %s", err))
				return
			}

			for key := range existingVariables {
				deleteVariableURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/variables", data.ID.ValueString())
				deleteVariableReqBody := fmt.Sprintf(`{"key": "%s"}`, key)
				deleteVariableReq, err := http.NewRequest(http.MethodDelete, deleteVariableURL, strings.NewReader(deleteVariableReqBody))
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete variable, got error: %s", err))
					return
				}
				deleteVariableReq.Header.Set("Content-Type", "application/json")

				deleteVariableResp, err := r.client.Do(deleteVariableReq)
				if err != nil {
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete variable, got error: %s", err))
					return
				}
				defer deleteVariableResp.Body.Close()

				if deleteVariableResp.StatusCode != http.StatusOK {
					resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to delete variable, got status code: %d", deleteVariableResp.StatusCode))
					return
				}
			}
		}

		// Add new variables
		variablesMap := data.Variables.Elements()
		stateVariablesMap := make(map[string]attr.Value)
		for key, value := range variablesMap {
			variableURL := fmt.Sprintf("https://huggingface.co/api/spaces/%s/variables", data.ID.ValueString())
			variableReqBody := fmt.Sprintf(`{"key": "%s", "value": "%s"}`, key, value.(types.String).ValueString())
			variableResp, err := r.client.Post(variableURL, "application/json", strings.NewReader(variableReqBody))
			if err != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add variable, got error: %s", err))
				return
			}
			defer variableResp.Body.Close()

			if variableResp.StatusCode != http.StatusOK {
				resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to add variable, got status code: %d", variableResp.StatusCode))
				return
			}
			stateVariablesMap[key] = value
		}
		state.Variables, _ = types.MapValue(types.StringType, stateVariablesMap)

	}

	// Check if the space hardware needs to be updated
	if state.Hardware.ValueString() != data.Hardware.ValueString() {
		url := fmt.Sprintf("https://huggingface.co/api/spaces/%s/hardware", data.ID.ValueString())
		reqBody := fmt.Sprintf(`{"flavor": "%s"}`, data.Hardware.ValueString())
		httpResp, err := r.client.Post(url, "application/json", strings.NewReader(reqBody))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space hardware, got error: %s", err))
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(httpResp.Body)
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update space hardware, got status code: %d, response body: %s", httpResp.StatusCode, string(respBody)))
			return
		}

		var hardwareResp map[string]interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&hardwareResp)
		if err != nil {
			resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode update space hardware response, got error: %s", err))
			return
		}

		state.Hardware = data.Hardware
	}

	// Check if the space storage needs to be updated
	if state.Storage.ValueString() != data.Storage.ValueString() {
		url := fmt.Sprintf("https://huggingface.co/api/spaces/%s/storage", data.ID.ValueString())
		reqBody := fmt.Sprintf(`{"tier": "%s"}`, data.Storage.ValueString())
		httpResp, err := r.client.Post(url, "application/json", strings.NewReader(reqBody))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space storage, got error: %s", err))
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(httpResp.Body)
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update space storage, got status code: %d, response body: %s", httpResp.StatusCode, string(respBody)))
			return
		}

		var storageResp map[string]interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&storageResp)
		if err != nil {
			resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode update space storage response, got error: %s", err))
			return
		}

		state.Storage = data.Storage
	}

	// Check if the space sleep time needs to be updated
	if state.SleepTime.ValueInt64() != data.SleepTime.ValueInt64() {
		url := fmt.Sprintf("https://huggingface.co/api/spaces/%s/sleeptime", data.ID.ValueString())
		reqBody := fmt.Sprintf(`{"seconds": %d}`, data.SleepTime.ValueInt64())
		httpResp, err := r.client.Post(url, "application/json", strings.NewReader(reqBody))
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update space sleep time, got error: %s", err))
			return
		}
		defer httpResp.Body.Close()

		if httpResp.StatusCode != http.StatusOK {
			respBody, _ := ioutil.ReadAll(httpResp.Body)
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to update space sleep time, got status code: %d, response body: %s", httpResp.StatusCode, string(respBody)))
			return
		}

		var sleepTimeResp map[string]interface{}
		err = json.NewDecoder(httpResp.Body).Decode(&sleepTimeResp)
		if err != nil {
			resp.Diagnostics.AddError("JSON Decode Error", fmt.Sprintf("Unable to decode update space sleep time response, got error: %s", err))
			return
		}

		state.SleepTime = data.SleepTime
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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
