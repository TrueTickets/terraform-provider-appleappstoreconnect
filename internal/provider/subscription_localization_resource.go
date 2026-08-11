// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SubscriptionLocalizationResource{}
var _ resource.ResourceWithImportState = &SubscriptionLocalizationResource{}

// NewSubscriptionLocalizationResource creates a new subscription localization
// resource.
func NewSubscriptionLocalizationResource() resource.Resource {
	return &SubscriptionLocalizationResource{}
}

// SubscriptionLocalizationResource defines the resource implementation.
type SubscriptionLocalizationResource struct {
	client *Client
}

// SubscriptionLocalizationResourceModel describes the resource data model.
type SubscriptionLocalizationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	Locale         types.String `tfsdk:"locale"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

func (r *SubscriptionLocalizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_localization"
}

func (r *SubscriptionLocalizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a localized name and description for an auto-renewable subscription in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the subscription localization.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription this localization belongs to. Changing this forces a new resource.",
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
				MarkdownDescription: "The localized display name of the subscription.",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The localized description of the subscription.",
				Optional:            true,
			},
		},
	}
}

func (r *SubscriptionLocalizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionLocalizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs := SubscriptionLocalizationCreateRequestAttributes{
		Locale: data.Locale.ValueString(),
		Name:   data.Name.ValueString(),
	}
	if !data.Description.IsNull() {
		v := data.Description.ValueString()
		attrs.Description = &v
	}

	createReq := SubscriptionLocalizationCreateRequest{
		Data: SubscriptionLocalizationCreateRequestData{
			Type:       "subscriptionLocalizations",
			Attributes: attrs,
			Relationships: SubscriptionLocalizationCreateRequestRelationships{
				Subscription: RelationshipOne{
					Data: RelationshipData{Type: "subscriptions", ID: data.SubscriptionID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating subscription localization", map[string]any{
		"subscription_id": data.SubscriptionID.ValueString(),
		"locale":          data.Locale.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/subscriptionLocalizations",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create subscription localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription localization response, got error: %s", err),
		)
		return
	}

	if loc.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created subscription localization",
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionLocalizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptionLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read subscription localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionLocalizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubscriptionLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	attrs := SubscriptionLocalizationUpdateRequestAttributes{
		Name: &name,
	}
	if !data.Description.IsNull() {
		v := data.Description.ValueString()
		attrs.Description = &v
	}

	updateReq := SubscriptionLocalizationUpdateRequest{
		Data: SubscriptionLocalizationUpdateRequestData{
			Type:       "subscriptionLocalizations",
			ID:         data.ID.ValueString(),
			Attributes: attrs,
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v1/subscriptionLocalizations/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update subscription localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionLocalizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/subscriptionLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete subscription localization, got error: %s", err),
		)
		return
	}
}

func (r *SubscriptionLocalizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the localization data returned by
// the API.
func (r *SubscriptionLocalizationResource) updateModel(model *SubscriptionLocalizationResourceModel, loc *SubscriptionLocalization) {
	model.ID = types.StringValue(loc.ID)
	model.Locale = types.StringValue(loc.Attributes.Locale)
	model.Name = types.StringValue(loc.Attributes.Name)

	if loc.Attributes.Description != "" {
		model.Description = types.StringValue(loc.Attributes.Description)
	}

	if loc.Relationships != nil && loc.Relationships.Subscription != nil && loc.Relationships.Subscription.Data != nil {
		model.SubscriptionID = types.StringValue(loc.Relationships.Subscription.Data.ID)
	}
}
