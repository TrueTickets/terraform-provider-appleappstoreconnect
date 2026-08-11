// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// SubscriptionPrice represents a per-territory price of a subscription in the
// App Store Connect API.
type SubscriptionPrice struct {
	Type          string                          `json:"type"`
	ID            string                          `json:"id"`
	Attributes    SubscriptionPriceAttributes     `json:"attributes"`
	Relationships *SubscriptionPriceRelationships `json:"relationships,omitempty"`
}

// SubscriptionPriceAttributes represents the attributes of a subscription
// price.
type SubscriptionPriceAttributes struct {
	StartDate            string `json:"startDate,omitempty"`
	PreserveCurrentPrice bool   `json:"preserveCurrentPrice,omitempty"`
}

// SubscriptionPriceRelationships represents the relationships of a subscription
// price.
type SubscriptionPriceRelationships struct {
	SubscriptionPricePoint *Relationship `json:"subscriptionPricePoint,omitempty"`
	Territory              *Relationship `json:"territory,omitempty"`
}

// SubscriptionPatchWithInlinePriceRequest mirrors PATCH /v1/subscriptions/{id}
// with an inline subscriptionPrices resource. App Store Connect's
// POST /v1/subscriptionPrices endpoint is reserved for *changes* to an existing
// price — the initial price of a subscription must be created inline via this
// PATCH using the ${local-id} convention (same pattern as IAP price schedule).
type SubscriptionPatchWithInlinePriceRequest struct {
	Data     SubscriptionPatchData     `json:"data"`
	Included []SubscriptionPriceInline `json:"included"`
}

// SubscriptionPatchData is the top-level "data" of the PATCH body.
type SubscriptionPatchData struct {
	Type          string                         `json:"type"`
	ID            string                         `json:"id"`
	Relationships SubscriptionPatchRelationships `json:"relationships"`
}

// SubscriptionPatchRelationships references the inlined prices by local-id.
type SubscriptionPatchRelationships struct {
	Prices SubscriptionPatchPricesRelationship `json:"prices"`
}

// SubscriptionPatchPricesRelationship lists the inlined price local-ids.
type SubscriptionPatchPricesRelationship struct {
	Data []RelationshipData `json:"data"`
}

// SubscriptionPriceInline is an inlined subscriptionPrices resource in the
// PATCH's "included" array. Its ID is a synthetic local-id wrapped in ${...}.
type SubscriptionPriceInline struct {
	Type          string                               `json:"type"`
	ID            string                               `json:"id"`
	Attributes    *SubscriptionPriceInlineAttributes   `json:"attributes,omitempty"`
	Relationships SubscriptionPriceInlineRelationships `json:"relationships"`
}

// SubscriptionPriceInlineAttributes carries the optional price attributes.
type SubscriptionPriceInlineAttributes struct {
	StartDate            *string `json:"startDate,omitempty"`
	PreserveCurrentPrice *bool   `json:"preserveCurrentPrice,omitempty"`
}

// SubscriptionPriceInlineRelationships carries the required relationships for
// the inlined price. `territory` is optional because the territory is encoded
// in the subscriptionPricePoint ID.
type SubscriptionPriceInlineRelationships struct {
	SubscriptionPricePoint RelationshipOne  `json:"subscriptionPricePoint"`
	Territory              *RelationshipOne `json:"territory,omitempty"`
}

// SubscriptionPriceResponse represents the response from the subscription price
// API.
type SubscriptionPriceResponse struct {
	Data  SubscriptionPrice `json:"data"`
	Links Links             `json:"links,omitempty"`
}
