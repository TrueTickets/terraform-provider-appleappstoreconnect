// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// IAPSubmission represents an In-App Purchase submission for App Review.
type IAPSubmission struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// IAPSubmissionCreateRequest represents the request body for submitting an
// In-App Purchase for review.
type IAPSubmissionCreateRequest struct {
	Data IAPSubmissionCreateRequestData `json:"data"`
}

// IAPSubmissionCreateRequestData represents the data for creating an In-App
// Purchase submission.
type IAPSubmissionCreateRequestData struct {
	Type          string                           `json:"type"`
	Relationships IAPSubmissionCreateRelationships `json:"relationships"`
}

// IAPSubmissionCreateRelationships represents the relationships for creating an
// In-App Purchase submission.
type IAPSubmissionCreateRelationships struct {
	InAppPurchase RelationshipOne `json:"inAppPurchase"`
}

// IAPSubmissionResponse represents the response from the In-App Purchase
// submission API.
type IAPSubmissionResponse struct {
	Data  IAPSubmission `json:"data"`
	Links Links         `json:"links,omitempty"`
}
