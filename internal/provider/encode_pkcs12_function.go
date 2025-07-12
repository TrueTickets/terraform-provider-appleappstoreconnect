// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"software.sslmate.com/src/go-pkcs12"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = &EncodePKCS12Function{}

type EncodePKCS12Function struct{}

func NewEncodePKCS12Function() function.Function {
	return &EncodePKCS12Function{}
}

func (f *EncodePKCS12Function) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "pkcs12_encode"
}

func (f *EncodePKCS12Function) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Encode certificate and private key to PKCS12 format",
		Description: "Encodes a certificate and private key pair into PKCS12 (P12) format. The output is base64 encoded for use in Terraform configurations.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "certificate_pem",
				Description: "The certificate in PEM format",
			},
			function.StringParameter{
				Name:        "private_key_pem",
				Description: "The private key in PEM format",
			},
			function.StringParameter{
				Name:        "password",
				Description: "Password to protect the PKCS12 file",
			},
		},
		Return: function.StringReturn{},
	}
}

func (f *EncodePKCS12Function) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var certPEM string
	var keyPEM string
	var password string

	// Read arguments
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &certPEM, &keyPEM, &password))
	if resp.Error != nil {
		return
	}

	// Parse certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		resp.Error = function.NewFuncError("Failed to parse certificate PEM")
		return
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to parse certificate: %s", err))
		return
	}

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		resp.Error = function.NewFuncError("Failed to parse private key PEM")
		return
	}

	var privateKey interface{}
	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	default:
		resp.Error = function.NewFuncError(fmt.Sprintf("Unsupported private key type: %s", keyBlock.Type))
		return
	}

	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to parse private key: %s", err))
		return
	}

	// Create PKCS12
	p12Data, err := pkcs12.Modern.Encode(privateKey, cert, nil, password)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to encode PKCS12: %s", err))
		return
	}

	// Encode to base64
	result := base64.StdEncoding.EncodeToString(p12Data)

	// Set result
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, result))
}
