// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"time"

	"software.sslmate.com/src/go-pkcs12"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	ID                    types.String `tfsdk:"id"`
	CertificateType       types.String `tfsdk:"certificate_type"`
	CsrContent            types.String `tfsdk:"csr_content"`
	PrivateKeyPEM         types.String `tfsdk:"private_key_pem"`
	CertificateContent    types.String `tfsdk:"certificate_content"`
	CertificateContentPEM types.String `tfsdk:"certificate_content_pem"`
	CertificateCAIssuers  types.List   `tfsdk:"certificate_ca_issuers"`
	DisplayName           types.String `tfsdk:"display_name"`
	Name                  types.String `tfsdk:"name"`
	Platform              types.String `tfsdk:"platform"`
	SerialNumber          types.String `tfsdk:"serial_number"`
	ExpirationDate        types.String `tfsdk:"expiration_date"`
	RecreateThreshold     types.Int64  `tfsdk:"recreate_threshold"`
	Relationships         types.Object `tfsdk:"relationships"`
	PKCS12BundlePassword  types.String `tfsdk:"pkcs12_bundle_password"`
	PKCS12BundleContent   types.String `tfsdk:"pkcs12_bundle_content"`
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
			"private_key_pem": schema.StringAttribute{
				MarkdownDescription: "The private key in PEM format. Only required if you want to generate a PKCS12 bundle. This is not sent to Apple's API and is only used locally for PKCS12 generation. Changes to this value do not require certificate replacement.",
				Optional:            true,
				Sensitive:           true,
			},
			"certificate_content": schema.StringAttribute{
				MarkdownDescription: "The certificate content in base64 encoded DER format.",
				Computed:            true,
				Sensitive:           true,
			},
			"certificate_content_pem": schema.StringAttribute{
				MarkdownDescription: "The certificate content in base64 encoded PEM format.",
				Computed:            true,
				Sensitive:           true,
			},
			"certificate_ca_issuers": schema.ListAttribute{
				MarkdownDescription: "A list of CA Issuer URIs from the Authority Information Access extension.",
				Computed:            true,
				ElementType:         types.StringType,
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
				PlanModifiers: []planmodifier.String{
					NewCertificateRecreateThresholdPlanModifier(),
				},
			},
			"recreate_threshold": schema.Int64Attribute{
				MarkdownDescription: "The number of seconds before certificate expiration when Terraform should recreate the certificate. Set to 0 to disable automatic recreation. Default is 2592000 seconds (30 days).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
					int64planmodifier.RequiresReplace(),
					NewRecreateThresholdDefaultPlanModifier(),
				},
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
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
			"pkcs12_bundle_password": schema.StringAttribute{
				MarkdownDescription: "Password to use for the PKCS12 bundle. When provided, a PKCS12 bundle will be generated and available in the `pkcs12_bundle_content` attribute. Changes to this value do not require certificate replacement.",
				Optional:            true,
				Sensitive:           true,
			},
			"pkcs12_bundle_content": schema.StringAttribute{
				MarkdownDescription: "The PKCS12 bundle content in base64 encoded format. Only available when `pkcs12_bundle_password` is provided.",
				Computed:            true,
				Sensitive:           true,
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
	var cert Certificate
	if err := json.Unmarshal(apiResp.Data, &cert); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Certificate response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data
	data.ID = types.StringValue(cert.ID)
	data.CertificateContent = types.StringValue(cert.Attributes.CertificateContent)
	data.DisplayName = types.StringValue(cert.Attributes.DisplayName)
	data.Name = types.StringValue(cert.Attributes.Name)
	data.Platform = types.StringValue(cert.Attributes.Platform)
	data.SerialNumber = types.StringValue(cert.Attributes.SerialNumber)

	// Convert DER to PEM format
	if cert.Attributes.CertificateContent != "" {
		pemContent, err := convertDERToPEM(cert.Attributes.CertificateContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Certificate Conversion Error",
				fmt.Sprintf("Unable to convert certificate to PEM format: %s", err),
			)
			return
		}
		data.CertificateContentPEM = types.StringValue(pemContent)
	} else {
		data.CertificateContentPEM = types.StringNull()
	}

	// Extract certificate CA issuers
	if cert.Attributes.CertificateContent != "" {
		caIssuers, err := extractCertificateCAIssuers(cert.Attributes.CertificateContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Certificate CA Issuers Parsing Error",
				fmt.Sprintf("Unable to parse certificate CA issuers: %s", err),
			)
			return
		}

		// Convert []string to types.List
		issuerValues := make([]attr.Value, len(caIssuers))
		for i, issuer := range caIssuers {
			issuerValues[i] = types.StringValue(issuer)
		}

		issuerList, diags := types.ListValue(types.StringType, issuerValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.CertificateCAIssuers = issuerList
	} else {
		data.CertificateCAIssuers = types.ListNull(types.StringType)
	}

	if cert.Attributes.ExpirationDate != nil {
		data.ExpirationDate = types.StringValue(cert.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
	} else {
		// Set to null if not provided by API
		data.ExpirationDate = types.StringNull()
	}

	// Set default recreate threshold if not provided in plan
	if data.RecreateThreshold.IsNull() || data.RecreateThreshold.IsUnknown() {
		data.RecreateThreshold = types.Int64Value(2592000) // 30 days
	}
	// Note: recreate_threshold is preserved from plan as it's not returned by Apple API

	// Generate PKCS12 bundle if needed
	if err := updatePKCS12Bundle(&data); err != nil {
		resp.Diagnostics.AddError(
			"PKCS12 Bundle Generation Error",
			fmt.Sprintf("Unable to generate PKCS12 bundle: %s", err),
		)
		return
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

	// Preserve PKCS12-related fields from existing state
	existingPrivateKeyPEM := data.PrivateKeyPEM
	existingPKCS12Password := data.PKCS12BundlePassword
	existingPKCS12Content := data.PKCS12BundleContent

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
	var cert Certificate
	if err := json.Unmarshal(apiResp.Data, &cert); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Certificate response, got error: %s", err),
		)
		return
	}

	// Update the model with the response data
	data.CertificateType = types.StringValue(cert.Attributes.CertificateType)
	data.CertificateContent = types.StringValue(cert.Attributes.CertificateContent)
	data.DisplayName = types.StringValue(cert.Attributes.DisplayName)
	data.Name = types.StringValue(cert.Attributes.Name)
	data.Platform = types.StringValue(cert.Attributes.Platform)
	data.SerialNumber = types.StringValue(cert.Attributes.SerialNumber)

	// Convert DER to PEM format
	if cert.Attributes.CertificateContent != "" {
		pemContent, err := convertDERToPEM(cert.Attributes.CertificateContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Certificate Conversion Error",
				fmt.Sprintf("Unable to convert certificate to PEM format: %s", err),
			)
			return
		}
		data.CertificateContentPEM = types.StringValue(pemContent)
	} else {
		data.CertificateContentPEM = types.StringNull()
	}

	// Extract certificate CA issuers
	if cert.Attributes.CertificateContent != "" {
		caIssuers, err := extractCertificateCAIssuers(cert.Attributes.CertificateContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Certificate CA Issuers Parsing Error",
				fmt.Sprintf("Unable to parse certificate CA issuers: %s", err),
			)
			return
		}

		// Convert []string to types.List
		issuerValues := make([]attr.Value, len(caIssuers))
		for i, issuer := range caIssuers {
			issuerValues[i] = types.StringValue(issuer)
		}

		issuerList, diags := types.ListValue(types.StringType, issuerValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.CertificateCAIssuers = issuerList
	} else {
		data.CertificateCAIssuers = types.ListNull(types.StringType)
	}

	if cert.Attributes.ExpirationDate != nil {
		data.ExpirationDate = types.StringValue(cert.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
	} else {
		// Set to null if not provided by API
		data.ExpirationDate = types.StringNull()
	}

	// Update relationships if present
	if cert.Relationships != nil && cert.Relationships.PassTypeId != nil && cert.Relationships.PassTypeId.Data != nil {
		relationshipsMap := map[string]attr.Value{
			"pass_type_id": types.StringValue(cert.Relationships.PassTypeId.Data.ID),
		}
		relationshipsObj, diags := types.ObjectValue(map[string]attr.Type{
			"pass_type_id": types.StringType,
		}, relationshipsMap)
		resp.Diagnostics.Append(diags...)
		data.Relationships = relationshipsObj
	}

	// Restore PKCS12-related fields from existing state to avoid unnecessary changes
	// PKCS12 bundle generation only happens during Create/Update operations
	data.PrivateKeyPEM = existingPrivateKeyPEM
	data.PKCS12BundlePassword = existingPKCS12Password
	data.PKCS12BundleContent = existingPKCS12Content

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	var state CertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if only PKCS12-related fields have changed
	certificateFieldsChanged := !plan.CertificateType.Equal(state.CertificateType) ||
		!plan.CsrContent.Equal(state.CsrContent) ||
		!plan.Relationships.Equal(state.Relationships)

	if certificateFieldsChanged {
		resp.Diagnostics.AddError(
			"Update Not Supported",
			"The certificate itself cannot be updated. Only pkcs12_bundle_password and private_key_pem can be modified without replacement. To change the certificate, you must delete and recreate the resource.",
		)
		return
	}

	// Copy all the computed fields from state to plan
	plan.ID = state.ID
	plan.CertificateContent = state.CertificateContent
	plan.CertificateContentPEM = state.CertificateContentPEM
	plan.CertificateCAIssuers = state.CertificateCAIssuers
	plan.DisplayName = state.DisplayName
	plan.Name = state.Name
	plan.Platform = state.Platform
	plan.SerialNumber = state.SerialNumber
	plan.ExpirationDate = state.ExpirationDate
	plan.RecreateThreshold = state.RecreateThreshold
	plan.Relationships = state.Relationships

	// Generate PKCS12 bundle with the new values
	if err := updatePKCS12Bundle(&plan); err != nil {
		resp.Diagnostics.AddError(
			"PKCS12 Bundle Generation Error",
			fmt.Sprintf("Unable to generate PKCS12 bundle: %s", err),
		)
		return
	}

	tflog.Trace(ctx, "Updated Certificate PKCS12 bundle", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CertificateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Removing Certificate from Terraform state", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Certificates cannot be revoked programmatically through the App Store Connect API.
	// According to Apple's documentation, certificates can only be revoked by Apple Developer Program Support.
	// Therefore, we only remove the certificate from Terraform state.

	// Add a warning to inform users about this limitation
	resp.Diagnostics.AddWarning(
		"Certificate Not Revoked",
		"The certificate has been removed from Terraform state, but it cannot be revoked programmatically through the App Store Connect API. "+
			"If you need to revoke this certificate, you must contact Apple Developer Program Support at https://developer.apple.com/support",
	)

	tflog.Trace(ctx, "Removed Certificate from Terraform state", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// generatePKCS12Bundle creates a PKCS12 bundle from certificate and private key.
func generatePKCS12Bundle(certPEM, privateKeyPEM, password string) (string, error) {
	// Parse certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		return "", fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(privateKeyPEM))
	if keyBlock == nil {
		return "", fmt.Errorf("failed to decode private key PEM")
	}

	var privateKey interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		return "", fmt.Errorf("unsupported private key type: %s", keyBlock.Type)
	}

	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	// Create PKCS12
	p12Data, err := pkcs12.Modern.Encode(privateKey, cert, nil, password)
	if err != nil {
		return "", fmt.Errorf("failed to encode PKCS12: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(p12Data), nil
}

// updatePKCS12Bundle generates PKCS12 bundle if both password and private key are provided.
func updatePKCS12Bundle(data *CertificateResourceModel) error {
	// Only generate PKCS12 if both password and private key are provided, and certificate is available
	if !data.PKCS12BundlePassword.IsNull() && !data.PKCS12BundlePassword.IsUnknown() &&
		!data.PrivateKeyPEM.IsNull() && !data.PrivateKeyPEM.IsUnknown() &&
		!data.CertificateContentPEM.IsNull() && !data.CertificateContentPEM.IsUnknown() {

		// Decode the base64-encoded PEM to get the raw PEM string
		certPEMBytes, err := base64.StdEncoding.DecodeString(data.CertificateContentPEM.ValueString())
		if err != nil {
			return fmt.Errorf("failed to decode base64 certificate PEM: %w", err)
		}

		pkcs12Content, err := generatePKCS12Bundle(
			string(certPEMBytes),
			data.PrivateKeyPEM.ValueString(),
			data.PKCS12BundlePassword.ValueString(),
		)
		if err != nil {
			return fmt.Errorf("failed to generate PKCS12 bundle: %w", err)
		}
		data.PKCS12BundleContent = types.StringValue(pkcs12Content)
	} else {
		data.PKCS12BundleContent = types.StringNull()
	}
	return nil
}

// CertificateRecreateThresholdPlanModifier is a custom plan modifier that triggers replacement
// when the certificate is within the recreate threshold of expiration.
type CertificateRecreateThresholdPlanModifier struct{}

// NewCertificateRecreateThresholdPlanModifier creates a new instance of the plan modifier.
func NewCertificateRecreateThresholdPlanModifier() planmodifier.String {
	return CertificateRecreateThresholdPlanModifier{}
}

// Description returns a human-readable description of the plan modifier.
func (m CertificateRecreateThresholdPlanModifier) Description(ctx context.Context) string {
	return "Recreates the certificate when it is within the recreate threshold of expiration."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m CertificateRecreateThresholdPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Recreates the certificate when it is within the recreate threshold of expiration."
}

// PlanModifyString implements the plan modifier logic.
func (m CertificateRecreateThresholdPlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// If the resource is being created, don't modify the plan
	if req.State.Raw.IsNull() {
		return
	}

	// If the resource is being destroyed, don't modify the plan
	if req.Plan.Raw.IsNull() {
		return
	}

	// If the expiration date is not set, don't modify the plan
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	// Get the current state
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the planned state
	var plan CertificateResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the recreate threshold (default to 30 days if not set)
	var thresholdSeconds int64 = 2592000 // 30 days
	if !plan.RecreateThreshold.IsNull() && !plan.RecreateThreshold.IsUnknown() {
		thresholdSeconds = plan.RecreateThreshold.ValueInt64()
	}

	// If threshold is 0, don't recreate
	if thresholdSeconds == 0 {
		return
	}

	// Parse the expiration date
	expirationStr := state.ExpirationDate.ValueString()
	if expirationStr == "" {
		return
	}

	expirationDate, err := time.Parse("2006-01-02T15:04:05Z", expirationStr)
	if err != nil {
		tflog.Warn(ctx, "Failed to parse expiration date", map[string]interface{}{
			"expiration_date": expirationStr,
			"error":           err.Error(),
		})
		return
	}

	// Calculate the threshold time
	thresholdTime := time.Now().Add(time.Duration(thresholdSeconds) * time.Second)

	// If the certificate expires within the threshold, require replacement
	if expirationDate.Before(thresholdTime) {
		tflog.Info(ctx, "Certificate expiration is within recreate threshold, requiring replacement", map[string]interface{}{
			"expiration_date":   expirationDate.Format("2006-01-02T15:04:05Z"),
			"threshold_time":    thresholdTime.Format("2006-01-02T15:04:05Z"),
			"threshold_seconds": thresholdSeconds,
		})
		resp.RequiresReplace = true
	}
}

// RecreateThresholdDefaultPlanModifier sets a default value for recreate_threshold.
type RecreateThresholdDefaultPlanModifier struct{}

// NewRecreateThresholdDefaultPlanModifier creates a new instance of the default plan modifier.
func NewRecreateThresholdDefaultPlanModifier() planmodifier.Int64 {
	return RecreateThresholdDefaultPlanModifier{}
}

// Description returns a human-readable description of the plan modifier.
func (m RecreateThresholdDefaultPlanModifier) Description(ctx context.Context) string {
	return "Sets default value of 2592000 (30 days) when recreate_threshold is not specified."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m RecreateThresholdDefaultPlanModifier) MarkdownDescription(ctx context.Context) string {
	return "Sets default value of 2592000 (30 days) when recreate_threshold is not specified."
}

// PlanModifyInt64 implements the plan modifier logic.
func (m RecreateThresholdDefaultPlanModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// If the value is null or unknown, set the default
	if req.ConfigValue.IsNull() {
		resp.PlanValue = types.Int64Value(2592000) // 30 days
	}
}
