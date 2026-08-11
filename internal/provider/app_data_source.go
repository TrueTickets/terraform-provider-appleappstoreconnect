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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AppDataSource{}

// NewAppDataSource creates a new App data source.
func NewAppDataSource() datasource.DataSource {
	return &AppDataSource{}
}

// AppDataSource defines the data source implementation.
type AppDataSource struct {
	client *Client
}

// AppDataSourceModel describes the data source data model.
type AppDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	BundleID      types.String `tfsdk:"bundle_id"`
	Name          types.String `tfsdk:"name"`
	SKU           types.String `tfsdk:"sku"`
	PrimaryLocale types.String `tfsdk:"primary_locale"`
}

func (d *AppDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (d *AppDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an app in App Store Connect, " +
			"either by its App Store Connect ID or by its bundle ID. This is the recommended way to obtain the " +
			"app ID required by the In-App Purchase and subscription resources.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The App Store Connect ID of the app. Provide either `id` or `bundle_id`.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("bundle_id"),
					),
				},
			},
			"bundle_id": schema.StringAttribute{
				MarkdownDescription: "The bundle ID of the app (e.g., `com.example.app`). Provide either `id` or `bundle_id`.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the app.",
				Computed:            true,
			},
			"sku": schema.StringAttribute{
				MarkdownDescription: "The SKU of the app.",
				Computed:            true,
			},
			"primary_locale": schema.StringAttribute{
				MarkdownDescription: "The primary locale of the app.",
				Computed:            true,
			},
		},
	}
}

func (d *AppDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AppDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AppDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, fetch the specific app.
	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching App by ID", map[string]any{
			"id": data.ID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v1/apps/%s", data.ID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read App, got error: %s", err),
			)
			return
		}

		var app App
		if err := json.Unmarshal(apiResp.Data, &app); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse App response, got error: %s", err),
			)
			return
		}

		d.updateModel(&data, &app)
	} else {
		// Fetch by bundle ID.
		tflog.Debug(ctx, "Fetching App by bundle ID", map[string]any{
			"bundle_id": data.BundleID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: "/v1/apps",
			Query: map[string]string{
				"filter[bundleId]": data.BundleID.ValueString(),
				"limit":            "200",
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list Apps, got error: %s", err),
			)
			return
		}

		var apps []App
		if err := json.Unmarshal(apiResp.Data, &apps); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Apps response, got error: %s", err),
			)
			return
		}

		if len(apps) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				fmt.Sprintf("No app found with bundle ID '%s'", data.BundleID.ValueString()),
			)
			return
		}

		if len(apps) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Multiple apps found with bundle ID '%s'", data.BundleID.ValueString()),
			)
			return
		}

		d.updateModel(&data, &apps[0])
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the App data.
func (d *AppDataSource) updateModel(model *AppDataSourceModel, app *App) {
	model.ID = types.StringValue(app.ID)
	model.BundleID = types.StringValue(app.Attributes.BundleID)
	model.Name = types.StringValue(app.Attributes.Name)
	model.SKU = types.StringValue(app.Attributes.SKU)
	model.PrimaryLocale = types.StringValue(app.Attributes.PrimaryLocale)
}
