// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// SubscriptionGroupLocalization represents a localized name for a subscription
// group in the App Store Connect API.
type SubscriptionGroupLocalization struct {
	Type          string                                      `json:"type"`
	ID            string                                      `json:"id"`
	Attributes    SubscriptionGroupLocalizationAttributes     `json:"attributes"`
	Relationships *SubscriptionGroupLocalizationRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks                               `json:"links,omitempty"`
}

// SubscriptionGroupLocalizationAttributes represents the attributes of a
// subscription group localization.
type SubscriptionGroupLocalizationAttributes struct {
	Locale        string `json:"locale,omitempty"`
	Name          string `json:"name,omitempty"`
	CustomAppName string `json:"customAppName,omitempty"`
}

// SubscriptionGroupLocalizationRelationships represents the relationships of a
// subscription group localization.
type SubscriptionGroupLocalizationRelationships struct {
	SubscriptionGroup *Relationship `json:"subscriptionGroup,omitempty"`
}

// SubscriptionGroupLocalizationCreateRequest represents the request body for
// creating a subscription group localization.
type SubscriptionGroupLocalizationCreateRequest struct {
	Data SubscriptionGroupLocalizationCreateRequestData `json:"data"`
}

// SubscriptionGroupLocalizationCreateRequestData represents the data for
// creating a subscription group localization.
type SubscriptionGroupLocalizationCreateRequestData struct {
	Type          string                                                  `json:"type"`
	Attributes    SubscriptionGroupLocalizationCreateRequestAttributes    `json:"attributes"`
	Relationships SubscriptionGroupLocalizationCreateRequestRelationships `json:"relationships"`
}

// SubscriptionGroupLocalizationCreateRequestAttributes represents the
// attributes for creating a subscription group localization.
type SubscriptionGroupLocalizationCreateRequestAttributes struct {
	Locale        string  `json:"locale"`
	Name          string  `json:"name"`
	CustomAppName *string `json:"customAppName,omitempty"`
}

// SubscriptionGroupLocalizationCreateRequestRelationships represents the
// relationships for creating a subscription group localization.
type SubscriptionGroupLocalizationCreateRequestRelationships struct {
	SubscriptionGroup RelationshipOne `json:"subscriptionGroup"`
}

// SubscriptionGroupLocalizationUpdateRequest represents the request body for
// updating a subscription group localization.
type SubscriptionGroupLocalizationUpdateRequest struct {
	Data SubscriptionGroupLocalizationUpdateRequestData `json:"data"`
}

// SubscriptionGroupLocalizationUpdateRequestData represents the data for
// updating a subscription group localization.
type SubscriptionGroupLocalizationUpdateRequestData struct {
	Type       string                                               `json:"type"`
	ID         string                                               `json:"id"`
	Attributes SubscriptionGroupLocalizationUpdateRequestAttributes `json:"attributes"`
}

// SubscriptionGroupLocalizationUpdateRequestAttributes represents the mutable
// attributes of a subscription group localization.
type SubscriptionGroupLocalizationUpdateRequestAttributes struct {
	Name          *string `json:"name,omitempty"`
	CustomAppName *string `json:"customAppName,omitempty"`
}

// SubscriptionGroupLocalizationResponse represents the response from the
// subscription group localization API.
type SubscriptionGroupLocalizationResponse struct {
	Data  SubscriptionGroupLocalization `json:"data"`
	Links Links                         `json:"links,omitempty"`
}
