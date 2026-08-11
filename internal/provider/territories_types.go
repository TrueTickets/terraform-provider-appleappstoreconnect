// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// Territory represents an App Store territory in the App Store Connect API.
// The resource ID is the territory code (e.g. "USA", "FRA", "GBR").
type Territory struct {
	Type       string              `json:"type"`
	ID         string              `json:"id"`
	Attributes TerritoryAttributes `json:"attributes"`
}

// TerritoryAttributes represents the attributes of a Territory.
type TerritoryAttributes struct {
	Currency string `json:"currency,omitempty"`
}
