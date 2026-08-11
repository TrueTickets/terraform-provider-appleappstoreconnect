// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// BundleID represents a Bundle ID in the App Store Connect API.
type BundleID struct {
	Type       string             `json:"type"`
	ID         string             `json:"id"`
	Attributes BundleIDAttributes `json:"attributes"`
	Links      ResourceLinks      `json:"links,omitempty"`
}

// BundleIDAttributes represents the attributes of a Bundle ID.
type BundleIDAttributes struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Platform   string `json:"platform"`
	SeedID     string `json:"seedId,omitempty"`
}

// BundleIDCreateRequest represents the request body for creating a Bundle ID.
type BundleIDCreateRequest struct {
	Data BundleIDCreateRequestData `json:"data"`
}

// BundleIDCreateRequestData represents the data for creating a Bundle ID.
type BundleIDCreateRequestData struct {
	Type       string                          `json:"type"`
	Attributes BundleIDCreateRequestAttributes `json:"attributes"`
}

// BundleIDCreateRequestAttributes represents the attributes for creating a
// Bundle ID. The identifier and platform are immutable once created; only the
// name can be changed afterwards.
type BundleIDCreateRequestAttributes struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Platform   string `json:"platform"`
	SeedID     string `json:"seedId,omitempty"`
}

// BundleIDUpdateRequest represents the request body for updating a Bundle ID.
type BundleIDUpdateRequest struct {
	Data BundleIDUpdateRequestData `json:"data"`
}

// BundleIDUpdateRequestData represents the data for updating a Bundle ID.
type BundleIDUpdateRequestData struct {
	Type       string                          `json:"type"`
	ID         string                          `json:"id"`
	Attributes BundleIDUpdateRequestAttributes `json:"attributes"`
}

// BundleIDUpdateRequestAttributes represents the attributes for updating a
// Bundle ID. The App Store Connect API only allows the name to be modified.
type BundleIDUpdateRequestAttributes struct {
	Name string `json:"name"`
}

// BundleIDResponse represents the response from the Bundle ID API.
type BundleIDResponse struct {
	Data  BundleID `json:"data"`
	Links Links    `json:"links,omitempty"`
}

// BundleIDsResponse represents the response for listing Bundle IDs.
type BundleIDsResponse struct {
	Data  []BundleID `json:"data"`
	Links Links      `json:"links,omitempty"`
	Meta  Meta       `json:"meta,omitempty"`
}
