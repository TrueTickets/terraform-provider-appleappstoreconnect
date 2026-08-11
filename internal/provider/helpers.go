// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// RelationshipOne is a JSON:API to-one relationship payload, used in the
// `relationships` object of create/update request bodies.
type RelationshipOne struct {
	Data RelationshipData `json:"data"`
}

// RelationshipMany is a JSON:API to-many relationship payload, used in the
// `relationships` object of create/update request bodies.
type RelationshipMany struct {
	Data []RelationshipData `json:"data"`
}

// doPaginated performs a GET request and follows the App Store Connect
// pagination links (`links.next`) until they are exhausted, returning every
// element of the `data` arrays across all pages.
//
// The App Store Connect API caps a single page at 200 items, so any listing
// that can exceed that (territories, price points, ...) must be fetched this
// way. Each returned json.RawMessage is a single element of the combined
// `data` arrays, leaving decoding to the caller.
//
// `links.next` is an absolute URL; only its host-relative part is reused so
// the request is reissued against the client's configured base URL (which also
// keeps the helper testable against an httptest server).
func doPaginated(ctx context.Context, client *Client, req Request) ([]json.RawMessage, error) {
	var all []json.RawMessage

	for {
		resp, err := client.Do(ctx, req)
		if err != nil {
			return nil, err
		}

		if len(resp.Data) > 0 {
			var page []json.RawMessage
			if err := json.Unmarshal(resp.Data, &page); err != nil {
				return nil, fmt.Errorf("failed to parse paginated data: %w", err)
			}
			all = append(all, page...)
		}

		if resp.Links.Next == "" {
			break
		}

		next, err := url.Parse(resp.Links.Next)
		if err != nil {
			return nil, fmt.Errorf("failed to parse next pagination link %q: %w", resp.Links.Next, err)
		}

		// Reissue against the configured base URL using the host-relative
		// path+query of the next link. Query is left empty as it is already
		// encoded in the endpoint.
		req = Request{
			Method:   http.MethodGet,
			Endpoint: next.RequestURI(),
		}
	}

	return all, nil
}
