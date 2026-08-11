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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InAppPurchasePriceScheduleResource{}
var _ resource.ResourceWithImportState = &InAppPurchasePriceScheduleResource{}

// NewInAppPurchasePriceScheduleResource creates a new In-App Purchase price
// schedule resource.
func NewInAppPurchasePriceScheduleResource() resource.Resource {
	return &InAppPurchasePriceScheduleResource{}
}

// InAppPurchasePriceScheduleResource defines the resource implementation.
type InAppPurchasePriceScheduleResource struct {
	client *Client
}

// InAppPurchasePriceScheduleResourceModel describes the resource data model.
type InAppPurchasePriceScheduleResourceModel struct {
	ID              types.String          `tfsdk:"id"`
	InAppPurchaseID types.String          `tfsdk:"in_app_purchase_id"`
	BaseTerritory   types.String          `tfsdk:"base_territory"`
	ManualPrices    []IAPManualPriceModel `tfsdk:"manual_prices"`
}

// IAPManualPriceModel describes a single territory price within the schedule.
type IAPManualPriceModel struct {
	PricePointID types.String `tfsdk:"price_point_id"`
	Territory    types.String `tfsdk:"territory"`
	StartDate    types.String `tfsdk:"start_date"`
}

func (r *InAppPurchasePriceScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase_price_schedule"
}

func (r *InAppPurchasePriceScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the price schedule of an In-App Purchase. An In-App Purchase has a single price schedule; " +
			"App Store Connect derives prices for other territories from the base territory, and any explicit per-territory " +
			"prices are provided via `manual_prices`. Price point IDs are resolved with the " +
			"`appleappstoreconnect_in_app_purchase_price_point` data source. Changing any attribute replaces the schedule.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the price schedule.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"in_app_purchase_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the In-App Purchase this schedule prices. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"base_territory": schema.StringAttribute{
				MarkdownDescription: "The base territory code (e.g. `USA`) Apple uses to derive prices in other territories. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"manual_prices": schema.SetNestedAttribute{
				MarkdownDescription: "The explicit per-territory prices. At minimum the base territory must be included. Changing the set forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"price_point_id": schema.StringAttribute{
							MarkdownDescription: "The price point ID for this territory (resolve it with the `appleappstoreconnect_in_app_purchase_price_point` data source).",
							Required:            true,
						},
						"territory": schema.StringAttribute{
							MarkdownDescription: "The territory code this price applies to (e.g. `USA`).",
							Required:            true,
						},
						"start_date": schema.StringAttribute{
							MarkdownDescription: "The date the price takes effect (ISO 8601, e.g. `2025-06-01`). Omit for immediate effect.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

func (r *InAppPurchasePriceScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InAppPurchasePriceScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InAppPurchasePriceScheduleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prices := make([]manualPrice, 0, len(data.ManualPrices))
	for _, p := range data.ManualPrices {
		prices = append(prices, manualPrice{
			PricePointID: p.PricePointID.ValueString(),
			Territory:    p.Territory.ValueString(),
			StartDate:    p.StartDate.ValueString(),
		})
	}

	createReq := buildIAPPriceScheduleCreateRequest(
		data.InAppPurchaseID.ValueString(),
		data.BaseTerritory.ValueString(),
		prices,
	)

	tflog.Debug(ctx, "Creating In-App Purchase price schedule", map[string]any{
		"in_app_purchase_id": data.InAppPurchaseID.ValueString(),
		"base_territory":     data.BaseTerritory.ValueString(),
		"manual_prices":      len(prices),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/inAppPurchasePriceSchedules",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create In-App Purchase price schedule, got error: %s", err),
		)
		return
	}

	var schedule IAPPriceSchedule
	if err := json.Unmarshal(apiResp.Data, &schedule); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse price schedule response, got error: %s", err),
		)
		return
	}

	if schedule.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created price schedule",
		)
		return
	}

	data.ID = types.StringValue(schedule.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchasePriceScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InAppPurchasePriceScheduleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the schedule still exists. The schedule is replace-only, so the
	// configured base_territory/manual_prices are retained from state rather
	// than reconciled per-price (Apple also derives additional territories).
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/inAppPurchasePriceSchedules/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read In-App Purchase price schedule, got error: %s", err),
		)
		return
	}

	var schedule IAPPriceSchedule
	if err := json.Unmarshal(apiResp.Data, &schedule); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse price schedule response, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(schedule.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchasePriceScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace, so Update is never reached with a
	// real change. Implemented to satisfy the interface.
	var data InAppPurchasePriceScheduleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchasePriceScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InAppPurchasePriceScheduleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// App Store Connect has no endpoint to delete a price schedule; a new
	// schedule replaces the previous one. Removing the resource only drops it
	// from Terraform state and leaves the current pricing in place.
	resp.Diagnostics.AddWarning(
		"Price Schedule Not Removed",
		"The price schedule has been removed from Terraform state, but App Store Connect does not support deleting a price "+
			"schedule. The In-App Purchase keeps its current pricing until a new schedule is applied.",
	)
}

func (r *InAppPurchasePriceScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
