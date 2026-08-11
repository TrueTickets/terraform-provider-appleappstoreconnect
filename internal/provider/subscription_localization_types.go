// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// SubscriptionLocalization represents a localized name/description of a
// subscription in the App Store Connect API.
type SubscriptionLocalization struct {
	Type          string                                 `json:"type"`
	ID            string                                 `json:"id"`
	Attributes    SubscriptionLocalizationAttributes     `json:"attributes"`
	Relationships *SubscriptionLocalizationRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks                          `json:"links,omitempty"`
}

// SubscriptionLocalizationAttributes represents the attributes of a
// subscription localization.
type SubscriptionLocalizationAttributes struct {
	Locale      string `json:"locale,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// SubscriptionLocalizationRelationships represents the relationships of a
// subscription localization.
type SubscriptionLocalizationRelationships struct {
	Subscription *Relationship `json:"subscription,omitempty"`
}

// SubscriptionLocalizationCreateRequest represents the request body for
// creating a subscription localization.
type SubscriptionLocalizationCreateRequest struct {
	Data SubscriptionLocalizationCreateRequestData `json:"data"`
}

// SubscriptionLocalizationCreateRequestData represents the data for creating a
// subscription localization.
type SubscriptionLocalizationCreateRequestData struct {
	Type          string                                             `json:"type"`
	Attributes    SubscriptionLocalizationCreateRequestAttributes    `json:"attributes"`
	Relationships SubscriptionLocalizationCreateRequestRelationships `json:"relationships"`
}

// SubscriptionLocalizationCreateRequestAttributes represents the attributes for
// creating a subscription localization.
type SubscriptionLocalizationCreateRequestAttributes struct {
	Locale      string  `json:"locale"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// SubscriptionLocalizationCreateRequestRelationships represents the
// relationships for creating a subscription localization.
type SubscriptionLocalizationCreateRequestRelationships struct {
	Subscription RelationshipOne `json:"subscription"`
}

// SubscriptionLocalizationUpdateRequest represents the request body for
// updating a subscription localization.
type SubscriptionLocalizationUpdateRequest struct {
	Data SubscriptionLocalizationUpdateRequestData `json:"data"`
}

// SubscriptionLocalizationUpdateRequestData represents the data for updating a
// subscription localization.
type SubscriptionLocalizationUpdateRequestData struct {
	Type       string                                          `json:"type"`
	ID         string                                          `json:"id"`
	Attributes SubscriptionLocalizationUpdateRequestAttributes `json:"attributes"`
}

// SubscriptionLocalizationUpdateRequestAttributes represents the mutable
// attributes of a subscription localization.
type SubscriptionLocalizationUpdateRequestAttributes struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SubscriptionLocalizationResponse represents the response from the
// subscription localization API.
type SubscriptionLocalizationResponse struct {
	Data  SubscriptionLocalization `json:"data"`
	Links Links                    `json:"links,omitempty"`
}
