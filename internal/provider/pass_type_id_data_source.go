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
var _ datasource.DataSource = &PassTypeIDDataSource{}

// NewPassTypeIDDataSource creates a new Pass Type ID data source.
func NewPassTypeIDDataSource() datasource.DataSource {
	return &PassTypeIDDataSource{}
}

// PassTypeIDDataSource defines the data source implementation.
type PassTypeIDDataSource struct {
	client *Client
}

// PassTypeIDDataSourceModel describes the data source data model.
type PassTypeIDDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Identifier  types.String `tfsdk:"identifier"`
	Description types.String `tfsdk:"description"`
	CreatedDate types.String `tfsdk:"created_date"`
	// Filter attributes
	Filter types.Object `tfsdk:"filter"`
}

// PassTypeIDFilterModel describes the filter criteria.
type PassTypeIDFilterModel struct {
	Identifier types.String `tfsdk:"identifier"`
}

func (d *PassTypeIDDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pass_type_id"
}

func (d *PassTypeIDDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing Pass Type ID in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the Pass Type ID.",
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
				MarkdownDescription: "The identifier for the Pass Type ID (e.g., 'pass.io.truetickets.test.membership').",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the Pass Type ID.",
				Computed:            true,
			},
			"created_date": schema.StringAttribute{
				MarkdownDescription: "The date when the Pass Type ID was created.",
				Computed:            true,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for finding a Pass Type ID.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"identifier": schema.StringAttribute{
						MarkdownDescription: "The identifier to search for (e.g., 'pass.io.truetickets.test.membership').",
						Required:            true,
					},
				},
			},
		},
	}
}

func (d *PassTypeIDDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PassTypeIDDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PassTypeIDDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, fetch specific Pass Type ID
	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching Pass Type ID by ID", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		// Make the API request
		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/passTypeIds/%s", data.ID.ValueString()),
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read Pass Type ID, got error: %s", err),
			)
			return
		}

		// Parse the response
		var passTypeID PassTypeID
		if err := json.Unmarshal(apiResp.Data, &passTypeID); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Pass Type ID response, got error: %s", err),
			)
			return
		}

		// Update the model with the response data
		d.updateModel(&data, &passTypeID)

	} else if !data.Filter.IsNull() {
		// Extract filter criteria
		var filter PassTypeIDFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "Fetching Pass Type IDs with filter", map[string]interface{}{
			"identifier": filter.Identifier.ValueString(),
		})

		// Make the API request to list all Pass Type IDs
		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: "/passTypeIds",
			Query: map[string]string{
				"filter[identifier]": filter.Identifier.ValueString(),
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list Pass Type IDs, got error: %s", err),
			)
			return
		}

		// Parse the response - the API returns an array directly in the data field
		var passTypeIDs []PassTypeID
		if err := json.Unmarshal(apiResp.Data, &passTypeIDs); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Pass Type IDs response, got error: %s", err),
			)
			return
		}

		// Check if we found exactly one result
		if len(passTypeIDs) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				fmt.Sprintf("No Pass Type ID found with identifier '%s'", filter.Identifier.ValueString()),
			)
			return
		}

		if len(passTypeIDs) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Multiple Pass Type IDs found with identifier '%s'", filter.Identifier.ValueString()),
			)
			return
		}

		// Update the model with the first (and only) result
		d.updateModel(&data, &passTypeIDs[0])
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the Pass Type ID data.
func (d *PassTypeIDDataSource) updateModel(model *PassTypeIDDataSourceModel, passTypeID *PassTypeID) {
	model.ID = types.StringValue(passTypeID.ID)
	model.Identifier = types.StringValue(passTypeID.Attributes.Identifier)
	model.Description = types.StringValue(passTypeID.Attributes.Name)
	if passTypeID.Attributes.CreatedDate != nil {
		model.CreatedDate = types.StringValue(passTypeID.Attributes.CreatedDate.Format("2006-01-02T15:04:05Z"))
	} else {
		// Set to null if not provided by API
		model.CreatedDate = types.StringNull()
	}
}
