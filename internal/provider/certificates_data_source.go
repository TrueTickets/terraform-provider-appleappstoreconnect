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
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CertificatesDataSource{}

// NewCertificatesDataSource creates a new Certificates data source.
func NewCertificatesDataSource() datasource.DataSource {
	return &CertificatesDataSource{}
}

// CertificatesDataSource defines the data source implementation.
type CertificatesDataSource struct {
	client *Client
}

// CertificatesDataSourceModel describes the data source data model.
type CertificatesDataSourceModel struct {
	Certificates types.List   `tfsdk:"certificates"`
	Filter       types.Object `tfsdk:"filter"`
}

// CertificatesFilterModel describes the filter criteria.
type CertificatesFilterModel struct {
	CertificateType types.String `tfsdk:"certificate_type"`
	DisplayName     types.String `tfsdk:"display_name"`
}

// CertificateListItemModel describes a certificate in the list.
type CertificateListItemModel struct {
	ID                 types.String `tfsdk:"id"`
	CertificateType    types.String `tfsdk:"certificate_type"`
	CertificateContent types.String `tfsdk:"certificate_content"`
	DisplayName        types.String `tfsdk:"display_name"`
	Name               types.String `tfsdk:"name"`
	Platform           types.String `tfsdk:"platform"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	Relationships      types.Object `tfsdk:"relationships"`
}

func (d *CertificatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *CertificatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve a list of Certificates from App Store Connect.",

		Attributes: map[string]schema.Attribute{
			"certificates": schema.ListNestedAttribute{
				MarkdownDescription: "List of certificates matching the filter criteria.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier of the Certificate.",
							Computed:            true,
						},
						"certificate_type": schema.StringAttribute{
							MarkdownDescription: "The type of certificate.",
							Computed:            true,
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
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"pass_type_id": schema.StringAttribute{
									MarkdownDescription: "The ID of the associated Pass Type ID.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Filter criteria for listing certificates.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"certificate_type": schema.StringAttribute{
						MarkdownDescription: "Filter by certificate type.",
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
					"display_name": schema.StringAttribute{
						MarkdownDescription: "Filter by display name (partial match).",
						Optional:            true,
					},
				},
			},
		},
	}
}

func (d *CertificatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificatesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build query parameters
	query := make(map[string]string)
	query["limit"] = "200" // Maximum allowed by API

	// Extract filter criteria if present
	if !data.Filter.IsNull() {
		var filter CertificatesFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		if !filter.CertificateType.IsNull() {
			query["filter[certificateType]"] = filter.CertificateType.ValueString()
		}

		tflog.Debug(ctx, "Fetching Certificates with filter", map[string]interface{}{
			"certificate_type": filter.CertificateType.ValueString(),
			"display_name":     filter.DisplayName.ValueString(),
		})
	} else {
		tflog.Debug(ctx, "Fetching all Certificates")
	}

	// Make the API request
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

	// Parse the response - apiResp.Data contains just the array from the "data" field
	var certificates []Certificate
	if err := json.Unmarshal(apiResp.Data, &certificates); err != nil {
		// Log the raw response for debugging
		tflog.Error(ctx, "Failed to parse certificates response", map[string]interface{}{
			"error":        err.Error(),
			"raw_response": string(apiResp.Data),
		})
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse Certificates response, got error: %s", err),
		)
		return
	}

	// Apply client-side filtering if needed
	var filteredCerts []Certificate
	if !data.Filter.IsNull() {
		var filter CertificatesFilterModel
		resp.Diagnostics.Append(data.Filter.As(ctx, &filter, basetypes.ObjectAsOptions{})...)
		if resp.Diagnostics.HasError() {
			return
		}

		for _, cert := range certificates {
			// Apply display name filter if present
			if !filter.DisplayName.IsNull() {
				displayNameFilter := filter.DisplayName.ValueString()
				if displayNameFilter != "" {
					// Simple substring match
					found := false
					if cert.Attributes.DisplayName != "" &&
						len(cert.Attributes.DisplayName) >= len(displayNameFilter) {
						for i := 0; i <= len(cert.Attributes.DisplayName)-len(displayNameFilter); i++ {
							if cert.Attributes.DisplayName[i:i+len(displayNameFilter)] == displayNameFilter {
								found = true
								break
							}
						}
					}
					if !found {
						continue
					}
				}
			}
			filteredCerts = append(filteredCerts, cert)
		}
	} else {
		filteredCerts = certificates
	}

	// Convert certificates to list items
	certItems := make([]CertificateListItemModel, 0, len(filteredCerts))
	for _, cert := range filteredCerts {
		item := CertificateListItemModel{
			ID:                 types.StringValue(cert.ID),
			CertificateType:    types.StringValue(cert.Attributes.CertificateType),
			CertificateContent: types.StringValue(cert.Attributes.CertificateContent),
			DisplayName:        types.StringValue(cert.Attributes.DisplayName),
			Name:               types.StringValue(cert.Attributes.Name),
			Platform:           types.StringValue(cert.Attributes.Platform),
			SerialNumber:       types.StringValue(cert.Attributes.SerialNumber),
		}

		if cert.Attributes.ExpirationDate != nil {
			item.ExpirationDate = types.StringValue(cert.Attributes.ExpirationDate.Format("2006-01-02T15:04:05Z"))
		} else {
			item.ExpirationDate = types.StringNull()
		}

		// Handle relationships
		if cert.Relationships != nil && cert.Relationships.PassTypeId != nil && cert.Relationships.PassTypeId.Data != nil {
			relationshipsMap := map[string]attr.Value{
				"pass_type_id": types.StringValue(cert.Relationships.PassTypeId.Data.ID),
			}
			relationshipsObj, diags := types.ObjectValue(map[string]attr.Type{
				"pass_type_id": types.StringType,
			}, relationshipsMap)
			resp.Diagnostics.Append(diags...)
			item.Relationships = relationshipsObj
		} else {
			// Set empty relationships object
			relationshipsMap := map[string]attr.Value{
				"pass_type_id": types.StringNull(),
			}
			relationshipsObj, diags := types.ObjectValue(map[string]attr.Type{
				"pass_type_id": types.StringType,
			}, relationshipsMap)
			resp.Diagnostics.Append(diags...)
			item.Relationships = relationshipsObj
		}

		certItems = append(certItems, item)
	}

	// Create the list value
	certList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":                  types.StringType,
			"certificate_type":    types.StringType,
			"certificate_content": types.StringType,
			"display_name":        types.StringType,
			"name":                types.StringType,
			"platform":            types.StringType,
			"serial_number":       types.StringType,
			"expiration_date":     types.StringType,
			"relationships": types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"pass_type_id": types.StringType,
				},
			},
		},
	}, certItems)
	resp.Diagnostics.Append(diags...)
	data.Certificates = certList

	tflog.Debug(ctx, "Found certificates", map[string]interface{}{
		"count": len(certItems),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
