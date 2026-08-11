// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InAppPurchasePricePointDataSource{}

// NewInAppPurchasePricePointDataSource creates a new In-App Purchase price
// point data source.
func NewInAppPurchasePricePointDataSource() datasource.DataSource {
	return &InAppPurchasePricePointDataSource{}
}

// InAppPurchasePricePointDataSource defines the data source implementation.
type InAppPurchasePricePointDataSource struct {
	client *Client
}

// InAppPurchasePricePointDataSourceModel describes the data source data model.
type InAppPurchasePricePointDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	InAppPurchaseID types.String `tfsdk:"in_app_purchase_id"`
	Territory       types.String `tfsdk:"territory"`
	CustomerPrice   types.String `tfsdk:"customer_price"`
	Proceeds        types.String `tfsdk:"proceeds"`
}

// IAPPricePoint represents an In-App Purchase price point in the App Store
// Connect API.
type IAPPricePoint struct {
	Type       string                  `json:"type"`
	ID         string                  `json:"id"`
	Attributes IAPPricePointAttributes `json:"attributes"`
}

// IAPPricePointAttributes represents the attributes of an In-App Purchase price
// point.
type IAPPricePointAttributes struct {
	CustomerPrice string `json:"customerPrice,omitempty"`
	Proceeds      string `json:"proceeds,omitempty"`
}

func (d *InAppPurchasePricePointDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase_price_point"
}

func (d *InAppPurchasePricePointDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resolves an In-App Purchase price point ID from a customer price and territory. " +
			"App Store Connect no longer uses price tiers; the resulting `id` is referenced by the " +
			"`appleappstoreconnect_in_app_purchase_price_schedule` resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resolved price point ID.",
				Computed:            true,
			},
			"in_app_purchase_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the In-App Purchase whose price points are searched. Price points are specific to each In-App Purchase.",
				Required:            true,
			},
			"territory": schema.StringAttribute{
				MarkdownDescription: "The territory code to look up the price in (e.g. `USA`).",
				Required:            true,
			},
			"customer_price": schema.StringAttribute{
				MarkdownDescription: "The customer price to match (e.g. `0.99`), as returned by the API for the territory.",
				Required:            true,
			},
			"proceeds": schema.StringAttribute{
				MarkdownDescription: "The proceeds (developer revenue) for the resolved price point.",
				Computed:            true,
			},
		},
	}
}

func (d *InAppPurchasePricePointDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *InAppPurchasePricePointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InAppPurchasePricePointDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Resolving In-App Purchase price point", map[string]any{
		"in_app_purchase_id": data.InAppPurchaseID.ValueString(),
		"territory":          data.Territory.ValueString(),
		"customer_price":     data.CustomerPrice.ValueString(),
	})

	elements, err := doPaginated(ctx, d.client, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v2/inAppPurchases/%s/pricePoints", data.InAppPurchaseID.ValueString()),
		Query: map[string]string{
			"filter[territory]": data.Territory.ValueString(),
			"limit":             "200",
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to list In-App Purchase price points, got error: %s", err),
		)
		return
	}

	want := data.CustomerPrice.ValueString()
	for _, element := range elements {
		var pp IAPPricePoint
		if err := json.Unmarshal(element, &pp); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse price point, got error: %s", err),
			)
			return
		}

		if pp.Attributes.CustomerPrice == want {
			data.ID = types.StringValue(pp.ID)
			data.Proceeds = types.StringValue(pp.Attributes.Proceeds)
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Not Found",
		fmt.Sprintf("No price point with customer price '%s' found in territory '%s' for In-App Purchase '%s'",
			want, data.Territory.ValueString(), data.InAppPurchaseID.ValueString()),
	)
}
