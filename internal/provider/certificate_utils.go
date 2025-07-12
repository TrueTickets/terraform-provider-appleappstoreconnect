// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
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

// extractCertificateExtensions parses a base64 encoded DER certificate and extracts its X509v3 extensions.
func extractCertificateExtensions(base64DER string) (map[string]string, error) {
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

	extensions := make(map[string]string)

	// Extract standard extensions
	for _, ext := range cert.Extensions {
		oidStr := ext.Id.String()

		// Convert the extension value to hex string for consistent representation
		valueHex := hex.EncodeToString(ext.Value)

		// Try to get a human-readable name for common OIDs
		name := getExtensionName(ext.Id)
		if name != "" {
			extensions[name] = valueHex
		} else {
			extensions[oidStr] = valueHex
		}
	}

	// Add some parsed extension values for common extensions
	if cert.KeyUsage != 0 {
		extensions["keyUsage_parsed"] = keyUsageToString(cert.KeyUsage)
	}

	if len(cert.ExtKeyUsage) > 0 {
		extensions["extKeyUsage_parsed"] = extKeyUsageToString(cert.ExtKeyUsage)
	}

	if len(cert.DNSNames) > 0 || len(cert.IPAddresses) > 0 || len(cert.EmailAddresses) > 0 || len(cert.URIs) > 0 {
		sanValues := []string{}
		for _, dns := range cert.DNSNames {
			sanValues = append(sanValues, "DNS:"+dns)
		}
		for _, ip := range cert.IPAddresses {
			sanValues = append(sanValues, "IP:"+ip.String())
		}
		for _, email := range cert.EmailAddresses {
			sanValues = append(sanValues, "email:"+email)
		}
		for _, uri := range cert.URIs {
			sanValues = append(sanValues, "URI:"+uri.String())
		}
		extensions["subjectAltName_parsed"] = strings.Join(sanValues, ",")
	}

	// Parse Authority Information Access extension to extract CA Issuers URIs
	if len(cert.IssuingCertificateURL) > 0 {
		extensions["authorityInfoAccess_caIssuers"] = strings.Join(cert.IssuingCertificateURL, ",")
	}

	// Parse Authority Information Access extension to extract OCSP Server URIs
	if len(cert.OCSPServer) > 0 {
		extensions["authorityInfoAccess_ocsp"] = strings.Join(cert.OCSPServer, ",")
	}

	return extensions, nil
}

// getExtensionName returns a human-readable name for common X509 extension OIDs.
func getExtensionName(oid asn1.ObjectIdentifier) string {
	oidMap := map[string]string{
		"2.5.29.15":               "keyUsage",
		"2.5.29.37":               "extKeyUsage",
		"2.5.29.17":               "subjectAltName",
		"2.5.29.18":               "issuerAltName",
		"2.5.29.19":               "basicConstraints",
		"2.5.29.14":               "subjectKeyIdentifier",
		"2.5.29.35":               "authorityKeyIdentifier",
		"2.5.29.31":               "cRLDistributionPoints",
		"2.5.29.32":               "certificatePolicies",
		"1.3.6.1.5.5.7.1.1":       "authorityInfoAccess",
		"1.3.6.1.5.5.7.1.11":      "subjectInfoAccess",
		"2.5.29.54":               "inhibitAnyPolicy",
		"2.5.29.46":               "freshestCRL",
		"2.5.29.36":               "policyConstraints",
		"2.5.29.30":               "nameConstraints",
		"2.5.29.33":               "policyMappings",
		"1.3.6.1.4.1.11129.2.4.2": "certificateTransparency",
	}

	return oidMap[oid.String()]
}

// keyUsageToString converts x509.KeyUsage flags to a human-readable string.
func keyUsageToString(usage x509.KeyUsage) string {
	var usages []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		usages = append(usages, "Digital Signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		usages = append(usages, "Content Commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		usages = append(usages, "Key Encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		usages = append(usages, "Data Encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		usages = append(usages, "Key Agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		usages = append(usages, "Certificate Sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		usages = append(usages, "CRL Sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		usages = append(usages, "Encipher Only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		usages = append(usages, "Decipher Only")
	}

	return strings.Join(usages, ", ")
}

// extKeyUsageToString converts []x509.ExtKeyUsage to a human-readable string.
func extKeyUsageToString(usage []x509.ExtKeyUsage) string {
	var usages []string

	for _, u := range usage {
		switch u {
		case x509.ExtKeyUsageAny:
			usages = append(usages, "Any")
		case x509.ExtKeyUsageServerAuth:
			usages = append(usages, "Server Authentication")
		case x509.ExtKeyUsageClientAuth:
			usages = append(usages, "Client Authentication")
		case x509.ExtKeyUsageCodeSigning:
			usages = append(usages, "Code Signing")
		case x509.ExtKeyUsageEmailProtection:
			usages = append(usages, "Email Protection")
		case x509.ExtKeyUsageIPSECEndSystem:
			usages = append(usages, "IPSEC End System")
		case x509.ExtKeyUsageIPSECTunnel:
			usages = append(usages, "IPSEC Tunnel")
		case x509.ExtKeyUsageIPSECUser:
			usages = append(usages, "IPSEC User")
		case x509.ExtKeyUsageTimeStamping:
			usages = append(usages, "Time Stamping")
		case x509.ExtKeyUsageOCSPSigning:
			usages = append(usages, "OCSP Signing")
		case x509.ExtKeyUsageMicrosoftServerGatedCrypto:
			usages = append(usages, "Microsoft Server Gated Crypto")
		case x509.ExtKeyUsageNetscapeServerGatedCrypto:
			usages = append(usages, "Netscape Server Gated Crypto")
		case x509.ExtKeyUsageMicrosoftCommercialCodeSigning:
			usages = append(usages, "Microsoft Commercial Code Signing")
		case x509.ExtKeyUsageMicrosoftKernelCodeSigning:
			usages = append(usages, "Microsoft Kernel Code Signing")
		}
	}

	return strings.Join(usages, ", ")
}
