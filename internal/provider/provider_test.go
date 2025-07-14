// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
//
//nolint:unused // This is used in acceptance tests
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"appleappstoreconnect": providerserver.NewProtocol6WithError(New("test")()),
}

//nolint:unused // This is used in acceptance tests
func testAccPreCheck(t *testing.T) {
	// Check for required environment variables
	if v := os.Getenv("APP_STORE_CONNECT_ISSUER_ID"); v == "" {
		t.Fatal("APP_STORE_CONNECT_ISSUER_ID must be set for acceptance tests")
	}

	if v := os.Getenv("APP_STORE_CONNECT_KEY_ID"); v == "" {
		t.Fatal("APP_STORE_CONNECT_KEY_ID must be set for acceptance tests")
	}

	if v := os.Getenv("APP_STORE_CONNECT_PRIVATE_KEY"); v == "" {
		t.Fatal("APP_STORE_CONNECT_PRIVATE_KEY must be set for acceptance tests")
	}
}

func TestProvider(t *testing.T) {
	// Simply test that the provider can be created
	p := New("test")()
	if p == nil {
		t.Fatal("provider should not be nil")
	}
}

func TestProviderMetadata(t *testing.T) {
	ctx := context.Background()
	p := &AppleAppStoreConnectProvider{version: "test"}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(ctx, req, resp)

	if resp.TypeName != "appleappstoreconnect" {
		t.Errorf("Expected TypeName 'appleappstoreconnect', got %s", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Errorf("Expected Version 'test', got %s", resp.Version)
	}
}

func TestProviderResources(t *testing.T) {
	ctx := context.Background()
	p := &AppleAppStoreConnectProvider{}

	resources := p.Resources(ctx)

	if len(resources) != 2 {
		t.Errorf("Expected 2 resources, got %d", len(resources))
	}
}

func TestProviderDataSources(t *testing.T) {
	ctx := context.Background()
	p := &AppleAppStoreConnectProvider{}

	dataSources := p.DataSources(ctx)

	if len(dataSources) != 3 {
		t.Errorf("Expected 3 data sources, got %d", len(dataSources))
	}
}
