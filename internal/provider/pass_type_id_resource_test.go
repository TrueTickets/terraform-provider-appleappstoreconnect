// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPassTypeIDResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPassTypeIDResourceConfig("pass.com.example.test", "Test Pass Type"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appleappstoreconnect_pass_type_id.test", "identifier", "pass.com.example.test"),
					resource.TestCheckResourceAttr("appleappstoreconnect_pass_type_id.test", "description", "Test Pass Type"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_pass_type_id.test", "id"),
					resource.TestCheckResourceAttrSet("appleappstoreconnect_pass_type_id.test", "created_date"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "appleappstoreconnect_pass_type_id.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccPassTypeIDResourceConfig(identifier, description string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_pass_type_id" "test" {
  identifier  = %[1]q
  description = %[2]q
}
`, identifier, description)
}

func TestIsValidPassTypeIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		want       bool
	}{
		{
			name:       "valid pass type identifier",
			identifier: "pass.com.example.membership",
			want:       true,
		},
		{
			name:       "valid pass type identifier with multiple segments",
			identifier: "pass.com.example.app.membership",
			want:       true,
		},
		{
			name:       "valid pass type identifier with dashes",
			identifier: "pass.com.my-company.membership",
			want:       true,
		},
		{
			name:       "valid pass type identifier with dashes in multiple segments",
			identifier: "pass.com.my-company.mobile-app.membership",
			want:       true,
		},
		{
			name:       "invalid - missing pass prefix",
			identifier: "com.example.membership",
			want:       false,
		},
		{
			name:       "invalid - wrong prefix",
			identifier: "app.com.example.membership",
			want:       false,
		},
		{
			name:       "invalid - too few segments",
			identifier: "pass.example",
			want:       false,
		},
		{
			name:       "invalid - empty",
			identifier: "",
			want:       false,
		},
		{
			name:       "invalid - just pass",
			identifier: "pass",
			want:       false,
		},
		{
			name:       "invalid - special characters",
			identifier: "pass.com.example.membership!",
			want:       false,
		},
		{
			name:       "invalid - dash at start of segment",
			identifier: "pass.com.-example.membership",
			want:       false,
		},
		{
			name:       "invalid - dash at end of segment",
			identifier: "pass.com.example-.membership",
			want:       false,
		},
		{
			name:       "invalid - consecutive dashes",
			identifier: "pass.com.example--test.membership",
			want:       true, // consecutive dashes are actually valid in domain names
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPassTypeIdentifier(tt.identifier); got != tt.want {
				t.Errorf("isValidPassTypeIdentifier(%q) = %v, want %v", tt.identifier, got, tt.want)
			}
		})
	}
}
