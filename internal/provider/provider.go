// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure AppleAppStoreConnectProvider satisfies various provider interfaces.
var _ provider.Provider = &AppleAppStoreConnectProvider{}
var _ provider.ProviderWithFunctions = &AppleAppStoreConnectProvider{}

// AppleAppStoreConnectProvider defines the provider implementation.
type AppleAppStoreConnectProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AppleAppStoreConnectProviderModel describes the provider data model.
type AppleAppStoreConnectProviderModel struct {
	IssuerID   types.String `tfsdk:"issuer_id"`
	KeyID      types.String `tfsdk:"key_id"`
	PrivateKey types.String `tfsdk:"private_key"`
}

func (p *AppleAppStoreConnectProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "appleappstoreconnect"
	resp.Version = p.version
}

func (p *AppleAppStoreConnectProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Apple App Store Connect provider allows you to manage App Store Connect resources such as Pass Type IDs and Certificates.",
		Attributes: map[string]schema.Attribute{
			"issuer_id": schema.StringAttribute{
				MarkdownDescription: "The issuer ID from the API keys page in App Store Connect. Can also be set via the `APP_STORE_CONNECT_ISSUER_ID` environment variable.",
				Optional:            true,
			},
			"key_id": schema.StringAttribute{
				MarkdownDescription: "The key ID from the API keys page in App Store Connect. Can also be set via the `APP_STORE_CONNECT_KEY_ID` environment variable.",
				Optional:            true,
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "The private key contents (.p8 file) for App Store Connect API authentication. Can also be set via the `APP_STORE_CONNECT_PRIVATE_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AppleAppStoreConnectProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Apple App Store Connect provider")

	var data AppleAppStoreConnectProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Check if any configuration values are unknown
	if data.IssuerID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("issuer_id"),
			"Unknown Apple App Store Connect Issuer ID",
			"The provider cannot create the Apple App Store Connect API client as there is an unknown configuration value for the issuer ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the APP_STORE_CONNECT_ISSUER_ID environment variable.",
		)
	}

	if data.KeyID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("key_id"),
			"Unknown Apple App Store Connect Key ID",
			"The provider cannot create the Apple App Store Connect API client as there is an unknown configuration value for the key ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the APP_STORE_CONNECT_KEY_ID environment variable.",
		)
	}

	if data.PrivateKey.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key"),
			"Unknown Apple App Store Connect Private Key",
			"The provider cannot create the Apple App Store Connect API client as there is an unknown configuration value for the private key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the APP_STORE_CONNECT_PRIVATE_KEY environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Set values from environment variables if not set in configuration
	issuerID := os.Getenv("APP_STORE_CONNECT_ISSUER_ID")
	keyID := os.Getenv("APP_STORE_CONNECT_KEY_ID")
	privateKey := os.Getenv("APP_STORE_CONNECT_PRIVATE_KEY")

	if !data.IssuerID.IsNull() {
		issuerID = data.IssuerID.ValueString()
	}

	if !data.KeyID.IsNull() {
		keyID = data.KeyID.ValueString()
	}

	if !data.PrivateKey.IsNull() {
		privateKey = data.PrivateKey.ValueString()
	}

	// Validate required fields
	if issuerID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("issuer_id"),
			"Missing Apple App Store Connect Issuer ID",
			"The provider cannot create the Apple App Store Connect API client as there is a missing or empty value for the issuer ID. "+
				"Set the issuer_id value in the configuration or use the APP_STORE_CONNECT_ISSUER_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if keyID == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("key_id"),
			"Missing Apple App Store Connect Key ID",
			"The provider cannot create the Apple App Store Connect API client as there is a missing or empty value for the key ID. "+
				"Set the key_id value in the configuration or use the APP_STORE_CONNECT_KEY_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if privateKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("private_key"),
			"Missing Apple App Store Connect Private Key",
			"The provider cannot create the Apple App Store Connect API client as there is a missing or empty value for the private key. "+
				"Set the private_key value in the configuration or use the APP_STORE_CONNECT_PRIVATE_KEY environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Apple App Store Connect API client", map[string]interface{}{
		"issuer_id": issuerID,
		"key_id":    keyID,
	})

	// Create API client
	client, err := NewClient(issuerID, keyID, privateKey)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Apple App Store Connect API Client",
			fmt.Sprintf("An unexpected error occurred when creating the Apple App Store Connect API client: %s", err.Error()),
		)
		return
	}

	// Make the client available for DataSources and Resources
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Apple App Store Connect provider", map[string]interface{}{
		"issuer_id": issuerID,
		"key_id":    keyID,
	})
}

func (p *AppleAppStoreConnectProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPassTypeIDResource,
		NewCertificateResource,
	}
}

func (p *AppleAppStoreConnectProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPassTypeIDDataSource,
		NewCertificateDataSource,
		NewCertificatesDataSource,
	}
}

func (p *AppleAppStoreConnectProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewEncodePKCS12Function,
		NewDecodePKCS12Function,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AppleAppStoreConnectProvider{
			version: version,
		}
	}
}
