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
var _ datasource.DataSource = &BundleIDDataSource{}

// NewBundleIDDataSource creates a new Bundle ID data source.
func NewBundleIDDataSource() datasource.DataSource {
	return &BundleIDDataSource{}
}

// BundleIDDataSource defines the data source implementation.
type BundleIDDataSource struct {
	client *Client
}

// BundleIDDataSourceModel describes the data source data model.
type BundleIDDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Platform   types.String `tfsdk:"platform"`
	SeedID     types.String `tfsdk:"seed_id"`
	// Filter attributes.
	Filter types.Object `tfsdk:"filter"`
}

// BundleIDFilterModel describes the filter criteria.
type BundleIDFilterModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

func (d *BundleIDDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bundle_id"
}

func (d *BundleIDDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing Bundle ID (App ID) in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (resource ID) of the Bundle ID.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("filter"),
					),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The reverse-DNS bundle identifier (e.g., 'com.example.app').",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Bundle ID.",
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform of the Bundle ID (`IOS`, `MAC_OS`, or `UNIVERSAL`).",
				Computed:            true,
			},
			"seed_id": schema.StringAttribute{
				MarkdownDescription: "The seed ID (team ID prefix) of the Bundle ID.",
				Computed:            true,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for finding a Bundle ID.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"identifier": schema.StringAttribute{
						MarkdownDescription: "The identifier to search for (e.g., 'com.example.app').",
						Required:            true,
					},
				},
			},
		},
	}
}

func (d *BundleIDDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BundleIDDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BundleIDDataSourceModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, fetch the specific Bundle ID.
	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching Bundle ID by ID", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		// Make the API request.
		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/bundleIds/%s", data.ID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read Bundle ID, got error: %s", err),
			)
			return
		}

		// Parse the response.
		var bundleID BundleID
		if err := json.Unmarshal(apiResp.Data, &bundleID); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Bundle ID response, got error: %s", err),
			)
			return
		}

		// Update the model with the response data.
		d.updateModel(&data, &bundleID)

	} else if !data.Filter.IsNull() {
		// Extract filter criteria.
		var filter BundleIDFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "Fetching Bundle IDs with filter", map[string]interface{}{
			"identifier": filter.Identifier.ValueString(),
		})

		// Make the API request to list Bundle IDs matching the identifier.
		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: "/bundleIds",
			Query: map[string]string{
				"filter[identifier]": filter.Identifier.ValueString(),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list Bundle IDs, got error: %s", err),
			)
			return
		}

		// Parse the response - the API returns an array directly in the data field.
		var bundleIDs []BundleID
		if err := json.Unmarshal(apiResp.Data, &bundleIDs); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Bundle IDs response, got error: %s", err),
			)
			return
		}

		// The identifier filter is a substring match on Apple's side, so make
		// sure exactly one result matches the requested identifier exactly.
		var matches []BundleID
		for _, b := range bundleIDs {
			if b.Attributes.Identifier == filter.Identifier.ValueString() {
				matches = append(matches, b)
			}
		}

		if len(matches) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				fmt.Sprintf("No Bundle ID found with identifier '%s'", filter.Identifier.ValueString()),
			)
			return
		}

		if len(matches) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Multiple Bundle IDs found with identifier '%s'", filter.Identifier.ValueString()),
			)
			return
		}

		// Update the model with the single matching result.
		d.updateModel(&data, &matches[0])
	}

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the Bundle ID data.
func (d *BundleIDDataSource) updateModel(model *BundleIDDataSourceModel, bundleID *BundleID) {
	model.ID = types.StringValue(bundleID.ID)
	model.Identifier = types.StringValue(bundleID.Attributes.Identifier)
	model.Name = types.StringValue(bundleID.Attributes.Name)
	model.Platform = types.StringValue(bundleID.Attributes.Platform)
	if bundleID.Attributes.SeedID != "" {
		model.SeedID = types.StringValue(bundleID.Attributes.SeedID)
	} else {
		model.SeedID = types.StringNull()
	}
}
