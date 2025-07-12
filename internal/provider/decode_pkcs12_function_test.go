// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"math/big"
	"regexp"
	"software.sslmate.com/src/go-pkcs12"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestDecodePKCS12Function(t *testing.T) {
	// Generate test certificate and key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"Test Org"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Create PKCS12
	p12Data, err := pkcs12.Modern.Encode(priv, cert, nil, "test123")
	if err != nil {
		t.Fatalf("Failed to encode PKCS12: %v", err)
	}

	p12Base64 := base64.StdEncoding.EncodeToString(p12Data)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDecodePKCS12FunctionConfig(p12Base64),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("output.cert_pem", "value", regexp.MustCompile(`-----BEGIN CERTIFICATE-----`)),
					resource.TestMatchResourceAttr("output.key_pem", "value", regexp.MustCompile(`-----BEGIN PRIVATE KEY-----`)),
				),
			},
		},
	})
}

func TestDecodePKCS12Function_InvalidBase64(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDecodePKCS12FunctionConfig("invalid base64!@#"),
				ExpectError: regexp.MustCompile("Failed to decode base64"),
			},
		},
	})
}

func TestDecodePKCS12Function_WrongPassword(t *testing.T) {
	// Generate test certificate and key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %v", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		t.Fatalf("Failed to parse certificate: %v", err)
	}

	// Create PKCS12 with one password
	p12Data, err := pkcs12.Modern.Encode(priv, cert, nil, "correct_password")
	if err != nil {
		t.Fatalf("Failed to encode PKCS12: %v", err)
	}

	p12Base64 := base64.StdEncoding.EncodeToString(p12Data)

	// Try to decode with wrong password
	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDecodePKCS12FunctionConfigWithPassword(p12Base64, "wrong_password"),
				ExpectError: regexp.MustCompile("Failed to decode PKCS12"),
			},
		},
	})
}

func testAccDecodePKCS12FunctionConfig(p12Base64 string) string {
	return testAccDecodePKCS12FunctionConfigWithPassword(p12Base64, "test123")
}

func testAccDecodePKCS12FunctionConfigWithPassword(p12Base64, password string) string {
	return `
provider "appleappstoreconnect" {
  issuer_id   = "test"
  key_id      = "test"
  private_key = "test"
}

locals {
  decoded = provider::appleappstoreconnect::pkcs12_decode("` + p12Base64 + `", "` + password + `")
}

output "cert_pem" {
  value = local.decoded.certificate_pem
}

output "key_pem" {
  value = local.decoded.private_key_pem
  sensitive = true
}
`
}
