// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &CertificateResource{}
var _ resource.ResourceWithImportState = &CertificateResource{}

// NewCertificateResource creates a new Certificate resource.
func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

// CertificateResource defines the resource implementation.
type CertificateResource struct {
	client *Client
}

// CertificateResourceModel describes the resource data model.
type CertificateResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	CertificateType    types.String `tfsdk:"certificate_type"`
	CsrContent         types.String `tfsdk:"csr_content"`
	CertificateContent types.String `tfsdk:"certificate_content"`
	DisplayName        types.String `tfsdk:"display_name"`
	Name               types.String `tfsdk:"name"`
	Platform           types.String `tfsdk:"platform"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	Relationships      types.Object `tfsdk:"relationships"`
}

// CertificateRelationshipsModel describes the relationships data model.
type CertificateRelationshipsModel struct {
	PassTypeId types.String `tfsdk:"pass_type_id"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Certificate in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the Certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_type": schema.StringAttribute{
				MarkdownDescription: "The type of certificate to create. Valid values are: `IOS_DEVELOPMENT`, `IOS_DISTRIBUTION`, `MAC_APP_DEVELOPMENT`, `MAC_APP_DISTRIBUTION`, `MAC_INSTALLER_DISTRIBUTION`, `PASS_TYPE_ID`, `PASS_TYPE_ID_WITH_NFC`, `DEVELOPER_ID_KEXT`, `DEVELOPER_ID_APPLICATION`, `DEVELOPMENT_PUSH_SSL`, `PRODUCTION_PUSH_SSL`, `PUSH_SSL`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						CertificateTypeIOSDevelopment,
						CertificateTypeIOSDistribution,
						CertificateTypeMacAppDevelopment,
						CertificateTypeMacAppDistribution,
						CertificateTypeMacInstallerDistribution,
						CertificateTypePassTypeID,
						CertificateTypePassTypeIDWithNFC,
						CertificateTypeDeveloperIDKext,
						CertificateTypeDeveloperIDApplication,
						CertificateTypeDevelopmentPushSSL,
						CertificateTypeProductionPushSSL,
						CertificateTypePushSSL,
					),
				},
			},
			"csr_content": schema.StringAttribute{
				MarkdownDescription: "The certificate signing request (CSR) content in PEM format.",
				Required:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_content": schema.StringAttribute{
				MarkdownDescription: "The certificate content in PEM format.",
				Computed:            true,
				Sensitive:           true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "The display name of the certificate.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the certificate.",
				Computed:            true,
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform for the certificate.",
				Computed:            true,
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "The serial number of the certificate.",
				Computed:            true,
			},
			"expiration_date": schema.StringAttribute{
				MarkdownDescription: "The expiration date of the certificate.",
				Computed:            true,
			},
			"relationships": schema.SingleNestedAttribute{
				MarkdownDescription: "The relationships for the certificate.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"pass_type_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the Pass Type ID to associate with this certificate. Required for PASS_TYPE_ID and PASS_TYPE_ID_WITH_NFC certificate types.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Extract relationships if present
	var relationships CertificateRelationshipsModel
	if !data.Relationships.IsNull() && !data.Relationships.IsUnknown() {
		resp.Diagnostics.Append(data.Relationships.As(ctx, &relationships, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Validate Pass Type ID requirement
	certType := data.CertificateType.ValueString()
	if (certType == CertificateTypePassTypeID || certType == CertificateTypePassTypeIDWithNFC) && relationships.PassTypeId.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("relationships").AtName("pass_type_id"),
			"Missing Pass Type ID",
			"Pass Type ID is required for PASS_TYPE_ID and PASS_TYPE_ID_WITH_NFC certificate types.",
		)
		return
	}

	// Create the request
	createReq := CertificateCreateRequest{
		Data: CertificateCreateRequestData{
			Type: "certificates",
			Attributes: CertificateCreateRequestAttributes{
				CertificateType: certType,
				CsrContent:      data.CsrContent.ValueString(),
			},
		},
	}

	// Add relationships if present
	if !relationships.PassTypeId.IsNull() {
		createReq.Data.Relationships = &CertificateCreateRequestRelationships{
			PassTypeId: &CertificateCreateRequestRelationship{
				Data: RelationshipData{
					Type: "passTypeIds",
					ID:   relationships.PassTypeId.ValueString(),
				},
			},
		}
	}

	tflog.Debug(ctx, "Creating Certificate", map[string]interface{}{
		"certificate_type": certType,
		"has_pass_type_id": !relationships.PassTypeId.IsNull(),
	})

	// Make the API request
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodPost,
		Endpoint: "/certificates",
		Body:     createReq,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to create Certificate, got error: %s", err),
		)
		return
	}

	// Parse the response
	var certResp CertificateResponse
	if err := json.Unmarshal(apiResp.Data, &certResp); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Certificate response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data
	data.ID = types.StringValue(certResp.Data.ID)
	data.CertificateContent = types.StringValue(certResp.Data.Attributes.CertificateContent)
	data.DisplayName = types.StringValue(certResp.Data.Attributes.DisplayName)
	data.Name = types.StringValue(certResp.Data.Attributes.Name)
	data.Platform = types.StringValue(certResp.Data.Attributes.Platform)
	data.SerialNumber = types.StringValue(certResp.Data.Attributes.SerialNumber)

	if certResp.Data.Attributes.ExpirationDate != nil {
		data.ExpirationDate = types.StringValue(certResp.Data.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
	}

	tflog.Trace(ctx, "Created Certificate", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CertificateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Certificate", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Make the API request
	apiResp, err := r.client.Do(ctx, Request{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("/certificates/%s", data.ID.ValueString()),
		Query: map[string]string{
			"include": "passTypeId",
		},
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to read Certificate, got error: %s", err),
		)
		return
	}

	// Parse the response
	var certResp CertificateResponse
	if err := json.Unmarshal(apiResp.Data, &certResp); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Certificate response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data
	data.CertificateType = types.StringValue(certResp.Data.Attributes.CertificateType)
	data.CertificateContent = types.StringValue(certResp.Data.Attributes.CertificateContent)
	data.DisplayName = types.StringValue(certResp.Data.Attributes.DisplayName)
	data.Name = types.StringValue(certResp.Data.Attributes.Name)
	data.Platform = types.StringValue(certResp.Data.Attributes.Platform)
	data.SerialNumber = types.StringValue(certResp.Data.Attributes.SerialNumber)

	if certResp.Data.Attributes.ExpirationDate != nil {
		data.ExpirationDate = types.StringValue(certResp.Data.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
	}

	// Update relationships if present
	if certResp.Data.Relationships != nil && certResp.Data.Relationships.PassTypeId != nil && certResp.Data.Relationships.PassTypeId.Data != nil {
		relationshipsMap := map[string]attr.Value{
			"pass_type_id": types.StringValue(certResp.Data.Relationships.PassTypeId.Data.ID),
		}
		relationshipsObj, diags := types.ObjectValue(map[string]attr.Type{
			"pass_type_id": types.StringType,
		}, relationshipsMap)
		resp.Diagnostics.Append(diags...)
		data.Relationships = relationshipsObj
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Certificates cannot be updated via the API
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Certificates cannot be updated. To change the certificate, you must delete and recreate the resource.",
	)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Revoking Certificate", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Make the API request to revoke the certificate
	_, err := r.client.Do(ctx, Request{
		Method:   http.MethodDelete,
		Endpoint: fmt.Sprintf("/certificates/%s", data.ID.ValueString()),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to revoke Certificate, got error: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "Revoked Certificate", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
