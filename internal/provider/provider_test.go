// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

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
