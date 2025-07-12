// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
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

func TestExtractCertificateExtensions(t *testing.T) {
	// Create a test certificate with known extensions
	cert := createTestCertificate(t)

	// Encode the certificate to base64 DER format
	base64DER := base64.StdEncoding.EncodeToString(cert.Raw)

	// Extract extensions
	extensions, err := extractCertificateExtensions(base64DER)
	if err != nil {
		t.Fatalf("extractCertificateExtensions failed: %v", err)
	}

	// Check that we got some extensions
	if len(extensions) == 0 {
		t.Error("Expected at least some extensions, got none")
	}

	// Check for specific extensions we know should be present
	expectedExtensions := []string{
		"keyUsage",
		"extKeyUsage",
		"subjectAltName",
		"keyUsage_parsed",
		"extKeyUsage_parsed",
		"subjectAltName_parsed",
	}

	for _, expectedExt := range expectedExtensions {
		if _, exists := extensions[expectedExt]; !exists {
			t.Errorf("Expected extension %s not found in extensions map", expectedExt)
		}
	}

	// Verify parsed key usage contains expected values
	if keyUsageParsed, exists := extensions["keyUsage_parsed"]; exists {
		if !strings.Contains(keyUsageParsed, "Digital Signature") {
			t.Error("Expected 'Digital Signature' in parsed key usage")
		}
		if !strings.Contains(keyUsageParsed, "Key Encipherment") {
			t.Error("Expected 'Key Encipherment' in parsed key usage")
		}
	}

	// Verify parsed extended key usage contains expected values
	if extKeyUsageParsed, exists := extensions["extKeyUsage_parsed"]; exists {
		if !strings.Contains(extKeyUsageParsed, "Server Authentication") {
			t.Error("Expected 'Server Authentication' in parsed extended key usage")
		}
		if !strings.Contains(extKeyUsageParsed, "Client Authentication") {
			t.Error("Expected 'Client Authentication' in parsed extended key usage")
		}
	}

	// Verify parsed SAN contains expected values
	if sanParsed, exists := extensions["subjectAltName_parsed"]; exists {
		if !strings.Contains(sanParsed, "DNS:localhost") {
			t.Error("Expected 'DNS:localhost' in parsed SAN")
		}
		if !strings.Contains(sanParsed, "DNS:test.example.com") {
			t.Error("Expected 'DNS:test.example.com' in parsed SAN")
		}
		if !strings.Contains(sanParsed, "email:test@example.com") {
			t.Error("Expected 'email:test@example.com' in parsed SAN")
		}
		if !strings.Contains(sanParsed, "IP:127.0.0.1") {
			t.Error("Expected 'IP:127.0.0.1' in parsed SAN")
		}
	}

	t.Logf("Found %d extensions", len(extensions))
	for k, v := range extensions {
		t.Logf("Extension %s: %s", k, v)
	}
}

func TestExtractCertificateExtensions_InvalidBase64(t *testing.T) {
	invalidBase64 := "not-valid-base64!"

	_, err := extractCertificateExtensions(invalidBase64)
	if err == nil {
		t.Error("Expected error for invalid base64, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode base64") {
		t.Errorf("Expected base64 decode error, got: %v", err)
	}
}

func TestExtractCertificateExtensions_InvalidCertificate(t *testing.T) {
	// Valid base64 but not a valid certificate
	invalidCert := base64.StdEncoding.EncodeToString([]byte("not a certificate"))

	_, err := extractCertificateExtensions(invalidCert)
	if err == nil {
		t.Error("Expected error for invalid certificate, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse X509 certificate") {
		t.Errorf("Expected certificate parse error, got: %v", err)
	}
}

func TestGetExtensionName(t *testing.T) {
	tests := []struct {
		oidParts []int
		oidStr   string
		expected string
	}{
		{[]int{2, 5, 29, 15}, "2.5.29.15", "keyUsage"},
		{[]int{2, 5, 29, 37}, "2.5.29.37", "extKeyUsage"},
		{[]int{2, 5, 29, 17}, "2.5.29.17", "subjectAltName"},
		{[]int{2, 5, 29, 19}, "2.5.29.19", "basicConstraints"},
		{[]int{1, 3, 6, 1, 5, 5, 7, 1, 1}, "1.3.6.1.5.5.7.1.1", "authorityInfoAccess"},
		{[]int{1, 2, 3, 4, 5}, "1.2.3.4.5", ""}, // Unknown OID should return empty string
	}

	for _, test := range tests {
		oid := asn1.ObjectIdentifier(test.oidParts)

		result := getExtensionName(oid)
		if result != test.expected {
			t.Errorf("getExtensionName(%s) = %s, expected %s", test.oidStr, result, test.expected)
		}
	}
}

func TestKeyUsageToString(t *testing.T) {
	tests := []struct {
		usage    x509.KeyUsage
		expected []string
	}{
		{
			x509.KeyUsageDigitalSignature,
			[]string{"Digital Signature"},
		},
		{
			x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
			[]string{"Digital Signature", "Key Encipherment"},
		},
		{
			x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			[]string{"Certificate Sign", "CRL Sign"},
		},
	}

	for _, test := range tests {
		result := keyUsageToString(test.usage)
		for _, expected := range test.expected {
			if !strings.Contains(result, expected) {
				t.Errorf("keyUsageToString(%d) = %s, expected to contain %s", test.usage, result, expected)
			}
		}
	}
}

func TestExtKeyUsageToString(t *testing.T) {
	tests := []struct {
		usage    []x509.ExtKeyUsage
		expected []string
	}{
		{
			[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			[]string{"Server Authentication"},
		},
		{
			[]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			[]string{"Server Authentication", "Client Authentication"},
		},
		{
			[]x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning, x509.ExtKeyUsageEmailProtection},
			[]string{"Code Signing", "Email Protection"},
		},
	}

	for _, test := range tests {
		result := extKeyUsageToString(test.usage)
		for _, expected := range test.expected {
			if !strings.Contains(result, expected) {
				t.Errorf("extKeyUsageToString(%v) = %s, expected to contain %s", test.usage, result, expected)
			}
		}
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

func TestExtractCertificateExtensions_WithAIA(t *testing.T) {
	// Create a test certificate with AIA extension
	cert := createTestCertificateWithAIA(t)

	// Encode the certificate to base64 DER format
	base64DER := base64.StdEncoding.EncodeToString(cert.Raw)

	// Extract extensions
	extensions, err := extractCertificateExtensions(base64DER)
	if err != nil {
		t.Fatalf("extractCertificateExtensions failed: %v", err)
	}

	// Check for Authority Information Access CA Issuers
	if caIssuers, exists := extensions["authorityInfoAccess_caIssuers"]; exists {
		if !strings.Contains(caIssuers, "http://ca.example.com/ca.crt") {
			t.Error("Expected 'http://ca.example.com/ca.crt' in CA Issuers")
		}
		if !strings.Contains(caIssuers, "http://backup-ca.example.com/ca.crt") {
			t.Error("Expected 'http://backup-ca.example.com/ca.crt' in CA Issuers")
		}
		t.Logf("CA Issuers: %s", caIssuers)
	} else {
		t.Error("Expected 'authorityInfoAccess_caIssuers' extension not found")
	}

	// Check for Authority Information Access OCSP
	if ocsp, exists := extensions["authorityInfoAccess_ocsp"]; exists {
		if !strings.Contains(ocsp, "http://ocsp.example.com") {
			t.Error("Expected 'http://ocsp.example.com' in OCSP servers")
		}
		if !strings.Contains(ocsp, "http://ocsp-backup.example.com") {
			t.Error("Expected 'http://ocsp-backup.example.com' in OCSP servers")
		}
		t.Logf("OCSP Servers: %s", ocsp)
	} else {
		t.Error("Expected 'authorityInfoAccess_ocsp' extension not found")
	}

	// Check that the raw authorityInfoAccess extension is also present
	if _, exists := extensions["authorityInfoAccess"]; !exists {
		t.Error("Expected raw 'authorityInfoAccess' extension not found")
	}

	t.Logf("Found %d extensions", len(extensions))
	for k, v := range extensions {
		if strings.Contains(k, "authorityInfoAccess") {
			t.Logf("AIA Extension %s: %s", k, v)
		}
	}
}
