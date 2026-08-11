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
var _ datasource.DataSource = &SubscriptionPricePointDataSource{}

// NewSubscriptionPricePointDataSource creates a new subscription price point
// data source.
func NewSubscriptionPricePointDataSource() datasource.DataSource {
	return &SubscriptionPricePointDataSource{}
}

// SubscriptionPricePointDataSource defines the data source implementation.
type SubscriptionPricePointDataSource struct {
	client *Client
}

// SubscriptionPricePointDataSourceModel describes the data source data model.
type SubscriptionPricePointDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	SubscriptionID types.String `tfsdk:"subscription_id"`
	Territory      types.String `tfsdk:"territory"`
	CustomerPrice  types.String `tfsdk:"customer_price"`
	Proceeds       types.String `tfsdk:"proceeds"`
}

// SubscriptionPricePoint represents a subscription price point in the App Store
// Connect API.
type SubscriptionPricePoint struct {
	Type       string                           `json:"type"`
	ID         string                           `json:"id"`
	Attributes SubscriptionPricePointAttributes `json:"attributes"`
}

// SubscriptionPricePointAttributes represents the attributes of a subscription
// price point.
type SubscriptionPricePointAttributes struct {
	CustomerPrice string `json:"customerPrice,omitempty"`
	Proceeds      string `json:"proceeds,omitempty"`
}

func (d *SubscriptionPricePointDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_price_point"
}

func (d *SubscriptionPricePointDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Resolves a subscription price point ID from a customer price and territory. The resulting `id` is " +
			"referenced by the `appleappstoreconnect_subscription_price` resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The resolved price point ID.",
				Computed:            true,
			},
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription whose price points are searched. Price points are specific to each subscription.",
				Required:            true,
			},
			"territory": schema.StringAttribute{
				MarkdownDescription: "The territory code to look up the price in (e.g. `USA`).",
				Required:            true,
			},
			"customer_price": schema.StringAttribute{
				MarkdownDescription: "The customer price to match (e.g. `4.99`), as returned by the API for the territory.",
				Required:            true,
			},
			"proceeds": schema.StringAttribute{
				MarkdownDescription: "The proceeds (developer revenue) for the resolved price point.",
				Computed:            true,
			},
		},
	}
}

func (d *SubscriptionPricePointDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubscriptionPricePointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubscriptionPricePointDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Resolving subscription price point", map[string]any{
		"subscription_id": data.SubscriptionID.ValueString(),
		"territory":       data.Territory.ValueString(),
		"customer_price":  data.CustomerPrice.ValueString(),
	})

	elements, err := doPaginated(ctx, d.client, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s/pricePoints", data.SubscriptionID.ValueString()),
		Query: map[string]string{
			"filter[territory]": data.Territory.ValueString(),
			"limit":             "200",
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to list subscription price points, got error: %s", err),
		)
		return
	}

	want := data.CustomerPrice.ValueString()
	for _, element := range elements {
		var pp SubscriptionPricePoint
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
		fmt.Sprintf("No price point with customer price '%s' found in territory '%s' for subscription '%s'",
			want, data.Territory.ValueString(), data.SubscriptionID.ValueString()),
	)
}
