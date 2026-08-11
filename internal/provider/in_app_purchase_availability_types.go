// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// IAPAvailability represents the territory availability of an In-App Purchase.
type IAPAvailability struct {
	Type          string                        `json:"type"`
	ID            string                        `json:"id"`
	Attributes    IAPAvailabilityAttributes     `json:"attributes"`
	Relationships *IAPAvailabilityRelationships `json:"relationships,omitempty"`
}

// IAPAvailabilityAttributes represents the attributes of an In-App Purchase
// availability.
type IAPAvailabilityAttributes struct {
	AvailableInNewTerritories bool `json:"availableInNewTerritories"`
}

// IAPAvailabilityRelationships represents the relationships of an In-App
// Purchase availability.
type IAPAvailabilityRelationships struct {
	InAppPurchase *Relationship `json:"inAppPurchase,omitempty"`
}

// IAPAvailabilityCreateRequest represents the request body for creating an
// In-App Purchase availability.
type IAPAvailabilityCreateRequest struct {
	Data IAPAvailabilityCreateRequestData `json:"data"`
}

// IAPAvailabilityCreateRequestData represents the data for creating an In-App
// Purchase availability.
type IAPAvailabilityCreateRequestData struct {
	Type          string                             `json:"type"`
	Attributes    IAPAvailabilityCreateAttributes    `json:"attributes"`
	Relationships IAPAvailabilityCreateRelationships `json:"relationships"`
}

// IAPAvailabilityCreateAttributes represents the attributes for creating an
// In-App Purchase availability.
type IAPAvailabilityCreateAttributes struct {
	AvailableInNewTerritories bool `json:"availableInNewTerritories"`
}

// IAPAvailabilityCreateRelationships represents the relationships for creating
// an In-App Purchase availability.
type IAPAvailabilityCreateRelationships struct {
	InAppPurchase        RelationshipOne  `json:"inAppPurchase"`
	AvailableTerritories RelationshipMany `json:"availableTerritories"`
}

// IAPAvailabilityResponse represents the response from the In-App Purchase
// availability API.
type IAPAvailabilityResponse struct {
	Data  IAPAvailability `json:"data"`
	Links Links           `json:"links,omitempty"`
}
