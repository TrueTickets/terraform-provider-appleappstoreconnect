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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SubscriptionResource{}
var _ resource.ResourceWithImportState = &SubscriptionResource{}

// NewSubscriptionResource creates a new subscription resource.
func NewSubscriptionResource() resource.Resource {
	return &SubscriptionResource{}
}

// SubscriptionResource defines the resource implementation.
type SubscriptionResource struct {
	client *Client
}

// SubscriptionResourceModel describes the resource data model.
type SubscriptionResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SubscriptionGroupID types.String `tfsdk:"subscription_group_id"`
	ProductID           types.String `tfsdk:"product_id"`
	Name                types.String `tfsdk:"name"`
	SubscriptionPeriod  types.String `tfsdk:"subscription_period"`
	FamilySharable      types.Bool   `tfsdk:"family_sharable"`
	GroupLevel          types.Int64  `tfsdk:"group_level"`
	ReviewNote          types.String `tfsdk:"review_note"`
	State               types.String `tfsdk:"state"`
}

func (r *SubscriptionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (r *SubscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an auto-renewable subscription within a subscription group in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the subscription.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subscription_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription group this subscription belongs to. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The unique product ID of the subscription (e.g., `com.example.app.premium.monthly`). Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The reference name of the subscription, used in App Store Connect and Sales and Trends reports. This can be updated in place.",
				Required:            true,
			},
			"subscription_period": schema.StringAttribute{
				MarkdownDescription: "The duration of a single subscription period. One of `ONE_WEEK`, `ONE_MONTH`, `TWO_MONTHS`, `THREE_MONTHS`, `SIX_MONTHS`, or `ONE_YEAR`. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						SubscriptionPeriodOneWeek,
						SubscriptionPeriodOneMonth,
						SubscriptionPeriodTwoMonths,
						SubscriptionPeriodThreeMonths,
						SubscriptionPeriodSixMonths,
						SubscriptionPeriodOneYear,
					),
				},
			},
			"family_sharable": schema.BoolAttribute{
				MarkdownDescription: "Whether the subscription is available through Family Sharing. Changing this forces a new resource.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"group_level": schema.Int64Attribute{
				MarkdownDescription: "The ranking of the subscription within its group (1 is the highest level/most service). Changing this forces a new resource.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"review_note": schema.StringAttribute{
				MarkdownDescription: "A note for the App Review team (max 4000 characters).",
				Optional:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the subscription.",
				Computed:            true,
			},
		},
	}
}

func (r *SubscriptionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs := SubscriptionCreateRequestAttributes{
		Name:               data.Name.ValueString(),
		ProductID:          data.ProductID.ValueString(),
		SubscriptionPeriod: data.SubscriptionPeriod.ValueString(),
	}
	if !data.FamilySharable.IsNull() && !data.FamilySharable.IsUnknown() {
		v := data.FamilySharable.ValueBool()
		attrs.FamilySharable = &v
	}
	if !data.GroupLevel.IsNull() && !data.GroupLevel.IsUnknown() {
		v := data.GroupLevel.ValueInt64()
		attrs.GroupLevel = &v
	}
	if !data.ReviewNote.IsNull() {
		v := data.ReviewNote.ValueString()
		attrs.ReviewNote = &v
	}

	createReq := SubscriptionCreateRequest{
		Data: SubscriptionCreateRequestData{
			Type:       "subscriptions",
			Attributes: attrs,
			Relationships: SubscriptionCreateRequestRelationships{
				Group: RelationshipOne{
					Data: RelationshipData{Type: "subscriptionGroups", ID: data.SubscriptionGroupID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating subscription", map[string]any{
		"product_id":            data.ProductID.ValueString(),
		"subscription_group_id": data.SubscriptionGroupID.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/subscriptions",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create subscription, got error: %s", err),
		)
		return
	}

	var subscription Subscription
	if err := json.Unmarshal(apiResp.Data, &subscription); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription response, got error: %s", err),
		)
		return
	}

	if subscription.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created subscription",
		)
		return
	}

	r.updateModel(&data, &subscription)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s", data.ID.ValueString()),
		Query: map[string]string{
			"include": "group",
		},
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read subscription, got error: %s", err),
		)
		return
	}

	var subscription Subscription
	if err := json.Unmarshal(apiResp.Data, &subscription); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &subscription)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubscriptionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	attrs := SubscriptionUpdateRequestAttributes{
		Name: &name,
	}
	if !data.ReviewNote.IsNull() {
		v := data.ReviewNote.ValueString()
		attrs.ReviewNote = &v
	}

	updateReq := SubscriptionUpdateRequest{
		Data: SubscriptionUpdateRequestData{
			Type:       "subscriptions",
			ID:         data.ID.ValueString(),
			Attributes: attrs,
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update subscription, got error: %s", err),
		)
		return
	}

	var subscription Subscription
	if err := json.Unmarshal(apiResp.Data, &subscription); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &subscription)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete subscription, got error: %s", err),
		)
		return
	}
}

func (r *SubscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the subscription data returned by
// the API.
func (r *SubscriptionResource) updateModel(model *SubscriptionResourceModel, subscription *Subscription) {
	model.ID = types.StringValue(subscription.ID)
	model.ProductID = types.StringValue(subscription.Attributes.ProductID)
	model.Name = types.StringValue(subscription.Attributes.Name)
	model.SubscriptionPeriod = types.StringValue(subscription.Attributes.SubscriptionPeriod)
	model.FamilySharable = types.BoolValue(subscription.Attributes.FamilySharable)
	model.GroupLevel = types.Int64Value(subscription.Attributes.GroupLevel)
	model.State = types.StringValue(subscription.Attributes.State)

	if subscription.Attributes.ReviewNote != "" {
		model.ReviewNote = types.StringValue(subscription.Attributes.ReviewNote)
	}

	if subscription.Relationships != nil && subscription.Relationships.Group != nil && subscription.Relationships.Group.Data != nil {
		model.SubscriptionGroupID = types.StringValue(subscription.Relationships.Group.Data.ID)
	}
}
