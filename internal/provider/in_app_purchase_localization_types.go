// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// InAppPurchaseLocalization represents a localized name/description of an
// In-App Purchase in the App Store Connect API.
type InAppPurchaseLocalization struct {
	Type          string                                  `json:"type"`
	ID            string                                  `json:"id"`
	Attributes    InAppPurchaseLocalizationAttributes     `json:"attributes"`
	Relationships *InAppPurchaseLocalizationRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks                           `json:"links,omitempty"`
}

// InAppPurchaseLocalizationAttributes represents the attributes of an In-App
// Purchase localization.
type InAppPurchaseLocalizationAttributes struct {
	Locale      string `json:"locale,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// InAppPurchaseLocalizationRelationships represents the relationships of an
// In-App Purchase localization.
type InAppPurchaseLocalizationRelationships struct {
	InAppPurchase *Relationship `json:"inAppPurchaseV2,omitempty"`
}

// InAppPurchaseLocalizationCreateRequest represents the request body for
// creating an In-App Purchase localization.
type InAppPurchaseLocalizationCreateRequest struct {
	Data InAppPurchaseLocalizationCreateRequestData `json:"data"`
}

// InAppPurchaseLocalizationCreateRequestData represents the data for creating
// an In-App Purchase localization.
type InAppPurchaseLocalizationCreateRequestData struct {
	Type          string                                              `json:"type"`
	Attributes    InAppPurchaseLocalizationCreateRequestAttributes    `json:"attributes"`
	Relationships InAppPurchaseLocalizationCreateRequestRelationships `json:"relationships"`
}

// InAppPurchaseLocalizationCreateRequestAttributes represents the attributes
// for creating an In-App Purchase localization.
type InAppPurchaseLocalizationCreateRequestAttributes struct {
	Locale      string  `json:"locale"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// InAppPurchaseLocalizationCreateRequestRelationships represents the
// relationships for creating an In-App Purchase localization.
type InAppPurchaseLocalizationCreateRequestRelationships struct {
	InAppPurchase RelationshipOne `json:"inAppPurchaseV2"`
}

// InAppPurchaseLocalizationUpdateRequest represents the request body for
// updating an In-App Purchase localization.
type InAppPurchaseLocalizationUpdateRequest struct {
	Data InAppPurchaseLocalizationUpdateRequestData `json:"data"`
}

// InAppPurchaseLocalizationUpdateRequestData represents the data for updating
// an In-App Purchase localization.
type InAppPurchaseLocalizationUpdateRequestData struct {
	Type       string                                           `json:"type"`
	ID         string                                           `json:"id"`
	Attributes InAppPurchaseLocalizationUpdateRequestAttributes `json:"attributes"`
}

// InAppPurchaseLocalizationUpdateRequestAttributes represents the mutable
// attributes of an In-App Purchase localization.
type InAppPurchaseLocalizationUpdateRequestAttributes struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// InAppPurchaseLocalizationResponse represents the response from the In-App
// Purchase localization API.
type InAppPurchaseLocalizationResponse struct {
	Data  InAppPurchaseLocalization `json:"data"`
	Links Links                     `json:"links,omitempty"`
}
