// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/x509"
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

// extractCertificateCAIssuers parses a base64 encoded DER certificate and extracts the CA Issuers URIs.
func extractCertificateCAIssuers(base64DER string) ([]string, error) {
	// Decode the base64 encoded DER
	derBytes, err := base64.StdEncoding.DecodeString(base64DER)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 certificate: %w", err)
	}

	// Parse the X509 certificate
	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse X509 certificate: %w", err)
	}

	// Return CA Issuers URIs from Authority Information Access extension
	return cert.IssuingCertificateURL, nil
}
