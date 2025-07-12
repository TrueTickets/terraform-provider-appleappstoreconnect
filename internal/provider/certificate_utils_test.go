// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
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

// createTestCertificate creates a self-signed certificate for testing.
func createTestCertificate(t *testing.T) *x509.Certificate {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{"CA"},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"123 Test St"},
			PostalCode:    []string{"12345"},
		},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:       []string{"localhost", "test.example.com"},
		EmailAddresses: []string{"test@example.com"},
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}

func TestExtractCertificateCAIssuers(t *testing.T) {
	// Create a test certificate with CA issuers
	cert := createTestCertificateWithAIA(t)

	// Encode the certificate to base64 DER format
	base64DER := base64.StdEncoding.EncodeToString(cert.Raw)

	// Extract CA issuers
	caIssuers, err := extractCertificateCAIssuers(base64DER)
	if err != nil {
		t.Fatalf("extractCertificateCAIssuers failed: %v", err)
	}

	// Check that we got the expected CA issuers
	expectedIssuers := []string{
		"http://ca.example.com/ca.crt",
		"http://backup-ca.example.com/ca.crt",
	}

	if len(caIssuers) != len(expectedIssuers) {
		t.Errorf("Expected %d CA issuers, got %d", len(expectedIssuers), len(caIssuers))
	}

	for _, expected := range expectedIssuers {
		found := false
		for _, issuer := range caIssuers {
			if issuer == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected CA issuer %s not found", expected)
		}
	}

	t.Logf("Found %d CA issuers: %v", len(caIssuers), caIssuers)
}

func TestExtractCertificateCAIssuers_InvalidBase64(t *testing.T) {
	invalidBase64 := "not-valid-base64!"

	_, err := extractCertificateCAIssuers(invalidBase64)
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64") {
		t.Errorf("Expected base64 decode error, got: %v", err)
	}
}

func TestExtractCertificateCAIssuers_InvalidCertificate(t *testing.T) {
	// Valid base64 but not a valid certificate
	invalidCert := base64.StdEncoding.EncodeToString([]byte("not a certificate"))

	_, err := extractCertificateCAIssuers(invalidCert)
	if err == nil {
		t.Error("Expected error for invalid certificate, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse X509 certificate") {
		t.Errorf("Expected certificate parse error, got: %v", err)
	}
}

func TestExtractCertificateCAIssuers_EmptyIssuers(t *testing.T) {
	// Create a certificate without CA issuers
	cert := createTestCertificate(t)

	// Encode the certificate to base64 DER format
	base64DER := base64.StdEncoding.EncodeToString(cert.Raw)

	// Extract CA issuers
	caIssuers, err := extractCertificateCAIssuers(base64DER)
	if err != nil {
		t.Fatalf("extractCertificateCAIssuers failed: %v", err)
	}

	// Should return empty slice for certificate without CA issuers
	if len(caIssuers) != 0 {
		t.Errorf("Expected 0 CA issuers, got %d", len(caIssuers))
	}
}

// createTestCertificateWithAIA creates a test certificate with Authority Information Access extension.
func createTestCertificateWithAIA(t *testing.T) *x509.Certificate {
	// Generate a private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	// Create certificate template with AIA extension
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
			Country:      []string{"US"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	// Add Authority Information Access URLs
	template.IssuingCertificateURL = []string{
		"http://ca.example.com/ca.crt",
		"http://backup-ca.example.com/ca.crt",
	}
	template.OCSPServer = []string{
		"http://ocsp.example.com",
		"http://ocsp-backup.example.com",
	}

	// Create the certificate
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	return cert
}
