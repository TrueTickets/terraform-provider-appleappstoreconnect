// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/base64"
	"encoding/pem"
	"strings"
	"testing"
)

func TestConvertDERToPEM(t *testing.T) {
	// Sample DER certificate data (this is just test data, not a real certificate)
	derData := []byte{
		0x30, 0x82, 0x01, 0x0a, 0x02, 0x82, 0x01, 0x01,
		0x00, 0xc4, 0xa0, 0x2a, 0x1a, 0xd3, 0x14, 0x5e,
	}

	// Encode to base64
	base64DER := base64.StdEncoding.EncodeToString(derData)

	// Test conversion
	base64PEMResult, err := convertDERToPEM(base64DER)
	if err != nil {
		t.Fatalf("convertDERToPEM failed: %v", err)
	}

	// Decode the base64 PEM result
	pemBytes, err := base64.StdEncoding.DecodeString(base64PEMResult)
	if err != nil {
		t.Fatalf("Failed to decode base64 PEM result: %v", err)
	}
	pemResult := string(pemBytes)

	// Verify PEM format
	if !strings.HasPrefix(pemResult, "-----BEGIN CERTIFICATE-----") {
		t.Error("PEM output should start with BEGIN CERTIFICATE")
	}
	if !strings.HasSuffix(strings.TrimSpace(pemResult), "-----END CERTIFICATE-----") {
		t.Error("PEM output should end with END CERTIFICATE")
	}

	// Verify we can decode the PEM
	block, _ := pem.Decode([]byte(pemResult))
	if block == nil {
		t.Fatal("Failed to decode PEM block")
	}
	if block.Type != "CERTIFICATE" {
		t.Errorf("Expected block type CERTIFICATE, got %s", block.Type)
	}

	// Verify the DER data matches
	if len(block.Bytes) != len(derData) {
		t.Errorf("DER data length mismatch: expected %d, got %d", len(derData), len(block.Bytes))
	}
	for i := range derData {
		if block.Bytes[i] != derData[i] {
			t.Error("DER data mismatch after PEM conversion")
			break
		}
	}
}

func TestConvertDERToPEM_InvalidBase64(t *testing.T) {
	invalidBase64 := "not-valid-base64!"

	_, err := convertDERToPEM(invalidBase64)
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64") {
		t.Errorf("Expected base64 decode error, got: %v", err)
	}
}

func TestConvertDERToPEM_EmptyInput(t *testing.T) {
	base64PEMResult, err := convertDERToPEM("")
	if err != nil {
		t.Fatalf("convertDERToPEM failed for empty input: %v", err)
	}

	// Decode the base64 PEM result
	pemBytes, err := base64.StdEncoding.DecodeString(base64PEMResult)
	if err != nil {
		t.Fatalf("Failed to decode base64 PEM result: %v", err)
	}
	pemResult := string(pemBytes)

	// Empty base64 should produce valid PEM with empty certificate
	if !strings.HasPrefix(pemResult, "-----BEGIN CERTIFICATE-----") {
		t.Error("PEM output should start with BEGIN CERTIFICATE")
	}
	if !strings.HasSuffix(strings.TrimSpace(pemResult), "-----END CERTIFICATE-----") {
		t.Error("PEM output should end with END CERTIFICATE")
	}
}
