// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// bundleIDPlatforms are the platform values accepted when creating a Bundle ID.
var bundleIDPlatforms = []string{"IOS", "MAC_OS", "UNIVERSAL"}

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BundleIDResource{}
var _ resource.ResourceWithImportState = &BundleIDResource{}

// NewBundleIDResource creates a new Bundle ID resource.
func NewBundleIDResource() resource.Resource {
	return &BundleIDResource{}
}

// BundleIDResource defines the resource implementation.
type BundleIDResource struct {
	client *Client
}

// BundleIDResourceModel describes the resource data model.
type BundleIDResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Platform   types.String `tfsdk:"platform"`
	SeedID     types.String `tfsdk:"seed_id"`
}

func (r *BundleIDResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundle_id"
}

func (r *BundleIDResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Bundle ID (App ID) in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (resource ID) of the Bundle ID in App Store Connect.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The reverse-DNS bundle identifier (e.g., 'com.example.app'). This must be unique and cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Bundle ID. This is the only attribute that can be modified in place.",
				Required:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform of the Bundle ID. One of `IOS`, `MAC_OS`, or `UNIVERSAL`. Cannot be changed after creation.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(bundleIDPlatforms...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"seed_id": schema.StringAttribute{
				MarkdownDescription: "The seed ID (team ID prefix) of the Bundle ID. If omitted, App Store Connect assigns it automatically. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

func (r *BundleIDResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BundleIDResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BundleIDResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate identifier format.
	if !isValidBundleIdentifier(data.Identifier.ValueString()) {
		resp.Diagnostics.AddAttributeError(
			path.Root("identifier"),
			"Invalid Bundle Identifier",
			"The identifier must follow reverse-DNS format (e.g., 'com.example.app'), optionally ending with a '.*' wildcard segment.",
		)
		return
	}

	// Build the request. The seed ID is optional and only sent when the user
	// explicitly configured it.
	createReq := BundleIDCreateRequest{
		Data: BundleIDCreateRequestData{
			Type: "bundleIds",
			Attributes: BundleIDCreateRequestAttributes{
				Identifier: data.Identifier.ValueString(),
				Name:       data.Name.ValueString(),
				Platform:   data.Platform.ValueString(),
			},
		},
	}
	if !data.SeedID.IsNull() && !data.SeedID.IsUnknown() {
		createReq.Data.Attributes.SeedID = data.SeedID.ValueString()
	}

	tflog.Debug(ctx, "Creating Bundle ID", map[string]interface{}{
		"identifier": data.Identifier.ValueString(),
		"name":       data.Name.ValueString(),
		"platform":   data.Platform.ValueString(),
	})

	// Make the API request.
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/bundleIds",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create Bundle ID, got error: %s", err),
		)
		return
	}

	// Parse the response.
	var bundleID BundleID
	if err := json.Unmarshal(apiResp.Data, &bundleID); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Bundle ID response, got error: %s", err),
		)
		return
	}

	// Validate that we got an ID from the API.
	if bundleID.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created Bundle ID",
		)
		return
	}

	// Update the model with the response data.
	r.updateModel(&data, &bundleID)

	tflog.Trace(ctx, "Created Bundle ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BundleIDResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BundleIDResourceModel

	// Read Terraform prior state data into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Bundle ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Validate that we have a valid ID.
	if data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource State",
			"The Bundle ID resource does not have a valid ID",
		)
		return
	}

	// Make the API request.
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/bundleIds/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read Bundle ID, got error: %s", err),
		)
		return
	}

	// Parse the response.
	var bundleID BundleID
	if err := json.Unmarshal(apiResp.Data, &bundleID); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Bundle ID response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data.
	r.updateModel(&data, &bundleID)

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BundleIDResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BundleIDResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Only the name can be modified; identifier, platform and seed_id force
	// replacement via the schema plan modifiers.
	updateReq := BundleIDUpdateRequest{
		Data: BundleIDUpdateRequestData{
			Type: "bundleIds",
			ID:   data.ID.ValueString(),
			Attributes: BundleIDUpdateRequestAttributes{
				Name: data.Name.ValueString(),
			},
		},
	}

	tflog.Debug(ctx, "Updating Bundle ID", map[string]interface{}{
		"id":   data.ID.ValueString(),
		"name": data.Name.ValueString(),
	})

	// Make the API request.
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/bundleIds/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update Bundle ID, got error: %s", err),
		)
		return
	}

	// Parse the response.
	var bundleID BundleID
	if err := json.Unmarshal(apiResp.Data, &bundleID); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Bundle ID response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data.
	r.updateModel(&data, &bundleID)

	tflog.Trace(ctx, "Updated Bundle ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save updated data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BundleIDResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BundleIDResourceModel

	// Read Terraform prior state data into the model.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting Bundle ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Validate that we have a valid ID.
	if data.ID.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource State",
			"The Bundle ID resource does not have a valid ID",
		)
		return
	}

	// Make the API request.
	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/bundleIds/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete Bundle ID, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "Deleted Bundle ID", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *BundleIDResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the Bundle ID data returned by
// the API.
func (r *BundleIDResource) updateModel(model *BundleIDResourceModel, bundleID *BundleID) {
	model.ID = types.StringValue(bundleID.ID)
	model.Identifier = types.StringValue(bundleID.Attributes.Identifier)
	model.Name = types.StringValue(bundleID.Attributes.Name)
	model.Platform = types.StringValue(bundleID.Attributes.Platform)
	if bundleID.Attributes.SeedID != "" {
		model.SeedID = types.StringValue(bundleID.Attributes.SeedID)
	} else {
		model.SeedID = types.StringNull()
	}
}

// isValidBundleIdentifier validates that the identifier follows reverse-DNS
// format, optionally ending with a '.*' wildcard segment (e.g. 'com.example.*').
// The App Store Connect API remains authoritative; this is a fast client-side
// sanity check.
func isValidBundleIdentifier(identifier string) bool {
	// At least two alphanumeric (and hyphen) segments separated by dots, with
	// an optional trailing wildcard segment. Segments cannot start or end with
	// a hyphen.
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)+(\.\*)?$`
	matched, _ := regexp.MatchString(pattern, identifier)
	return matched
}
