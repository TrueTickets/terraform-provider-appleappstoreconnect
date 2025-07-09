// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"time"
)

// Certificate represents a Certificate in the App Store Connect API.
type Certificate struct {
	Type          string                    `json:"type"`
	ID            string                    `json:"id"`
	Attributes    CertificateAttributes     `json:"attributes"`
	Relationships *CertificateRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks             `json:"links,omitempty"`
}

// CertificateAttributes represents the attributes of a Certificate.
type CertificateAttributes struct {
	SerialNumber       string     `json:"serialNumber,omitempty"`
	CertificateContent string     `json:"certificateContent,omitempty"`
	DisplayName        string     `json:"displayName,omitempty"`
	Name               string     `json:"name,omitempty"`
	CertificateType    string     `json:"certificateType"`
	Platform           string     `json:"platform,omitempty"`
	ExpirationDate     *time.Time `json:"expirationDate,omitempty"`
	RequesterEmail     string     `json:"requesterEmail,omitempty"`
	RequesterFirstName string     `json:"requesterFirstName,omitempty"`
	RequesterLastName  string     `json:"requesterLastName,omitempty"`
}

// CertificateRelationships represents the relationships of a Certificate.
type CertificateRelationships struct {
	PassTypeId *Relationship `json:"passTypeId,omitempty"`
}

// Relationship represents a generic relationship.
type Relationship struct {
	Links *RelationshipLinks `json:"links,omitempty"`
	Data  *RelationshipData  `json:"data,omitempty"`
}

// RelationshipLinks represents the links in a relationship.
type RelationshipLinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}

// RelationshipData represents the data in a relationship.
type RelationshipData struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// CertificateCreateRequest represents the request body for creating a Certificate.
type CertificateCreateRequest struct {
	Data CertificateCreateRequestData `json:"data"`
}

// CertificateCreateRequestData represents the data for creating a Certificate.
type CertificateCreateRequestData struct {
	Type          string                                 `json:"type"`
	Attributes    CertificateCreateRequestAttributes     `json:"attributes"`
	Relationships *CertificateCreateRequestRelationships `json:"relationships,omitempty"`
}

// CertificateCreateRequestAttributes represents the attributes for creating a Certificate.
type CertificateCreateRequestAttributes struct {
	CertificateType string `json:"certificateType"`
	CsrContent      string `json:"csrContent"`
}

// CertificateCreateRequestRelationships represents the relationships for creating a Certificate.
type CertificateCreateRequestRelationships struct {
	PassTypeId *CertificateCreateRequestRelationship `json:"passTypeId,omitempty"`
}

// CertificateCreateRequestRelationship represents a relationship in the create request.
type CertificateCreateRequestRelationship struct {
	Data RelationshipData `json:"data"`
}

// CertificateResponse represents the response from the Certificate API.
type CertificateResponse struct {
	Data  Certificate `json:"data"`
	Links Links       `json:"links,omitempty"`
}

// CertificatesResponse represents the response for listing Certificates.
type CertificatesResponse struct {
	Data  []Certificate `json:"data"`
	Links Links         `json:"links,omitempty"`
	Meta  Meta          `json:"meta,omitempty"`
}

// Certificate types.
const (
	CertificateTypeIOSDevelopment           = "IOS_DEVELOPMENT"
	CertificateTypeIOSDistribution          = "IOS_DISTRIBUTION"
	CertificateTypeMacAppDevelopment        = "MAC_APP_DEVELOPMENT"
	CertificateTypeMacAppDistribution       = "MAC_APP_DISTRIBUTION"
	CertificateTypeMacInstallerDistribution = "MAC_INSTALLER_DISTRIBUTION"
	CertificateTypePassTypeID               = "PASS_TYPE_ID"
	CertificateTypePassTypeIDWithNFC        = "PASS_TYPE_ID_WITH_NFC"
	CertificateTypeDeveloperIDKext          = "DEVELOPER_ID_KEXT"
	CertificateTypeDeveloperIDApplication   = "DEVELOPER_ID_APPLICATION"
	CertificateTypeDevelopmentPushSSL       = "DEVELOPMENT_PUSH_SSL"
	CertificateTypeProductionPushSSL        = "PRODUCTION_PUSH_SSL"
	CertificateTypePushSSL                  = "PUSH_SSL"
)
