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
var _ datasource.DataSource = &SubscriptionDataSource{}

// NewSubscriptionDataSource creates a new subscription data source.
func NewSubscriptionDataSource() datasource.DataSource {
	return &SubscriptionDataSource{}
}

// SubscriptionDataSource defines the data source implementation.
type SubscriptionDataSource struct {
	client *Client
}

// SubscriptionDataSourceModel describes the data source data model.
type SubscriptionDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SubscriptionGroupID types.String `tfsdk:"subscription_group_id"`
	ProductID           types.String `tfsdk:"product_id"`
	Name                types.String `tfsdk:"name"`
	SubscriptionPeriod  types.String `tfsdk:"subscription_period"`
	FamilySharable      types.Bool   `tfsdk:"family_sharable"`
	GroupLevel          types.Int64  `tfsdk:"group_level"`
	ReviewNote          types.String `tfsdk:"review_note"`
	State               types.String `tfsdk:"state"`
	Filter              types.Object `tfsdk:"filter"`
}

// SubscriptionFilterModel describes the filter criteria. Product IDs are only
// unique within a subscription group, so both fields are required.
type SubscriptionFilterModel struct {
	SubscriptionGroupID types.String `tfsdk:"subscription_group_id"`
	ProductID           types.String `tfsdk:"product_id"`
}

func (d *SubscriptionDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

func (d *SubscriptionDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing auto-renewable subscription, either by its ID or by subscription group ID and product ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the subscription.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("filter"),
					),
				},
			},
			"subscription_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription group the subscription belongs to.",
				Computed:            true,
			},
			"product_id": schema.StringAttribute{
				MarkdownDescription: "The product ID of the subscription.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The reference name of the subscription.",
				Computed:            true,
			},
			"subscription_period": schema.StringAttribute{
				MarkdownDescription: "The subscription period.",
				Computed:            true,
			},
			"family_sharable": schema.BoolAttribute{
				MarkdownDescription: "Whether the subscription is available through Family Sharing.",
				Computed:            true,
			},
			"group_level": schema.Int64Attribute{
				MarkdownDescription: "The ranking of the subscription within its group.",
				Computed:            true,
			},
			"review_note": schema.StringAttribute{
				MarkdownDescription: "The note for the App Review team.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The state of the subscription.",
				Computed:            true,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for finding a subscription.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"subscription_group_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the subscription group to search within.",
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

func (d *SubscriptionDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SubscriptionDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SubscriptionDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching subscription by ID", map[string]any{
			"id": data.ID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v1/subscriptions/%s", data.ID.ValueString()),
			Query: map[string]string{
				"include": "group",
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to read subscription, got error: %s", err),
			)
			return
		}

		var subscription Subscription
		if err := json.Unmarshal(apiResp.Data, &subscription); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse subscription response, got error: %s", err),
			)
			return
		}

		d.updateModel(&data, &subscription)
	} else {
		var filter SubscriptionFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, "Fetching subscription by filter", map[string]any{
			"subscription_group_id": filter.SubscriptionGroupID.ValueString(),
			"product_id":            filter.ProductID.ValueString(),
		})

		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: fmt.Sprintf("/v1/subscriptionGroups/%s/subscriptions", filter.SubscriptionGroupID.ValueString()),
			Query: map[string]string{
				"filter[productId]": filter.ProductID.ValueString(),
				"limit":             "200",
			},
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list subscriptions, got error: %s", err),
			)
			return
		}

		var subscriptions []Subscription
		if err := json.Unmarshal(apiResp.Data, &subscriptions); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse subscriptions response, got error: %s", err),
			)
			return
		}

		if len(subscriptions) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				fmt.Sprintf("No subscription found with product ID '%s' in group '%s'", filter.ProductID.ValueString(), filter.SubscriptionGroupID.ValueString()),
			)
			return
		}

		if len(subscriptions) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Multiple subscriptions found with product ID '%s' in group '%s'", filter.ProductID.ValueString(), filter.SubscriptionGroupID.ValueString()),
			)
			return
		}

		d.updateModel(&data, &subscriptions[0])
		data.SubscriptionGroupID = filter.SubscriptionGroupID
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the subscription data.
func (d *SubscriptionDataSource) updateModel(model *SubscriptionDataSourceModel, subscription *Subscription) {
	model.ID = types.StringValue(subscription.ID)
	model.ProductID = types.StringValue(subscription.Attributes.ProductID)
	model.Name = types.StringValue(subscription.Attributes.Name)
	model.SubscriptionPeriod = types.StringValue(subscription.Attributes.SubscriptionPeriod)
	model.FamilySharable = types.BoolValue(subscription.Attributes.FamilySharable)
	model.GroupLevel = types.Int64Value(subscription.Attributes.GroupLevel)
	model.ReviewNote = types.StringValue(subscription.Attributes.ReviewNote)
	model.State = types.StringValue(subscription.Attributes.State)

	if subscription.Relationships != nil && subscription.Relationships.Group != nil && subscription.Relationships.Group.Data != nil {
		model.SubscriptionGroupID = types.StringValue(subscription.Relationships.Group.Data.ID)
	}
}
