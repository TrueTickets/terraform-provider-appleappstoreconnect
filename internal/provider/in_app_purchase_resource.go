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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InAppPurchaseResource{}
var _ resource.ResourceWithImportState = &InAppPurchaseResource{}

// NewInAppPurchaseResource creates a new In-App Purchase resource.
func NewInAppPurchaseResource() resource.Resource {
	return &InAppPurchaseResource{}
}

// InAppPurchaseResource defines the resource implementation.
type InAppPurchaseResource struct {
	client *Client
}

// InAppPurchaseResourceModel describes the resource data model.
type InAppPurchaseResourceModel struct {
	ID                types.String `tfsdk:"id"`
	AppID             types.String `tfsdk:"app_id"`
	ProductID         types.String `tfsdk:"product_id"`
	Name              types.String `tfsdk:"name"`
	InAppPurchaseType types.String `tfsdk:"in_app_purchase_type"`
	FamilySharable    types.Bool   `tfsdk:"family_sharable"`
	ReviewNote        types.String `tfsdk:"review_note"`
	State             types.String `tfsdk:"state"`
}

func (r *InAppPurchaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase"
}

func (r *InAppPurchaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an In-App Purchase (consumable, non-consumable, or non-renewing subscription) for an app in App Store Connect. " +
			"Auto-renewable subscriptions are managed with the `appleappstoreconnect_subscription` resource instead.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the In-App Purchase.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The App Store Connect ID of the app this In-App Purchase belongs to. Use the `appleappstoreconnect_app` data source to look it up. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The unique product ID of the In-App Purchase (e.g., `com.example.app.premium`). Must be unique within the app. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The reference name of the In-App Purchase, used in App Store Connect and Sales and Trends reports (max 64 characters). This can be updated in place.",
				Required:            true,
			},
			"in_app_purchase_type": schema.StringAttribute{
				MarkdownDescription: "The type of In-App Purchase. One of `CONSUMABLE`, `NON_CONSUMABLE`, or `NON_RENEWING_SUBSCRIPTION`. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						InAppPurchaseTypeConsumable,
						InAppPurchaseTypeNonConsumable,
						InAppPurchaseTypeNonRenewingSubscription,
					),
				},
			},
			"family_sharable": schema.BoolAttribute{
				MarkdownDescription: "Whether the In-App Purchase is available through Family Sharing. Defaults to the value assigned by App Store Connect.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"review_note": schema.StringAttribute{
				MarkdownDescription: "A note for the App Review team (max 4000 characters).",
				Optional:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the In-App Purchase (e.g., `MISSING_METADATA`, `READY_TO_SUBMIT`, `APPROVED`).",
				Computed:            true,
			},
		},
	}
}

func (r *InAppPurchaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InAppPurchaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InAppPurchaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	attrs := InAppPurchaseCreateRequestAttributes{
		Name:              data.Name.ValueString(),
		ProductID:         data.ProductID.ValueString(),
		InAppPurchaseType: data.InAppPurchaseType.ValueString(),
	}
	if !data.FamilySharable.IsNull() && !data.FamilySharable.IsUnknown() {
		v := data.FamilySharable.ValueBool()
		attrs.FamilySharable = &v
	}
	if !data.ReviewNote.IsNull() {
		v := data.ReviewNote.ValueString()
		attrs.ReviewNote = &v
	}

	createReq := InAppPurchaseCreateRequest{
		Data: InAppPurchaseCreateRequestData{
			Type:       "inAppPurchases",
			Attributes: attrs,
			Relationships: InAppPurchaseCreateRequestRelationships{
				App: RelationshipOne{
					Data: RelationshipData{
						Type: "apps",
						ID:   data.AppID.ValueString(),
					},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating In-App Purchase", map[string]any{
		"product_id": data.ProductID.ValueString(),
		"app_id":     data.AppID.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v2/inAppPurchases",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create In-App Purchase, got error: %s", err),
		)
		return
	}

	var iap InAppPurchaseV2
	if err := json.Unmarshal(apiResp.Data, &iap); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase response, got error: %s", err),
		)
		return
	}

	if iap.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created In-App Purchase",
		)
		return
	}

	r.updateModel(&data, &iap)

	tflog.Trace(ctx, "Created In-App Purchase", map[string]any{
		"id": data.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InAppPurchaseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v2/inAppPurchases/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read In-App Purchase, got error: %s", err),
		)
		return
	}

	var iap InAppPurchaseV2
	if err := json.Unmarshal(apiResp.Data, &iap); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &iap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data InAppPurchaseResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	attrs := InAppPurchaseUpdateRequestAttributes{
		Name: &name,
	}
	if !data.FamilySharable.IsNull() && !data.FamilySharable.IsUnknown() {
		v := data.FamilySharable.ValueBool()
		attrs.FamilySharable = &v
	}
	if !data.ReviewNote.IsNull() {
		v := data.ReviewNote.ValueString()
		attrs.ReviewNote = &v
	}

	updateReq := InAppPurchaseUpdateRequest{
		Data: InAppPurchaseUpdateRequestData{
			Type:       "inAppPurchases",
			ID:         data.ID.ValueString(),
			Attributes: attrs,
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v2/inAppPurchases/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update In-App Purchase, got error: %s", err),
		)
		return
	}

	var iap InAppPurchaseV2
	if err := json.Unmarshal(apiResp.Data, &iap); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse In-App Purchase response, got error: %s", err),
		)
		return
	}

	r.updateModel(&data, &iap)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InAppPurchaseResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v2/inAppPurchases/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete In-App Purchase, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "Deleted In-App Purchase", map[string]any{
		"id": data.ID.ValueString(),
	})
}

func (r *InAppPurchaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// updateModel updates the resource model with the In-App Purchase data
// returned by the API. The app relationship is only present when the response
// includes it (Read uses include=app); otherwise the existing app_id is kept.
func (r *InAppPurchaseResource) updateModel(model *InAppPurchaseResourceModel, iap *InAppPurchaseV2) {
	model.ID = types.StringValue(iap.ID)
	model.ProductID = types.StringValue(iap.Attributes.ProductID)
	model.Name = types.StringValue(iap.Attributes.Name)
	model.InAppPurchaseType = types.StringValue(iap.Attributes.InAppPurchaseType)
	model.FamilySharable = types.BoolValue(iap.Attributes.FamilySharable)
	model.State = types.StringValue(iap.Attributes.State)

	if iap.Attributes.ReviewNote != "" {
		model.ReviewNote = types.StringValue(iap.Attributes.ReviewNote)
	}

	if iap.Relationships != nil && iap.Relationships.App != nil && iap.Relationships.App.Data != nil {
		model.AppID = types.StringValue(iap.Relationships.App.Data.ID)
	}
}
