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
var _ resource.Resource = &InAppPurchaseSubmissionResource{}
var _ resource.ResourceWithImportState = &InAppPurchaseSubmissionResource{}

// NewInAppPurchaseSubmissionResource creates a new In-App Purchase submission
// resource.
func NewInAppPurchaseSubmissionResource() resource.Resource {
	return &InAppPurchaseSubmissionResource{}
}

// InAppPurchaseSubmissionResource defines the resource implementation.
type InAppPurchaseSubmissionResource struct {
	client *Client
}

// InAppPurchaseSubmissionResourceModel describes the resource data model.
type InAppPurchaseSubmissionResourceModel struct {
	ID              types.String `tfsdk:"id"`
	InAppPurchaseID types.String `tfsdk:"in_app_purchase_id"`
}

func (r *InAppPurchaseSubmissionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_in_app_purchase_submission"
}

func (r *InAppPurchaseSubmissionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Submits an In-App Purchase for App Review. This is a one-shot action: creating the resource submits the " +
			"In-App Purchase (which must have complete metadata, localization, pricing and availability). The submission cannot be " +
			"updated or withdrawn through this provider; destroying the resource only removes it from Terraform state.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the submission.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"in_app_purchase_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the In-App Purchase to submit for review. Changing this forces a new submission.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *InAppPurchaseSubmissionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *InAppPurchaseSubmissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data InAppPurchaseSubmissionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := IAPSubmissionCreateRequest{
		Data: IAPSubmissionCreateRequestData{
			Type: "inAppPurchaseSubmissions",
			Relationships: IAPSubmissionCreateRelationships{
				InAppPurchase: RelationshipOne{
					Data: RelationshipData{Type: "inAppPurchases", ID: data.InAppPurchaseID.ValueString()},
				},
			},
		},
	}

	tflog.Debug(ctx, "Submitting In-App Purchase for review", map[string]any{
		"in_app_purchase_id": data.InAppPurchaseID.ValueString(),
	})

	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/v1/inAppPurchaseSubmissions",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to submit In-App Purchase, got error: %s", err),
		)
		return
	}

	var submission IAPSubmission
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

func (r *InAppPurchaseSubmissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data InAppPurchaseSubmissionResourceModel

	// A submission is a write-once action with no meaningful attributes to
	// refresh, so state is preserved as-is.
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseSubmissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// in_app_purchase_id is RequiresReplace, so Update is never reached with a
	// real change. Implemented to satisfy the interface.
	var data InAppPurchaseSubmissionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *InAppPurchaseSubmissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data InAppPurchaseSubmissionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// A submission cannot be withdrawn through the API; removing the resource
	// only drops it from Terraform state.
	resp.Diagnostics.AddWarning(
		"Submission Not Withdrawn",
		"The submission has been removed from Terraform state, but App Store Connect submissions cannot be withdrawn "+
			"programmatically. The In-App Purchase keeps its current review status.",
	)
}

func (r *InAppPurchaseSubmissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
