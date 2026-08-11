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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &InAppPurchaseAvailabilityResource{}
var _ resource.ResourceWithImportState = &InAppPurchaseAvailabilityResource{}

// NewInAppPurchaseAvailabilityResource creates a new In-App Purchase
// availability resource.
func NewInAppPurchaseAvailabilityResource() resource.Resource {
	return &InAppPurchaseAvailabilityResource{}
}

// InAppPurchaseAvailabilityResource defines the resource implementation.
type InAppPurchaseAvailabilityResource struct {
	client *Client
}

// InAppPurchaseAvailabilityResourceModel describes the resource data model.
type InAppPurchaseAvailabilityResourceModel struct {
	ID                        types.String `tfsdk:"id"`
	InAppPurchaseID           types.String `tfsdk:"in_app_purchase_id"`
	AvailableInNewTerritories types.Bool   `tfsdk:"available_in_new_territories"`
	AvailableTerritories      types.Set    `tfsdk:"available_territories"`
}

func (r *InAppPurchaseAvailabilityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase_availability"
}

func (r *InAppPurchaseAvailabilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the territory availability of an In-App Purchase. Changing any attribute replaces the availability configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the availability configuration.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"in_app_purchase_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the In-App Purchase. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"available_in_new_territories": schema.BoolAttribute{
				MarkdownDescription: "Whether the In-App Purchase is automatically made available in territories Apple adds in the future. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"available_territories": schema.SetAttribute{
				MarkdownDescription: "The set of territory codes (e.g. `USA`, `FRA`) where the In-App Purchase is available. Changing the set forces a new resource.",
				Required:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *InAppPurchaseAvailabilityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InAppPurchaseAvailabilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InAppPurchaseAvailabilityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var territories []string
	resp.Diagnostics.Append(data.AvailableTerritories.ElementsAs(ctx, &territories, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	territoryRefs := make([]RelationshipData, 0, len(territories))
	for _, t := range territories {
		territoryRefs = append(territoryRefs, RelationshipData{Type: "territories", ID: t})
	}

	createReq := IAPAvailabilityCreateRequest{
		Data: IAPAvailabilityCreateRequestData{
			Type: "inAppPurchaseAvailabilities",
			Attributes: IAPAvailabilityCreateAttributes{
				AvailableInNewTerritories: data.AvailableInNewTerritories.ValueBool(),
			},
			Relationships: IAPAvailabilityCreateRelationships{
				InAppPurchase: RelationshipOne{
					Data: RelationshipData{Type: "inAppPurchases", ID: data.InAppPurchaseID.ValueString()},
				},
				AvailableTerritories: RelationshipMany{Data: territoryRefs},
			},
		},
	}

	tflog.Debug(ctx, "Creating In-App Purchase availability", map[string]any{
		"in_app_purchase_id": data.InAppPurchaseID.ValueString(),
		"territories":        len(territoryRefs),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/inAppPurchaseAvailabilities",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create In-App Purchase availability, got error: %s", err),
		)
		return
	}

	var availability IAPAvailability
	if err := json.Unmarshal(apiResp.Data, &availability); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse availability response, got error: %s", err),
		)
		return
	}

	if availability.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created availability",
		)
		return
	}

	data.ID = types.StringValue(availability.ID)
	data.AvailableInNewTerritories = types.BoolValue(availability.Attributes.AvailableInNewTerritories)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseAvailabilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InAppPurchaseAvailabilityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/inAppPurchaseAvailabilities/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read In-App Purchase availability, got error: %s", err),
		)
		return
	}

	var availability IAPAvailability
	if err := json.Unmarshal(apiResp.Data, &availability); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse availability response, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(availability.ID)
	data.AvailableInNewTerritories = types.BoolValue(availability.Attributes.AvailableInNewTerritories)
	if availability.Relationships != nil && availability.Relationships.InAppPurchase != nil && availability.Relationships.InAppPurchase.Data != nil {
		data.InAppPurchaseID = types.StringValue(availability.Relationships.InAppPurchase.Data.ID)
	}

	// Fetch the (potentially paginated) list of available territories.
	elements, err := doPaginated(ctx, r.client, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/inAppPurchaseAvailabilities/%s/availableTerritories", data.ID.ValueString()),
		Query:    map[string]string{"limit": "200"},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read available territories, got error: %s", err),
		)
		return
	}

	territories := make([]string, 0, len(elements))
	for _, element := range elements {
		var territory Territory
		if err := json.Unmarshal(element, &territory); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse territory, got error: %s", err),
			)
			return
		}
		territories = append(territories, territory.ID)
	}

	territorySet, diags := types.SetValueFrom(ctx, types.StringType, territories)
	resp.Diagnostics.Append(diags...)
	data.AvailableTerritories = territorySet

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseAvailabilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace, so Update is never reached with a
	// real change. Implemented to satisfy the interface.
	var data InAppPurchaseAvailabilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseAvailabilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InAppPurchaseAvailabilityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// App Store Connect has no endpoint to delete an availability; it is
	// replaced by posting a new one. Removing the resource only drops it from
	// Terraform state.
	resp.Diagnostics.AddWarning(
		"Availability Not Removed",
		"The availability has been removed from Terraform state, but App Store Connect does not support deleting an "+
			"availability configuration. The In-App Purchase keeps its current territory availability until a new one is applied.",
	)
}

func (r *InAppPurchaseAvailabilityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
