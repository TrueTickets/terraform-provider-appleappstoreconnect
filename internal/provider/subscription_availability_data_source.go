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
var _ datasource.DataSource = &SubscriptionAvailabilityDataSource{}

// NewSubscriptionAvailabilityDataSource creates a new subscription availability
// data source.
func NewSubscriptionAvailabilityDataSource() datasource.DataSource {
	return &SubscriptionAvailabilityDataSource{}
}

// SubscriptionAvailabilityDataSource defines the data source implementation.
type SubscriptionAvailabilityDataSource struct {
	client *Client
}

// SubscriptionAvailabilityDataSourceModel describes the data source data model.
type SubscriptionAvailabilityDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	SubscriptionID            types.String `tfsdk:"subscription_id"`
	AvailableInNewTerritories types.Bool   `tfsdk:"available_in_new_territories"`
	AvailableTerritories      types.Set    `tfsdk:"available_territories"`
}

// SubscriptionAvailability represents the territory availability of a
// subscription in the App Store Connect API.
type SubscriptionAvailability struct {
	Type       string                             `json:"type"`
	ID         string                             `json:"id"`
	Attributes SubscriptionAvailabilityAttributes `json:"attributes"`
}

// SubscriptionAvailabilityAttributes represents the attributes of a
// subscription availability.
type SubscriptionAvailabilityAttributes struct {
	AvailableInNewTerritories bool `json:"availableInNewTerritories"`
}

func (d *SubscriptionAvailabilityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_availability"
}

func (d *SubscriptionAvailabilityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads the territory availability of a subscription. The App Store Connect API does not currently expose a " +
			"public endpoint to modify subscription availability, so this is provided as a read-only data source.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the availability configuration.",
				Computed:            true,
			},
			"subscription_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription whose availability is read.",
				Required:            true,
			},
			"available_in_new_territories": schema.BoolAttribute{
				MarkdownDescription: "Whether the subscription is automatically made available in territories Apple adds in the future.",
				Computed:            true,
			},
			"available_territories": schema.SetAttribute{
				MarkdownDescription: "The set of territory codes where the subscription is available.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (d *SubscriptionAvailabilityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubscriptionAvailabilityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubscriptionAvailabilityDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading subscription availability", map[string]any{
		"subscription_id": data.SubscriptionID.ValueString(),
	})

	apiResp, err := d.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s/subscriptionAvailability", data.SubscriptionID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read subscription availability, got error: %s", err),
		)
		return
	}

	var availability SubscriptionAvailability
	if err := json.Unmarshal(apiResp.Data, &availability); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription availability response, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(availability.ID)
	data.AvailableInNewTerritories = types.BoolValue(availability.Attributes.AvailableInNewTerritories)

	// Fetch the (potentially paginated) list of available territories.
	elements, err := doPaginated(ctx, d.client, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptions/%s/availableTerritories", data.SubscriptionID.ValueString()),
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
