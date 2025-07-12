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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CertificateDataSource{}

// NewCertificateDataSource creates a new Certificate data source.
func NewCertificateDataSource() datasource.DataSource {
	return &CertificateDataSource{}
}

// CertificateDataSource defines the data source implementation.
type CertificateDataSource struct {
	client *Client
}

// CertificateDataSourceModel describes the data source data model.
type CertificateDataSourceModel struct {
	ID                    types.String `tfsdk:"id"`
	CertificateType       types.String `tfsdk:"certificate_type"`
	CertificateContent    types.String `tfsdk:"certificate_content"`
	CertificateContentPEM types.String `tfsdk:"certificate_content_pem"`
	CertificateExtensions types.Map    `tfsdk:"certificate_extensions"`
	DisplayName           types.String `tfsdk:"display_name"`
	Name                  types.String `tfsdk:"name"`
	Platform              types.String `tfsdk:"platform"`
	SerialNumber          types.String `tfsdk:"serial_number"`
	ExpirationDate        types.String `tfsdk:"expiration_date"`
	Relationships         types.Object `tfsdk:"relationships"`
	// Filter attributes
	Filter types.Object `tfsdk:"filter"`
}

// CertificateFilterModel describes the filter criteria.
type CertificateFilterModel struct {
	CertificateType types.String `tfsdk:"certificate_type"`
	SerialNumber    types.String `tfsdk:"serial_number"`
}

func (d *CertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about an existing Certificate in App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the Certificate.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("id"),
						path.MatchRoot("filter"),
					),
				},
			},
			"certificate_type": schema.StringAttribute{
				MarkdownDescription: "The type of certificate.",
				Computed:            true,
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
			"certificate_extensions": schema.MapAttribute{
				MarkdownDescription: "A map of X509v3 certificate extensions. Keys are extension names or OIDs, values are hex-encoded extension data. For common extensions, human-readable parsed values are also provided with '_parsed' suffix.",
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
			},
			"relationships": schema.SingleNestedAttribute{
				MarkdownDescription: "The relationships for the certificate.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"pass_type_id": schema.StringAttribute{
						MarkdownDescription: "The ID of the associated Pass Type ID.",
						Computed:            true,
					},
				},
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for finding a Certificate.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"certificate_type": schema.StringAttribute{
						MarkdownDescription: "The certificate type to filter by.",
						Optional:            true,
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
					"serial_number": schema.StringAttribute{
						MarkdownDescription: "The serial number to search for.",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (d *CertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificateDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If ID is provided, fetch specific Certificate
	if !data.ID.IsNull() {
		tflog.Debug(ctx, "Fetching Certificate by ID", map[string]interface{}{
			"id": data.ID.ValueString(),
		})

		// Make the API request
		apiResp, err := d.client.Do(ctx, Request{
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
		d.updateModel(&data, &certResp.Data, resp)

	} else if !data.Filter.IsNull() {
		// Extract filter criteria
		var filter CertificateFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Build query parameters
		query := make(map[string]string)
		query["include"] = "passTypeId"

		if !filter.CertificateType.IsNull() {
			query["filter[certificateType]"] = filter.CertificateType.ValueString()
		}

		tflog.Debug(ctx, "Fetching Certificates with filter", map[string]interface{}{
			"certificate_type": filter.CertificateType.ValueString(),
			"serial_number":    filter.SerialNumber.ValueString(),
		})

		// Make the API request to list certificates
		apiResp, err := d.client.Do(ctx, Request{
			Method:   http.MethodGet,
			Endpoint: "/certificates",
			Query:    query,
		})
		if err != nil {
			resp.Diagnostics.AddError(
				"Client Error",
				fmt.Sprintf("Unable to list Certificates, got error: %s", err),
			)
			return
		}

		// Parse the response
		var certsResp CertificatesResponse
		if err := json.Unmarshal(apiResp.Data, &certsResp); err != nil {
			resp.Diagnostics.AddError(
				"Parse Error",
				fmt.Sprintf("Unable to parse Certificates response, got error: %s", err),
			)
			return
		}

		// Filter by serial number if provided
		var matchingCerts []Certificate
		if !filter.SerialNumber.IsNull() {
			serialNumber := filter.SerialNumber.ValueString()
			for _, cert := range certsResp.Data {
				if cert.Attributes.SerialNumber == serialNumber {
					matchingCerts = append(matchingCerts, cert)
				}
			}
		} else {
			matchingCerts = certsResp.Data
		}

		// Check if we found exactly one result
		if len(matchingCerts) == 0 {
			resp.Diagnostics.AddError(
				"Not Found",
				"No Certificate found matching the filter criteria",
			)
			return
		}

		if len(matchingCerts) > 1 {
			resp.Diagnostics.AddError(
				"Multiple Results",
				fmt.Sprintf("Found %d Certificates matching the filter criteria. Please refine your filter.", len(matchingCerts)),
			)
			return
		}

		// Update the model with the first (and only) result
		d.updateModel(&data, &matchingCerts[0], resp)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateModel updates the data source model with the Certificate data.
func (d *CertificateDataSource) updateModel(model *CertificateDataSourceModel, cert *Certificate, resp *datasource.ReadResponse) {
	model.ID = types.StringValue(cert.ID)
	model.CertificateType = types.StringValue(cert.Attributes.CertificateType)
	model.CertificateContent = types.StringValue(cert.Attributes.CertificateContent)
	model.DisplayName = types.StringValue(cert.Attributes.DisplayName)
	model.Name = types.StringValue(cert.Attributes.Name)
	model.Platform = types.StringValue(cert.Attributes.Platform)
	model.SerialNumber = types.StringValue(cert.Attributes.SerialNumber)

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
		model.CertificateContentPEM = types.StringValue(pemContent)
	} else {
		model.CertificateContentPEM = types.StringNull()
	}

	// Extract certificate extensions
	if cert.Attributes.CertificateContent != "" {
		extensions, err := extractCertificateExtensions(cert.Attributes.CertificateContent)
		if err != nil {
			resp.Diagnostics.AddError(
				"Certificate Extension Parsing Error",
				fmt.Sprintf("Unable to parse certificate extensions: %s", err),
			)
			return
		}

		// Convert map[string]string to types.Map
		extensionValues := make(map[string]attr.Value)
		for k, v := range extensions {
			extensionValues[k] = types.StringValue(v)
		}

		extensionMap, diagnostics := types.MapValue(types.StringType, extensionValues)
		resp.Diagnostics.Append(diagnostics...)
		if resp.Diagnostics.HasError() {
			return
		}
		model.CertificateExtensions = extensionMap
	} else {
		model.CertificateExtensions = types.MapNull(types.StringType)
	}

	if cert.Attributes.ExpirationDate != nil {
		model.ExpirationDate = types.StringValue(cert.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
	}

	// Update relationships if present
	if cert.Relationships != nil && cert.Relationships.PassTypeId != nil && cert.Relationships.PassTypeId.Data != nil {
		relationshipsMap := map[string]attr.Value{
			"pass_type_id": types.StringValue(cert.Relationships.PassTypeId.Data.ID),
		}
		relationshipsObj, diagnostics := types.ObjectValue(map[string]attr.Type{
			"pass_type_id": types.StringType,
		}, relationshipsMap)
		resp.Diagnostics.Append(diagnostics...)
		model.Relationships = relationshipsObj
	} else {
		// Set empty relationships object
		relationshipsMap := map[string]attr.Value{
			"pass_type_id": types.StringNull(),
		}
		relationshipsObj, diagnostics := types.ObjectValue(map[string]attr.Type{
			"pass_type_id": types.StringType,
		}, relationshipsMap)
		resp.Diagnostics.Append(diagnostics...)
		model.Relationships = relationshipsObj
	}
}
