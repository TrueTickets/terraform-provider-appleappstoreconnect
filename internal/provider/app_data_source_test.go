// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccPreCheckApp ensures the env vars needed to read a real app are set.
// Apps cannot be created via the API, so acceptance tests reference an
// existing app identified by these variables.
//
//nolint:unused // This is used in acceptance tests
func testAccPreCheckApp(t *testing.T) {
	testAccPreCheck(t)

	if v := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID"); v == "" {
		t.Fatal("APP_STORE_CONNECT_TEST_APP_ID must be set for app acceptance tests")
	}

	if v := os.Getenv("APP_STORE_CONNECT_TEST_BUNDLE_ID"); v == "" {
		t.Fatal("APP_STORE_CONNECT_TEST_BUNDLE_ID must be set for app acceptance tests")
	}
}

func TestAccAppDataSource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	bundleID := os.Getenv("APP_STORE_CONNECT_TEST_BUNDLE_ID")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read by ID.
			{
				Config: testAccAppDataSourceConfigByID(appID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.appleappstoreconnect_app.test", "id", appID),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_app.test", "bundle_id", bundleID),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_app.test", "name"),
				),
			},
			// Read by bundle ID.
			{
				Config: testAccAppDataSourceConfigByBundleID(bundleID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.appleappstoreconnect_app.test", "id", appID),
					resource.TestCheckResourceAttr("data.appleappstoreconnect_app.test", "bundle_id", bundleID),
					resource.TestCheckResourceAttrSet("data.appleappstoreconnect_app.test", "name"),
				),
			},
		},
	})
}

func testAccAppDataSourceConfigByID(appID string) string {
	return fmt.Sprintf(`
data "appleappstoreconnect_app" "test" {
  id = %[1]q
}
`, appID)
}

func testAccAppDataSourceConfigByBundleID(bundleID string) string {
	return fmt.Sprintf(`
data "appleappstoreconnect_app" "test" {
  bundle_id = %[1]q
}
`, bundleID)
}
