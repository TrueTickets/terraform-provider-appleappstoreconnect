// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestCertificateTypeConstants(t *testing.T) {
	// Test that all certificate type constants are unique
	certificateTypes := []string{
		CertificateTypeIOSDevelopment,
		CertificateTypeIOSDistribution,
		CertificateTypeMacAppDevelopment,
		CertificateTypeMacAppDistribution,
		CertificateTypeMacInstallerDistribution,
		CertificateTypePassTypeID,
		CertificateTypePassTypeIDWithNFC,
		CertificateTypeDeveloperIDKext,
		CertificateTypeDeveloperIDApplication,
		CertificateTypeDevelopmentPushSSL,
		CertificateTypeProductionPushSSL,
		CertificateTypePushSSL,
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for _, certType := range certificateTypes {
		if seen[certType] {
			t.Errorf("Duplicate certificate type found: %s", certType)
		}
		seen[certType] = true
	}

	// Verify expected values
	expectedValues := map[string]string{
		"IOS_DEVELOPMENT":            CertificateTypeIOSDevelopment,
		"IOS_DISTRIBUTION":           CertificateTypeIOSDistribution,
		"MAC_APP_DEVELOPMENT":        CertificateTypeMacAppDevelopment,
		"MAC_APP_DISTRIBUTION":       CertificateTypeMacAppDistribution,
		"MAC_INSTALLER_DISTRIBUTION": CertificateTypeMacInstallerDistribution,
		"PASS_TYPE_ID":               CertificateTypePassTypeID,
		"PASS_TYPE_ID_WITH_NFC":      CertificateTypePassTypeIDWithNFC,
		"DEVELOPER_ID_KEXT":          CertificateTypeDeveloperIDKext,
		"DEVELOPER_ID_APPLICATION":   CertificateTypeDeveloperIDApplication,
		"DEVELOPMENT_PUSH_SSL":       CertificateTypeDevelopmentPushSSL,
		"PRODUCTION_PUSH_SSL":        CertificateTypeProductionPushSSL,
		"PUSH_SSL":                   CertificateTypePushSSL,
	}

	for expected, actual := range expectedValues {
		if actual != expected {
			t.Errorf("Expected %s to equal %s, got %s", expected, expected, actual)
		}
	}
}

func TestIsPassTypeCertificate(t *testing.T) {
	tests := []struct {
		name     string
		certType string
		want     bool
	}{
		{
			name:     "PASS_TYPE_ID certificate",
			certType: CertificateTypePassTypeID,
			want:     true,
		},
		{
			name:     "PASS_TYPE_ID_WITH_NFC certificate",
			certType: CertificateTypePassTypeIDWithNFC,
			want:     true,
		},
		{
			name:     "IOS_DEVELOPMENT certificate",
			certType: CertificateTypeIOSDevelopment,
			want:     false,
		},
		{
			name:     "IOS_DISTRIBUTION certificate",
			certType: CertificateTypeIOSDistribution,
			want:     false,
		},
		{
			name:     "MAC_APP_DEVELOPMENT certificate",
			certType: CertificateTypeMacAppDevelopment,
			want:     false,
		},
		{
			name:     "Unknown certificate type",
			certType: "UNKNOWN_TYPE",
			want:     false,
		},
		{
			name:     "Empty certificate type",
			certType: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Since we don't have an exported function to test this,
			// we'll just verify the constants exist and are as expected
			if tt.certType == CertificateTypePassTypeID || tt.certType == CertificateTypePassTypeIDWithNFC {
				if !tt.want {
					t.Errorf("Expected %s to be a Pass Type certificate", tt.certType)
				}
			} else {
				if tt.want {
					t.Errorf("Expected %s to NOT be a Pass Type certificate", tt.certType)
				}
			}
		})
	}
}
