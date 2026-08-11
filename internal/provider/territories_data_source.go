// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TerritoriesDataSource{}

// NewTerritoriesDataSource creates a new Territories data source.
func NewTerritoriesDataSource() datasource.DataSource {
	return &TerritoriesDataSource{}
}

// TerritoriesDataSource defines the data source implementation.
type TerritoriesDataSource struct {
	client *Client
}

// TerritoriesDataSourceModel describes the data source data model.
type TerritoriesDataSourceModel struct {
	Territories types.List `tfsdk:"territories"`
}

// TerritoryItemModel describes a territory in the list.
type TerritoryItemModel struct {
	ID       types.String `tfsdk:"id"`
	Currency types.String `tfsdk:"currency"`
}

func (d *TerritoriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_territories"
}

func (d *TerritoriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve the list of App Store territories. The territory `id` " +
			"(e.g. `USA`, `FRA`, `GBR`) is referenced by In-App Purchase and subscription pricing and availability " +
			"resources.",

		Attributes: map[string]schema.Attribute{
			"territories": schema.ListNestedAttribute{
				MarkdownDescription: "The list of available App Store territories.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The territory code (e.g. `USA`, `FRA`, `GBR`).",
							Computed:            true,
						},
						"currency": schema.StringAttribute{
							MarkdownDescription: "The ISO 4217 currency code used in the territory (e.g. `USD`).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *TerritoriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TerritoriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TerritoriesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Fetching all territories")

	elements, err := doPaginated(ctx, d.client, Request{
		Method:   http.MethodGet,
		Endpoint: "/v1/territories",
		Query:    map[string]string{"limit": "200"},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to list territories, got error: %s", err),
		)
		return
	}

	items := make([]TerritoryItemModel, 0, len(elements))
	for _, element := range elements {
		var territory Territory
		if err := json.Unmarshal(element, &territory); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse territory, got error: %s", err),
			)
			return
		}

		items = append(items, TerritoryItemModel{
			ID:       types.StringValue(territory.ID),
			Currency: types.StringValue(territory.Attributes.Currency),
		})
	}

	territoryList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":       types.StringType,
			"currency": types.StringType,
		},
	}, items)
	resp.Diagnostics.Append(diags...)
	data.Territories = territoryList

	tflog.Debug(ctx, "Found territories", map[string]any{
		"count": len(items),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
