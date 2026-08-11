// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InAppPurchaseLocalizationResource{}
var _ resource.ResourceWithImportState = &InAppPurchaseLocalizationResource{}

// NewInAppPurchaseLocalizationResource creates a new In-App Purchase
// localization resource.
func NewInAppPurchaseLocalizationResource() resource.Resource {
	return &InAppPurchaseLocalizationResource{}
}

// InAppPurchaseLocalizationResource defines the resource implementation.
type InAppPurchaseLocalizationResource struct {
	client *Client
}

// InAppPurchaseLocalizationResourceModel describes the resource data model.
type InAppPurchaseLocalizationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	InAppPurchaseID types.String `tfsdk:"in_app_purchase_id"`
	Locale          types.String `tfsdk:"locale"`
	Name            types.String `tfsdk:"name"`
	Description     types.String `tfsdk:"description"`
}

func (r *InAppPurchaseLocalizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase_localization"
}

func (r *InAppPurchaseLocalizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a localized name and description for an In-App Purchase in App Store Connect. " +
			"At least one localization (typically the app's primary locale) is required before an In-App Purchase can be submitted for review.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the In-App Purchase localization.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"in_app_purchase_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the In-App Purchase this localization belongs to. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "The locale of the localization (e.g., `en-US`, `fr-FR`). Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The localized display name of the In-App Purchase (max 30 characters).",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(30),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The localized description of the In-App Purchase (max 4000 characters).",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.LengthAtMost(4000),
				},
			},
		},
	}
}

func (r *InAppPurchaseLocalizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InAppPurchaseLocalizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InAppPurchaseLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs := InAppPurchaseLocalizationCreateRequestAttributes{
		Locale: data.Locale.ValueString(),
		Name:   data.Name.ValueString(),
	}
	if !data.Description.IsNull() {
		v := data.Description.ValueString()
		attrs.Description = &v
	}

	createReq := InAppPurchaseLocalizationCreateRequest{
		Data: InAppPurchaseLocalizationCreateRequestData{
			Type:       "inAppPurchaseLocalizations",
			Attributes: attrs,
			Relationships: InAppPurchaseLocalizationCreateRequestRelationships{
				InAppPurchase: RelationshipOne{
					Data: RelationshipData{
						Type: "inAppPurchases",
						ID:   data.InAppPurchaseID.ValueString(),
					},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating In-App Purchase localization", map[string]any{
		"in_app_purchase_id": data.InAppPurchaseID.ValueString(),
		"locale":             data.Locale.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/inAppPurchaseLocalizations",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create In-App Purchase localization, got error: %s", err),
		)
		return
	}

	var loc InAppPurchaseLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase localization response, got error: %s", err),
		)
		return
	}

	if loc.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created In-App Purchase localization",
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseLocalizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InAppPurchaseLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/inAppPurchaseLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read In-App Purchase localization, got error: %s", err),
		)
		return
	}

	var loc InAppPurchaseLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseLocalizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InAppPurchaseLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	attrs := InAppPurchaseLocalizationUpdateRequestAttributes{
		Name: &name,
	}
	if !data.Description.IsNull() {
		v := data.Description.ValueString()
		attrs.Description = &v
	}

	updateReq := InAppPurchaseLocalizationUpdateRequest{
		Data: InAppPurchaseLocalizationUpdateRequestData{
			Type:       "inAppPurchaseLocalizations",
			ID:         data.ID.ValueString(),
			Attributes: attrs,
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v1/inAppPurchaseLocalizations/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update In-App Purchase localization, got error: %s", err),
		)
		return
	}

	var loc InAppPurchaseLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseLocalizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InAppPurchaseLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/inAppPurchaseLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete In-App Purchase localization, got error: %s", err),
		)
		return
	}
}

func (r *InAppPurchaseLocalizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the localization data returned
// by the API.
func (r *InAppPurchaseLocalizationResource) updateModel(model *InAppPurchaseLocalizationResourceModel, loc *InAppPurchaseLocalization) {
	model.ID = types.StringValue(loc.ID)
	model.Locale = types.StringValue(loc.Attributes.Locale)
	model.Name = types.StringValue(loc.Attributes.Name)

	if loc.Attributes.Description != "" {
		model.Description = types.StringValue(loc.Attributes.Description)
	}

	if loc.Relationships != nil && loc.Relationships.InAppPurchase != nil && loc.Relationships.InAppPurchase.Data != nil {
		model.InAppPurchaseID = types.StringValue(loc.Relationships.InAppPurchase.Data.ID)
	}
}
