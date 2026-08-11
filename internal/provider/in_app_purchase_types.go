// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// In-App Purchase types.
const (
	InAppPurchaseTypeConsumable              = "CONSUMABLE"
	InAppPurchaseTypeNonConsumable           = "NON_CONSUMABLE"
	InAppPurchaseTypeNonRenewingSubscription = "NON_RENEWING_SUBSCRIPTION"
)

// InAppPurchaseV2 represents an In-App Purchase in the App Store Connect API
// (the v2 `inAppPurchases` resource).
type InAppPurchaseV2 struct {
	Type          string                        `json:"type"`
	ID            string                        `json:"id"`
	Attributes    InAppPurchaseV2Attributes     `json:"attributes"`
	Relationships *InAppPurchaseV2Relationships `json:"relationships,omitempty"`
	Links         ResourceLinks                 `json:"links,omitempty"`
}

// InAppPurchaseV2Attributes represents the attributes of an In-App Purchase.
type InAppPurchaseV2Attributes struct {
	Name              string `json:"name,omitempty"`
	ProductID         string `json:"productId,omitempty"`
	InAppPurchaseType string `json:"inAppPurchaseType,omitempty"`
	State             string `json:"state,omitempty"`
	ReviewNote        string `json:"reviewNote,omitempty"`
	FamilySharable    bool   `json:"familySharable,omitempty"`
}

// InAppPurchaseV2Relationships represents the relationships of an In-App
// Purchase that this provider reads back.
type InAppPurchaseV2Relationships struct {
	App *Relationship `json:"app,omitempty"`
}

// InAppPurchaseCreateRequest represents the request body for creating an
// In-App Purchase.
type InAppPurchaseCreateRequest struct {
	Data InAppPurchaseCreateRequestData `json:"data"`
}

// InAppPurchaseCreateRequestData represents the data for creating an In-App
// Purchase.
type InAppPurchaseCreateRequestData struct {
	Type          string                                  `json:"type"`
	Attributes    InAppPurchaseCreateRequestAttributes    `json:"attributes"`
	Relationships InAppPurchaseCreateRequestRelationships `json:"relationships"`
}

// InAppPurchaseCreateRequestAttributes represents the attributes for creating
// an In-App Purchase. Booleans/strings that are optional use pointers so a
// false/empty value is only sent when explicitly configured.
type InAppPurchaseCreateRequestAttributes struct {
	Name              string  `json:"name"`
	ProductID         string  `json:"productId"`
	InAppPurchaseType string  `json:"inAppPurchaseType"`
	FamilySharable    *bool   `json:"familySharable,omitempty"`
	ReviewNote        *string `json:"reviewNote,omitempty"`
}

// InAppPurchaseCreateRequestRelationships represents the relationships for
// creating an In-App Purchase.
type InAppPurchaseCreateRequestRelationships struct {
	App RelationshipOne `json:"app"`
}

// InAppPurchaseUpdateRequest represents the request body for updating an
// In-App Purchase.
type InAppPurchaseUpdateRequest struct {
	Data InAppPurchaseUpdateRequestData `json:"data"`
}

// InAppPurchaseUpdateRequestData represents the data for updating an In-App
// Purchase.
type InAppPurchaseUpdateRequestData struct {
	Type       string                               `json:"type"`
	ID         string                               `json:"id"`
	Attributes InAppPurchaseUpdateRequestAttributes `json:"attributes"`
}

// InAppPurchaseUpdateRequestAttributes represents the mutable attributes of an
// In-App Purchase.
type InAppPurchaseUpdateRequestAttributes struct {
	Name           *string `json:"name,omitempty"`
	FamilySharable *bool   `json:"familySharable,omitempty"`
	ReviewNote     *string `json:"reviewNote,omitempty"`
}

// InAppPurchaseV2Response represents the response from the In-App Purchase API.
type InAppPurchaseV2Response struct {
	Data  InAppPurchaseV2 `json:"data"`
	Links Links           `json:"links,omitempty"`
}
