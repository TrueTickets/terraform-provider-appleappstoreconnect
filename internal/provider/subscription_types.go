// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// Subscription periods accepted by the App Store Connect API.
const (
	SubscriptionPeriodOneWeek     = "ONE_WEEK"
	SubscriptionPeriodOneMonth    = "ONE_MONTH"
	SubscriptionPeriodTwoMonths   = "TWO_MONTHS"
	SubscriptionPeriodThreeMonths = "THREE_MONTHS"
	SubscriptionPeriodSixMonths   = "SIX_MONTHS"
	SubscriptionPeriodOneYear     = "ONE_YEAR"
)

// Subscription represents an auto-renewable subscription in the App Store
// Connect API.
type Subscription struct {
	Type          string                     `json:"type"`
	ID            string                     `json:"id"`
	Attributes    SubscriptionAttributes     `json:"attributes"`
	Relationships *SubscriptionRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks              `json:"links,omitempty"`
}

// SubscriptionAttributes represents the attributes of a subscription.
type SubscriptionAttributes struct {
	Name               string `json:"name,omitempty"`
	ProductID          string `json:"productId,omitempty"`
	SubscriptionPeriod string `json:"subscriptionPeriod,omitempty"`
	FamilySharable     bool   `json:"familySharable,omitempty"`
	GroupLevel         int64  `json:"groupLevel,omitempty"`
	ReviewNote         string `json:"reviewNote,omitempty"`
	State              string `json:"state,omitempty"`
}

// SubscriptionRelationships represents the relationships of a subscription.
type SubscriptionRelationships struct {
	Group *Relationship `json:"group,omitempty"`
}

// SubscriptionCreateRequest represents the request body for creating a
// subscription.
type SubscriptionCreateRequest struct {
	Data SubscriptionCreateRequestData `json:"data"`
}

// SubscriptionCreateRequestData represents the data for creating a
// subscription.
type SubscriptionCreateRequestData struct {
	Type          string                                 `json:"type"`
	Attributes    SubscriptionCreateRequestAttributes    `json:"attributes"`
	Relationships SubscriptionCreateRequestRelationships `json:"relationships"`
}

// SubscriptionCreateRequestAttributes represents the attributes for creating a
// subscription.
type SubscriptionCreateRequestAttributes struct {
	Name               string  `json:"name"`
	ProductID          string  `json:"productId"`
	SubscriptionPeriod string  `json:"subscriptionPeriod"`
	FamilySharable     *bool   `json:"familySharable,omitempty"`
	GroupLevel         *int64  `json:"groupLevel,omitempty"`
	ReviewNote         *string `json:"reviewNote,omitempty"`
}

// SubscriptionCreateRequestRelationships represents the relationships for
// creating a subscription.
type SubscriptionCreateRequestRelationships struct {
	Group RelationshipOne `json:"group"`
}

// SubscriptionUpdateRequest represents the request body for updating a
// subscription.
type SubscriptionUpdateRequest struct {
	Data SubscriptionUpdateRequestData `json:"data"`
}

// SubscriptionUpdateRequestData represents the data for updating a
// subscription.
type SubscriptionUpdateRequestData struct {
	Type       string                              `json:"type"`
	ID         string                              `json:"id"`
	Attributes SubscriptionUpdateRequestAttributes `json:"attributes"`
}

// SubscriptionUpdateRequestAttributes represents the mutable attributes of a
// subscription.
type SubscriptionUpdateRequestAttributes struct {
	Name       *string `json:"name,omitempty"`
	ReviewNote *string `json:"reviewNote,omitempty"`
}

// SubscriptionResponse represents the response from the subscription API.
type SubscriptionResponse struct {
	Data  Subscription `json:"data"`
	Links Links        `json:"links,omitempty"`
}
