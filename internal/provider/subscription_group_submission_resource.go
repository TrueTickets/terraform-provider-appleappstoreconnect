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
var _ resource.Resource = &SubscriptionGroupSubmissionResource{}
var _ resource.ResourceWithImportState = &SubscriptionGroupSubmissionResource{}

// NewSubscriptionGroupSubmissionResource creates a new subscription group
// submission resource.
func NewSubscriptionGroupSubmissionResource() resource.Resource {
	return &SubscriptionGroupSubmissionResource{}
}

// SubscriptionGroupSubmissionResource defines the resource implementation.
type SubscriptionGroupSubmissionResource struct {
	client *Client
}

// SubscriptionGroupSubmissionResourceModel describes the resource data model.
type SubscriptionGroupSubmissionResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	SubscriptionGroupID types.String `tfsdk:"subscription_group_id"`
}

func (r *SubscriptionGroupSubmissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_group_submission"
}

func (r *SubscriptionGroupSubmissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Submits a subscription group (and all its subscriptions, localizations and prices) for App Review. This is a " +
			"one-shot action: creating the resource submits the group. The submission cannot be updated or withdrawn through this " +
			"provider; destroying the resource only removes it from Terraform state.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the submission.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"subscription_group_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the subscription group to submit for review. Changing this forces a new submission.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SubscriptionGroupSubmissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SubscriptionGroupSubmissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriptionGroupSubmissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := SubscriptionGroupSubmissionCreateRequest{
		Data: SubscriptionGroupSubmissionCreateRequestData{
			Type: "subscriptionGroupSubmissions",
			Relationships: SubscriptionGroupSubmissionCreateRequestRelationships{
				SubscriptionGroup: RelationshipOne{
					Data: RelationshipData{Type: "subscriptionGroups", ID: data.SubscriptionGroupID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Submitting subscription group for review", map[string]any{
		"subscription_group_id": data.SubscriptionGroupID.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/subscriptionGroupSubmissions",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to submit subscription group, got error: %s", err),
		)
		return
	}

	var submission SubscriptionGroupSubmission
	if err := json.Unmarshal(apiResp.Data, &submission); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse submission response, got error: %s", err),
		)
		return
	}

	if submission.ID == "" {
		resp.Diagnostics.AddError(
			"Invalid API Response",
			"The API response did not contain a valid ID for the created submission",
		)
		return
	}

	data.ID = types.StringValue(submission.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupSubmissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriptionGroupSubmissionResourceModel

	// A submission is a write-once action with no meaningful attributes to
	// refresh, so state is preserved as-is.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupSubmissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// subscription_group_id is RequiresReplace, so Update is never reached with
	// a real change. Implemented to satisfy the interface.
	var data SubscriptionGroupSubmissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriptionGroupSubmissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriptionGroupSubmissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A submission cannot be withdrawn through the API; removing the resource
	// only drops it from Terraform state.
	resp.Diagnostics.AddWarning(
		"Submission Not Withdrawn",
		"The submission has been removed from Terraform state, but App Store Connect submissions cannot be withdrawn "+
			"programmatically. The subscription group keeps its current review status.",
	)
}

func (r *SubscriptionGroupSubmissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
