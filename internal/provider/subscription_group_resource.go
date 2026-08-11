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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SubscriptionGroupResource{}
var _ resource.ResourceWithImportState = &SubscriptionGroupResource{}

// NewSubscriptionGroupResource creates a new subscription group resource.
func NewSubscriptionGroupResource() resource.Resource {
	return &SubscriptionGroupResource{}
}

// SubscriptionGroupResource defines the resource implementation.
type SubscriptionGroupResource struct {
	client *Client
}

// SubscriptionGroupResourceModel describes the resource data model.
type SubscriptionGroupResourceModel struct {
	ID            types.String `tfsdk:"id"`
	AppID         types.String `tfsdk:"app_id"`
	ReferenceName types.String `tfsdk:"reference_name"`
}

func (r *SubscriptionGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_group"
}

func (r *SubscriptionGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a subscription group for an app in App Store Connect. A subscription group organizes related " +
			"auto-renewable subscriptions; a customer can have only one active subscription per group at a time.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the subscription group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The App Store Connect ID of the app this subscription group belongs to. Use the `appleappstoreconnect_app` data source to look it up. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"reference_name": schema.StringAttribute{
				MarkdownDescription: "The reference name of the subscription group, used in App Store Connect and Sales and Trends reports.",
				Required:            true,
			},
		},
	}
}

func (r *SubscriptionGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := SubscriptionGroupCreateRequest{
		Data: SubscriptionGroupCreateRequestData{
			Type: "subscriptionGroups",
			Attributes: SubscriptionGroupCreateRequestAttributes{
				ReferenceName: data.ReferenceName.ValueString(),
			},
			Relationships: SubscriptionGroupCreateRequestRelationships{
				App: RelationshipOne{
					Data: RelationshipData{Type: "apps", ID: data.AppID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Creating subscription group", map[string]any{
		"app_id":         data.AppID.ValueString(),
		"reference_name": data.ReferenceName.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/subscriptionGroups",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create subscription group, got error: %s", err),
		)
		return
	}

	var group SubscriptionGroup
	if err := json.Unmarshal(apiResp.Data, &group); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group response, got error: %s", err),
		)
		return
	}

	if group.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created subscription group",
		)
		return
	}

	data.ID = types.StringValue(group.ID)
	data.ReferenceName = types.StringValue(group.Attributes.ReferenceName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroups/%s", data.ID.ValueString()),
	})
	if err != nil {
		if IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read subscription group, got error: %s", err),
		)
		return
	}

	var group SubscriptionGroup
	if err := json.Unmarshal(apiResp.Data, &group); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group response, got error: %s", err),
		)
		return
	}

	data.ID = types.StringValue(group.ID)
	data.ReferenceName = types.StringValue(group.Attributes.ReferenceName)
	if group.Relationships != nil && group.Relationships.App != nil && group.Relationships.App.Data != nil {
		data.AppID = types.StringValue(group.Relationships.App.Data.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SubscriptionGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	referenceName := data.ReferenceName.ValueString()
	updateReq := SubscriptionGroupUpdateRequest{
		Data: SubscriptionGroupUpdateRequestData{
			Type: "subscriptionGroups",
			ID:   data.ID.ValueString(),
			Attributes: SubscriptionGroupUpdateRequestAttributes{
				ReferenceName: &referenceName,
			},
		},
	}

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPatch,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroups/%s", data.ID.ValueString()),
		Body:     updateReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update subscription group, got error: %s", err),
		)
		return
	}

	var group SubscriptionGroup
	if err := json.Unmarshal(apiResp.Data, &group); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse subscription group response, got error: %s", err),
		)
		return
	}

	data.ReferenceName = types.StringValue(group.Attributes.ReferenceName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/v1/subscriptionGroups/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to delete subscription group, got error: %s. Note that a subscription group can only be deleted when it contains no subscriptions.", err),
		)
		return
	}
}

func (r *SubscriptionGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
