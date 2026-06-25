// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccBundleIDResource(t *testing.T) {
	testIdentifier := fmt.Sprintf("io.truetickets.test.bundle%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccBundleIDResourceConfig(testIdentifier, "Test Bundle ID", "IOS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_bundle_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("appleappstoreconnect_bundle_id.test", "name", "Test Bundle ID"),
					resource.TestCheckResourceAttr("appleappstoreconnect_bundle_id.test", "platform", "IOS"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_bundle_id.test", "id"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_bundle_id.test", "seed_id"),
				),
			},
			// Update testing (name is mutable in place).
			{
				Config: testAccBundleIDResourceConfig(testIdentifier, "Test Bundle ID Renamed", "IOS"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_bundle_id.test", "identifier", testIdentifier),
					resource.TestCheckResourceAttr("appleappstoreconnect_bundle_id.test", "name", "Test Bundle ID Renamed"),
				),
			},
			// ImportState testing.
			{
				ResourceName:      "appleappstoreconnect_bundle_id.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccBundleIDResourceConfig(identifier, name, platform string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_bundle_id" "test" {
  identifier = %[1]q
  name       = %[2]q
  platform   = %[3]q
}
`, identifier, name, platform)
}

func TestIsValidBundleIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{
			name:       "valid reverse-DNS identifier",
			identifier: "io.truetickets.test.app",
			want:       true,
		},
		{
			name:       "valid two-segment identifier",
			identifier: "com.example",
			want:       true,
		},
		{
			name:       "valid identifier with dashes",
			identifier: "com.my-company.app",
			want:       true,
		},
		{
			name:       "valid wildcard identifier",
			identifier: "com.example.*",
			want:       true,
		},
		{
			name:       "invalid - single segment",
			identifier: "com",
			want:       false,
		},
		{
			name:       "invalid - empty",
			identifier: "",
			want:       false,
		},
		{
			name:       "invalid - leading dot",
			identifier: ".com.example",
			want:       false,
		},
		{
			name:       "invalid - trailing dot",
			identifier: "com.example.",
			want:       false,
		},
		{
			name:       "invalid - dash at start of segment",
			identifier: "com.-example.app",
			want:       false,
		},
		{
			name:       "invalid - dash at end of segment",
			identifier: "com.example-.app",
			want:       false,
		},
		{
			name:       "invalid - special characters",
			identifier: "com.example.app!",
			want:       false,
		},
		{
			name:       "invalid - wildcard in middle",
			identifier: "com.*.app",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidBundleIdentifier(tt.identifier); got != tt.want {
				t.Errorf("isValidBundleIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}
