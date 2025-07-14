// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PassTypeIDResource{}
var _ resource.ResourceWithImportState = &PassTypeIDResource{}

// NewPassTypeIDResource creates a new Pass Type ID resource.
func NewPassTypeIDResource() resource.Resource {
	return &PassTypeIDResource{}
}

// PassTypeIDResource defines the resource implementation.
type PassTypeIDResource struct {
	client *Client
}

// PassTypeIDResourceModel describes the resource data model.
type PassTypeIDResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Identifier  types.String `tfsdk:"identifier"`
	Description types.String `tfsdk:"description"`
	CreatedDate types.String `tfsdk:"created_date"`
}

func (r *PassTypeIDResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pass_type_id"
}

func (r *PassTypeIDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Pass Type ID in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the Pass Type ID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier for the Pass Type ID (e.g., 'pass.io.truetickets.test.membership'). This must be unique and follow reverse-DNS format.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of the Pass Type ID.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_date": schema.StringAttribute{
				MarkdownDescription: "The date when the Pass Type ID was created.",
				Computed:            true,
			},
		},
	}
}

func (r *PassTypeIDResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PassTypeIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PassTypeIDResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate identifier format
	if !isValidPassTypeIdentifier(data.Identifier.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("identifier"),
			"Invalid Pass Type Identifier",
			"The identifier must follow reverse-DNS format (e.g., 'pass.io.truetickets.test.membership').",
		)
		return
	}

	// Create the request
	createReq := PassTypeIDCreateRequest{
		Data: PassTypeIDCreateRequestData{
			Type: "passTypeIds",
			Attributes: PassTypeIDCreateRequestAttributes{
				Identifier: data.Identifier.ValueString(),
				Name:       data.Description.ValueString(),
			},
		},
	}

	tflog.Debug(ctx, "Creating Pass Type ID", map[string]interface{}{
		"identifier":  data.Identifier.ValueString(),
		"description": data.Description.ValueString(),
	})

	// Log the request body for debugging
	requestBody, _ := json.Marshal(createReq)
	tflog.Debug(ctx, "Request body", map[string]interface{}{
		"request_body": string(requestBody),
	})

	// Make the API request
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/passTypeIds",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create Pass Type ID, got error: %s", err),
		)
		return
	}

	// Parse the response
	var passTypeID PassTypeID
	if err := json.Unmarshal(apiResp.Data, &passTypeID); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Pass Type ID response, got error: %s", err),
		)
		return
	}

	// Log the raw response for debugging
	tflog.Debug(ctx, "Raw API response", map[string]interface{}{
		"raw_response": string(apiResp.Data),
	})

	// Update the model with the response data
	tflog.Debug(ctx, "Pass Type ID create response", map[string]interface{}{
		"response_id": passTypeID.ID,
		"identifier":  passTypeID.Attributes.Identifier,
		"name":        passTypeID.Attributes.Name,
	})

	// Validate that we got an ID from the API
	if passTypeID.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created Pass Type ID",
		)
		return
	}

	data.ID = types.StringValue(passTypeID.ID)
	if passTypeID.Attributes.CreatedDate != nil {
		data.CreatedDate = types.StringValue(passTypeID.Attributes.CreatedDate.Format("2006-01-02T15:04:05Z"))
	} else {
		// Set to null if not provided by API
		data.CreatedDate = types.StringNull()
	}

	tflog.Trace(ctx, "Created Pass Type ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PassTypeIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PassTypeIDResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Pass Type ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Validate that we have a valid ID
	if data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource State",
			"The Pass Type ID resource does not have a valid ID",
		)
		return
	}

	// Make the API request
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/passTypeIds/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read Pass Type ID, got error: %s", err),
		)
		return
	}

	// Parse the response
	var passTypeID PassTypeID
	if err := json.Unmarshal(apiResp.Data, &passTypeID); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Pass Type ID response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data
	data.Identifier = types.StringValue(passTypeID.Attributes.Identifier)
	data.Description = types.StringValue(passTypeID.Attributes.Name)
	if passTypeID.Attributes.CreatedDate != nil {
		data.CreatedDate = types.StringValue(passTypeID.Attributes.CreatedDate.Format("2006-01-02T15:04:05Z"))
	} else {
		// Set to null if not provided by API
		data.CreatedDate = types.StringNull()
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PassTypeIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PassTypeIDResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Note: The API might not support updating Pass Type IDs
	// If it doesn't, we should add a diagnostic error here
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Pass Type IDs cannot be updated. To change the identifier, you must delete and recreate the resource.",
	)
}

func (r *PassTypeIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PassTypeIDResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Pass Type ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Validate that we have a valid ID
	if data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource State",
			"The Pass Type ID resource does not have a valid ID",
		)
		return
	}

	// Make the API request
	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/passTypeIds/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete Pass Type ID, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "Deleted Pass Type ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *PassTypeIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// isValidPassTypeIdentifier validates that the identifier follows reverse-DNS format.
func isValidPassTypeIdentifier(identifier string) bool {
	// Pattern for reverse-DNS format starting with "pass."
	// Each segment can contain alphanumeric characters and hyphens, but cannot start or end with a hyphen
	pattern := `^pass\.([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+(\.([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?))+$`
	matched, _ := regexp.MatchString(pattern, identifier)
	return matched
}
