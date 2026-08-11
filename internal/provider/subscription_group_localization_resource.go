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
var _ resource.Resource = &SubscriptionGroupLocalizationResource{}
var _ resource.ResourceWithImportState = &SubscriptionGroupLocalizationResource{}

// NewSubscriptionGroupLocalizationResource creates a new subscription group
// localization resource.
func NewSubscriptionGroupLocalizationResource() resource.Resource {
	return &SubscriptionGroupLocalizationResource{}
}

// SubscriptionGroupLocalizationResource defines the resource implementation.
type SubscriptionGroupLocalizationResource struct {
	client *Client
}

// SubscriptionGroupLocalizationResourceModel describes the resource data model.
type SubscriptionGroupLocalizationResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SubscriptionGroupID types.String `tfsdk:"subscription_group_id"`
	Locale              types.String `tfsdk:"locale"`
	Name                types.String `tfsdk:"name"`
	CustomAppName       types.String `tfsdk:"custom_app_name"`
}

func (r *SubscriptionGroupLocalizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_group_localization"
}

func (r *SubscriptionGroupLocalizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a localized name for a subscription group in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the subscription group localization.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subscription_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription group this localization belongs to. Changing this forces a new resource.",
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
				MarkdownDescription: "The localized name of the subscription group.",
				Required:            true,
			},
			"custom_app_name": schema.StringAttribute{
				MarkdownDescription: "An optional custom app name displayed for this subscription group in this locale.",
				Optional:            true,
			},
		},
	}
}

func (r *SubscriptionGroupLocalizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionGroupLocalizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionGroupLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs := SubscriptionGroupLocalizationCreateRequestAttributes{
		Locale: data.Locale.ValueString(),
		Name:   data.Name.ValueString(),
	}
	if !data.CustomAppName.IsNull() {
		v := data.CustomAppName.ValueString()
		attrs.CustomAppName = &v
	}

	createReq := SubscriptionGroupLocalizationCreateRequest{
		Data: SubscriptionGroupLocalizationCreateRequestData{
			Type:       "subscriptionGroupLocalizations",
			Attributes: attrs,
			Relationships: SubscriptionGroupLocalizationCreateRequestRelationships{
				SubscriptionGroup: RelationshipOne{
					Data: RelationshipData{Type: "subscriptionGroups", ID: data.SubscriptionGroupID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating subscription group localization", map[string]any{
		"subscription_group_id": data.SubscriptionGroupID.ValueString(),
		"locale":                data.Locale.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/subscriptionGroupLocalizations",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create subscription group localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionGroupLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group localization response, got error: %s", err),
		)
		return
	}

	if loc.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created subscription group localization",
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupLocalizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionGroupLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroupLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read subscription group localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionGroupLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupLocalizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubscriptionGroupLocalizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	attrs := SubscriptionGroupLocalizationUpdateRequestAttributes{
		Name: &name,
	}
	if !data.CustomAppName.IsNull() {
		v := data.CustomAppName.ValueString()
		attrs.CustomAppName = &v
	}

	updateReq := SubscriptionGroupLocalizationUpdateRequest{
		Data: SubscriptionGroupLocalizationUpdateRequestData{
			Type:       "subscriptionGroupLocalizations",
			ID:         data.ID.ValueString(),
			Attributes: attrs,
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroupLocalizations/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update subscription group localization, got error: %s", err),
		)
		return
	}

	var loc SubscriptionGroupLocalization
	if err := json.Unmarshal(apiResp.Data, &loc); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group localization response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &loc)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupLocalizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionGroupLocalizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroupLocalizations/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete subscription group localization, got error: %s", err),
		)
		return
	}
}

func (r *SubscriptionGroupLocalizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the localization data returned by
// the API.
func (r *SubscriptionGroupLocalizationResource) updateModel(model *SubscriptionGroupLocalizationResourceModel, loc *SubscriptionGroupLocalization) {
	model.ID = types.StringValue(loc.ID)
	model.Locale = types.StringValue(loc.Attributes.Locale)
	model.Name = types.StringValue(loc.Attributes.Name)

	if loc.Attributes.CustomAppName != "" {
		model.CustomAppName = types.StringValue(loc.Attributes.CustomAppName)
	}

	if loc.Relationships != nil && loc.Relationships.SubscriptionGroup != nil && loc.Relationships.SubscriptionGroup.Data != nil {
		model.SubscriptionGroupID = types.StringValue(loc.Relationships.SubscriptionGroup.Data.ID)
	}
}
