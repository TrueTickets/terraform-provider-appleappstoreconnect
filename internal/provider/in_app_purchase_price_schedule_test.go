// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestBuildIAPPriceScheduleCreateRequest verifies the JSON:API "included"
// payload: each manual price must be inlined and referenced from
// data.relationships.manualPrices by a matching synthetic ID.
func TestBuildIAPPriceScheduleCreateRequest(t *testing.T) {
	req := buildIAPPriceScheduleCreateRequest("IAP123", "USA", []manualPrice{
		{PricePointID: "PP_USA", Territory: "USA", StartDate: ""},
		{PricePointID: "PP_FRA", Territory: "FRA", StartDate: "2025-06-01"},
	})

	if req.Data.Type != "inAppPurchasePriceSchedules" {
		t.Errorf("data.type = %q, want inAppPurchasePriceSchedules", req.Data.Type)
	}
	if got := req.Data.Relationships.InAppPurchase.Data.ID; got != "IAP123" {
		t.Errorf("inAppPurchase id = %q, want IAP123", got)
	}
	if got := req.Data.Relationships.BaseTerritory.Data.ID; got != "USA" {
		t.Errorf("baseTerritory id = %q, want USA", got)
	}

	refs := req.Data.Relationships.ManualPrices.Data
	if len(refs) != 2 {
		t.Fatalf("manualPrices refs = %d, want 2", len(refs))
	}
	if len(req.Included) != 2 {
		t.Fatalf("included = %d, want 2", len(req.Included))
	}

	// Every manualPrices ref must resolve to an included resource with the
	// same id, and the type must be inAppPurchasePrices.
	includedByID := map[string]IAPPriceInline{}
	for _, inc := range req.Included {
		includedByID[inc.ID] = inc
	}
	for i, ref := range refs {
		wantID := fmt.Sprintf("${price-%d}", i)
		if ref.ID != wantID {
			t.Errorf("ref[%d].id = %q, want %q", i, ref.ID, wantID)
		}
		if ref.Type != "inAppPurchasePrices" {
			t.Errorf("ref[%d].type = %q, want inAppPurchasePrices", i, ref.Type)
		}
		inc, ok := includedByID[ref.ID]
		if !ok {
			t.Errorf("no included resource for ref %q", ref.ID)
			continue
		}
		if inc.Relationships.InAppPurchasePricePoint.Data.Type != "inAppPurchasePricePoints" {
			t.Errorf("included[%q] price point type = %q", ref.ID, inc.Relationships.InAppPurchasePricePoint.Data.Type)
		}
	}

	// First price has no start date; second does.
	if includedByID["${price-0}"].Attributes.StartDate != nil {
		t.Errorf("price-0 startDate should be nil")
	}
	if sd := includedByID["${price-1}"].Attributes.StartDate; sd == nil || *sd != "2025-06-01" {
		t.Errorf("price-1 startDate = %v, want 2025-06-01", sd)
	}

	// Sanity check that it marshals.
	if _, err := json.Marshal(req); err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
}

func TestAccInAppPurchasePriceScheduleResource(t *testing.T) {
	appID := os.Getenv("APP_STORE_CONNECT_TEST_APP_ID")
	productID := fmt.Sprintf("com.truetickets.test.iapprice%d", time.Now().Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckApp(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInAppPurchasePriceScheduleResourceConfig(appID, productID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("appleappstoreconnect_in_app_purchase_price_schedule.test", "id"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_price_schedule.test", "base_territory", "USA"),
					resource.TestCheckResourceAttr("appleappstoreconnect_in_app_purchase_price_schedule.test", "manual_prices.#", "1"),
				),
			},
		},
	})
}

func testAccInAppPurchasePriceScheduleResourceConfig(appID, productID string) string {
	return fmt.Sprintf(`
resource "appleappstoreconnect_in_app_purchase" "test" {
  app_id               = %[1]q
  product_id           = %[2]q
  name                 = "Price Schedule Test"
  in_app_purchase_type = "NON_CONSUMABLE"
}

data "appleappstoreconnect_in_app_purchase_price_point" "usd_099" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  territory          = "USA"
  customer_price     = "0.99"
}

resource "appleappstoreconnect_in_app_purchase_price_schedule" "test" {
  in_app_purchase_id = appleappstoreconnect_in_app_purchase.test.id
  base_territory     = "USA"

  manual_prices = [
    {
      price_point_id = data.appleappstoreconnect_in_app_purchase_price_point.usd_099.id
      territory      = "USA"
    },
  ]
}
`, appID, productID)
}
