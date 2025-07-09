// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"time"
)

// PassTypeID represents a Pass Type ID in the App Store Connect API.
type PassTypeID struct {
	Type       string               `json:"type"`
	ID         string               `json:"id"`
	Attributes PassTypeIDAttributes `json:"attributes"`
	Links      ResourceLinks        `json:"links,omitempty"`
}

// PassTypeIDAttributes represents the attributes of a Pass Type ID.
type PassTypeIDAttributes struct {
	Identifier  string     `json:"identifier"`
	Description string     `json:"description"`
	CreatedDate *time.Time `json:"createdDate,omitempty"`
}

// ResourceLinks represents the links for a resource.
type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}

// PassTypeIDCreateRequest represents the request body for creating a Pass Type ID.
type PassTypeIDCreateRequest struct {
	Data PassTypeIDCreateRequestData `json:"data"`
}

// PassTypeIDCreateRequestData represents the data for creating a Pass Type ID.
type PassTypeIDCreateRequestData struct {
	Type       string                            `json:"type"`
	Attributes PassTypeIDCreateRequestAttributes `json:"attributes"`
}

// PassTypeIDCreateRequestAttributes represents the attributes for creating a Pass Type ID.
type PassTypeIDCreateRequestAttributes struct {
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
}

// PassTypeIDResponse represents the response from the Pass Type ID API.
type PassTypeIDResponse struct {
	Data  PassTypeID `json:"data"`
	Links Links      `json:"links,omitempty"`
}

// PassTypeIDsResponse represents the response for listing Pass Type IDs.
type PassTypeIDsResponse struct {
	Data  []PassTypeID `json:"data"`
	Links Links        `json:"links,omitempty"`
	Meta  Meta         `json:"meta,omitempty"`
}
