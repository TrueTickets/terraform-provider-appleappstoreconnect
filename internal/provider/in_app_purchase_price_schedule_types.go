// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "fmt"

// manualPrice is the resolved input for a single territory price within a
// schedule, decoupled from the Terraform types so the request builder is
// independently testable.
type manualPrice struct {
	PricePointID string
	Territory    string
	StartDate    string // empty means "effective immediately"
}

// IAPPriceScheduleCreateRequest represents the request body for creating an
// In-App Purchase price schedule. It uses the JSON:API "included" pattern: the
// manual prices are inlined under `included` and referenced from
// `data.relationships.manualPrices` by client-assigned synthetic IDs.
type IAPPriceScheduleCreateRequest struct {
	Data     IAPPriceScheduleCreateRequestData `json:"data"`
	Included []IAPPriceInline                  `json:"included"`
}

// IAPPriceScheduleCreateRequestData represents the data for creating a price
// schedule.
type IAPPriceScheduleCreateRequestData struct {
	Type          string                              `json:"type"`
	Relationships IAPPriceScheduleCreateRelationships `json:"relationships"`
}

// IAPPriceScheduleCreateRelationships represents the relationships of a price
// schedule create request.
type IAPPriceScheduleCreateRelationships struct {
	InAppPurchase RelationshipOne  `json:"inAppPurchase"`
	BaseTerritory RelationshipOne  `json:"baseTerritory"`
	ManualPrices  RelationshipMany `json:"manualPrices"`
}

// IAPPriceInline is an inlined inAppPurchasePrices resource in the `included`
// array of a price schedule create request.
type IAPPriceInline struct {
	Type          string                      `json:"type"`
	ID            string                      `json:"id"`
	Attributes    IAPPriceInlineAttributes    `json:"attributes"`
	Relationships IAPPriceInlineRelationships `json:"relationships"`
}

// IAPPriceInlineAttributes represents the attributes of an inlined price.
type IAPPriceInlineAttributes struct {
	StartDate *string `json:"startDate,omitempty"`
}

// IAPPriceInlineRelationships represents the relationships of an inlined price.
type IAPPriceInlineRelationships struct {
	InAppPurchasePricePoint RelationshipOne `json:"inAppPurchasePricePoint"`
	Territory               RelationshipOne `json:"territory"`
}

// IAPPriceSchedule represents an In-App Purchase price schedule resource.
type IAPPriceSchedule struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// IAPPriceScheduleResponse represents the response from the price schedule API.
type IAPPriceScheduleResponse struct {
	Data  IAPPriceSchedule `json:"data"`
	Links Links            `json:"links,omitempty"`
}

// buildIAPPriceScheduleCreateRequest assembles the create request body,
// linking each manual price to an inlined inAppPurchasePrices resource via a
// synthetic local-id reference. App Store Connect requires inline IDs to be
// wrapped in ${...} (e.g. "${price-0}", "${price-1}").
func buildIAPPriceScheduleCreateRequest(inAppPurchaseID, baseTerritory string, prices []manualPrice) IAPPriceScheduleCreateRequest {
	refs := make([]RelationshipData, 0, len(prices))
	included := make([]IAPPriceInline, 0, len(prices))

	for i, p := range prices {
		ref := fmt.Sprintf("${price-%d}", i)
		refs = append(refs, RelationshipData{Type: "inAppPurchasePrices", ID: ref})

		inline := IAPPriceInline{
			Type: "inAppPurchasePrices",
			ID:   ref,
			Relationships: IAPPriceInlineRelationships{
				InAppPurchasePricePoint: RelationshipOne{
					Data: RelationshipData{Type: "inAppPurchasePricePoints", ID: p.PricePointID},
				},
				Territory: RelationshipOne{
					Data: RelationshipData{Type: "territories", ID: p.Territory},
				},
			},
		}
		if p.StartDate != "" {
			startDate := p.StartDate
			inline.Attributes.StartDate = &startDate
		}
		included = append(included, inline)
	}

	return IAPPriceScheduleCreateRequest{
		Data: IAPPriceScheduleCreateRequestData{
			Type: "inAppPurchasePriceSchedules",
			Relationships: IAPPriceScheduleCreateRelationships{
				InAppPurchase: RelationshipOne{
					Data: RelationshipData{Type: "inAppPurchases", ID: inAppPurchaseID},
				},
				BaseTerritory: RelationshipOne{
					Data: RelationshipData{Type: "territories", ID: baseTerritory},
				},
				ManualPrices: RelationshipMany{Data: refs},
			},
		},
		Included: included,
	}
}
