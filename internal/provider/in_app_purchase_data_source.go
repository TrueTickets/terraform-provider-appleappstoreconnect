// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InAppPurchaseDataSource{}

// NewInAppPurchaseDataSource creates a new In-App Purchase data source.
func NewInAppPurchaseDataSource() datasource.DataSource {
	return &InAppPurchaseDataSource{}
}

// InAppPurchaseDataSource defines the data source implementation.
type InAppPurchaseDataSource struct {
	client *Client
}

// InAppPurchaseDataSourceModel describes the data source data model.
type InAppPurchaseDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	AppID             types.String `tfsdk:"app_id"`
	ProductID         types.String `tfsdk:"product_id"`
	Name              types.String `tfsdk:"name"`
	InAppPurchaseType types.String `tfsdk:"in_app_purchase_type"`
	FamilySharable    types.Bool   `tfsdk:"family_sharable"`
	ReviewNote        types.String `tfsdk:"review_note"`
	State             types.String `tfsdk:"state"`
	Filter            types.Object `tfsdk:"filter"`
}

// InAppPurchaseFilterModel describes the filter criteria. Product IDs are only
// unique within an app, so both fields are required when filtering.
type InAppPurchaseFilterModel struct {
	AppID     types.String `tfsdk:"app_id"`
	ProductID types.String `tfsdk:"product_id"`
}

func (d *InAppPurchaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase"
}

func (d *InAppPurchaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing In-App Purchase, either by its ID or by app ID and product ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the In-App Purchase.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("filter"),
					),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The App Store Connect ID of the app the In-App Purchase belongs to.",
				Computed:            true,
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The product ID of the In-App Purchase.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The reference name of the In-App Purchase.",
				Computed:            true,
			},
			"in_app_purchase_type": schema.StringAttribute{
				MarkdownDescription: "The type of In-App Purchase.",
				Computed:            true,
			},
			"family_sharable": schema.BoolAttribute{
				MarkdownDescription: "Whether the In-App Purchase is available through Family Sharing.",
				Computed:            true,
			},
			"review_note": schema.StringAttribute{
				MarkdownDescription: "The note for the App Review team.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the In-App Purchase.",
				Computed:            true,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for finding an In-App Purchase.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: "The App Store Connect ID of the app to search within.",
						Required:            true,
					},
					"product_id": schema.StringAttribute{
						MarkdownDescription: "The product ID to search for.",
						Required:            true,
					},
				},
			},
		},
	}
}

func (d *InAppPurchaseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *InAppPurchaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InAppPurchaseDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching In-App Purchase by ID", map[string]any{
			"id": data.ID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v2/inAppPurchases/%s", data.ID.ValueString()),
		})
		if err != nil {
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

		d.updateModel(&data, &iap)
	} else {
		var filter InAppPurchaseFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "Fetching In-App Purchase by filter", map[string]any{
			"app_id":     filter.AppID.ValueString(),
			"product_id": filter.ProductID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v1/apps/%s/inAppPurchasesV2", filter.AppID.ValueString()),
			Query: map[string]string{
				"filter[productId]": filter.ProductID.ValueString(),
				"limit":             "200",
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list In-App Purchases, got error: %s", err),
			)
			return
		}

		var iaps []InAppPurchaseV2
		if err := json.Unmarshal(apiResp.Data, &iaps); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse In-App Purchases response, got error: %s", err),
			)
			return
		}

		if len(iaps) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				fmt.Sprintf("No In-App Purchase found with product ID '%s' in app '%s'", filter.ProductID.ValueString(), filter.AppID.ValueString()),
			)
			return
		}

		if len(iaps) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Multiple In-App Purchases found with product ID '%s' in app '%s'", filter.ProductID.ValueString(), filter.AppID.ValueString()),
			)
			return
		}

		d.updateModel(&data, &iaps[0])
		// The list endpoint is scoped to the app, so the app ID is known.
		data.AppID = filter.AppID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the In-App Purchase data.
func (d *InAppPurchaseDataSource) updateModel(model *InAppPurchaseDataSourceModel, iap *InAppPurchaseV2) {
	model.ID = types.StringValue(iap.ID)
	model.ProductID = types.StringValue(iap.Attributes.ProductID)
	model.Name = types.StringValue(iap.Attributes.Name)
	model.InAppPurchaseType = types.StringValue(iap.Attributes.InAppPurchaseType)
	model.FamilySharable = types.BoolValue(iap.Attributes.FamilySharable)
	model.ReviewNote = types.StringValue(iap.Attributes.ReviewNote)
	model.State = types.StringValue(iap.Attributes.State)

	if iap.Relationships != nil && iap.Relationships.App != nil && iap.Relationships.App.Data != nil {
		model.AppID = types.StringValue(iap.Relationships.App.Data.ID)
	}
}
