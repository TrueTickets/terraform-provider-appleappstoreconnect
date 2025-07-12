// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

// convertDERToPEM converts a base64 encoded DER certificate to base64 encoded PEM format.
func convertDERToPEM(base64DER string) (string, error) {
	// Decode the base64 encoded DER
	derBytes, err := base64.StdEncoding.DecodeString(base64DER)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 certificate: %w", err)
	}

	// Create PEM block
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}

	// Encode to PEM
	pemBytes := pem.EncodeToMemory(pemBlock)
	if pemBytes == nil {
		return "", fmt.Errorf("failed to encode certificate to PEM")
	}

	// Encode PEM to base64
	base64PEM := base64.StdEncoding.EncodeToString(pemBytes)

	return base64PEM, nil
}
