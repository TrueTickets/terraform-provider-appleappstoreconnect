// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// SubscriptionGroupSubmission represents a submission of a subscription group
// for App Review.
type SubscriptionGroupSubmission struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// SubscriptionGroupSubmissionCreateRequest represents the request body for
// submitting a subscription group for review.
type SubscriptionGroupSubmissionCreateRequest struct {
	Data SubscriptionGroupSubmissionCreateRequestData `json:"data"`
}

// SubscriptionGroupSubmissionCreateRequestData represents the data for creating
// a subscription group submission.
type SubscriptionGroupSubmissionCreateRequestData struct {
	Type          string                                                `json:"type"`
	Relationships SubscriptionGroupSubmissionCreateRequestRelationships `json:"relationships"`
}

// SubscriptionGroupSubmissionCreateRequestRelationships represents the
// relationships for creating a subscription group submission.
type SubscriptionGroupSubmissionCreateRequestRelationships struct {
	SubscriptionGroup RelationshipOne `json:"subscriptionGroup"`
}

// SubscriptionGroupSubmissionResponse represents the response from the
// subscription group submission API.
type SubscriptionGroupSubmissionResponse struct {
	Data  SubscriptionGroupSubmission `json:"data"`
	Links Links                       `json:"links,omitempty"`
}
