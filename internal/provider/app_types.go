// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

// App represents an App in the App Store Connect API.
type App struct {
	Type       string        `json:"type"`
	ID         string        `json:"id"`
	Attributes AppAttributes `json:"attributes"`
	Links      ResourceLinks `json:"links,omitempty"`
}

// AppAttributes represents the attributes of an App.
type AppAttributes struct {
	Name          string `json:"name,omitempty"`
	BundleID      string `json:"bundleId,omitempty"`
	SKU           string `json:"sku,omitempty"`
	PrimaryLocale string `json:"primaryLocale,omitempty"`
}

// AppResponse represents the response for reading a single App.
type AppResponse struct {
	Data  App   `json:"data"`
	Links Links `json:"links,omitempty"`
}

// AppsResponse represents the response for listing Apps.
type AppsResponse struct {
	Data  []App `json:"data"`
	Links Links `json:"links,omitempty"`
	Meta  Meta  `json:"meta,omitempty"`
}
