// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// SubscriptionGroup represents a subscription group in the App Store Connect
// API. A subscription group organizes related auto-renewable subscriptions for
// an app.
type SubscriptionGroup struct {
	Type          string                          `json:"type"`
	ID            string                          `json:"id"`
	Attributes    SubscriptionGroupAttributes     `json:"attributes"`
	Relationships *SubscriptionGroupRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks                   `json:"links,omitempty"`
}

// SubscriptionGroupAttributes represents the attributes of a subscription
// group.
type SubscriptionGroupAttributes struct {
	ReferenceName string `json:"referenceName,omitempty"`
}

// SubscriptionGroupRelationships represents the relationships of a subscription
// group.
type SubscriptionGroupRelationships struct {
	App *Relationship `json:"app,omitempty"`
}

// SubscriptionGroupCreateRequest represents the request body for creating a
// subscription group.
type SubscriptionGroupCreateRequest struct {
	Data SubscriptionGroupCreateRequestData `json:"data"`
}

// SubscriptionGroupCreateRequestData represents the data for creating a
// subscription group.
type SubscriptionGroupCreateRequestData struct {
	Type          string                                      `json:"type"`
	Attributes    SubscriptionGroupCreateRequestAttributes    `json:"attributes"`
	Relationships SubscriptionGroupCreateRequestRelationships `json:"relationships"`
}

// SubscriptionGroupCreateRequestAttributes represents the attributes for
// creating a subscription group.
type SubscriptionGroupCreateRequestAttributes struct {
	ReferenceName string `json:"referenceName"`
}

// SubscriptionGroupCreateRequestRelationships represents the relationships for
// creating a subscription group.
type SubscriptionGroupCreateRequestRelationships struct {
	App RelationshipOne `json:"app"`
}

// SubscriptionGroupUpdateRequest represents the request body for updating a
// subscription group.
type SubscriptionGroupUpdateRequest struct {
	Data SubscriptionGroupUpdateRequestData `json:"data"`
}

// SubscriptionGroupUpdateRequestData represents the data for updating a
// subscription group.
type SubscriptionGroupUpdateRequestData struct {
	Type       string                                   `json:"type"`
	ID         string                                   `json:"id"`
	Attributes SubscriptionGroupUpdateRequestAttributes `json:"attributes"`
}

// SubscriptionGroupUpdateRequestAttributes represents the mutable attributes of
// a subscription group.
type SubscriptionGroupUpdateRequestAttributes struct {
	ReferenceName *string `json:"referenceName,omitempty"`
}

// SubscriptionGroupResponse represents the response from the subscription group
// API.
type SubscriptionGroupResponse struct {
	Data  SubscriptionGroup `json:"data"`
	Links Links             `json:"links,omitempty"`
}
