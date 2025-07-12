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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ function.Function = &DecodePKCS12Function{}

type DecodePKCS12Function struct{}

func NewDecodePKCS12Function() function.Function {
	return &DecodePKCS12Function{}
}

func (f *DecodePKCS12Function) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "pkcs12_decode"
}

func (f *DecodePKCS12Function) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Decode PKCS12 format to certificate and private key",
		Description: "Decodes a PKCS12 (P12) file to extract the certificate and private key in PEM format.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "pkcs12_base64",
				Description: "The PKCS12 data in base64 encoded format",
			},
			function.StringParameter{
				Name:        "password",
				Description: "Password to decrypt the PKCS12 file",
			},
		},
		Return: function.ObjectReturn{
			AttributeTypes: map[string]attr.Type{
				"certificate_pem": types.StringType,
				"private_key_pem": types.StringType,
			},
		},
	}
}

func (f *DecodePKCS12Function) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var pkcs12Base64 string
	var password string

	// Read arguments
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &pkcs12Base64, &password))
	if resp.Error != nil {
		return
	}

	// Decode base64
	p12Data, err := base64.StdEncoding.DecodeString(pkcs12Base64)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to decode base64: %s", err))
		return
	}

	// Decode PKCS12
	privateKey, cert, err := pkcs12.Decode(p12Data, password)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to decode PKCS12: %s", err))
		return
	}

	// Encode certificate to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	// Encode private key to PEM
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		resp.Error = function.NewFuncError(fmt.Sprintf("Failed to marshal private key: %s", err))
		return
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyDER,
	})

	// Create result object
	result := map[string]attr.Value{
		"certificate_pem": types.StringValue(string(certPEM)),
		"private_key_pem": types.StringValue(string(keyPEM)),
	}

	resultValue, diags := types.ObjectValue(map[string]attr.Type{
		"certificate_pem": types.StringType,
		"private_key_pem": types.StringType,
	}, result)

	if diags.HasError() {
		resp.Error = function.NewFuncError("Failed to create result object")
		return
	}

	// Set result
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, resultValue))
}
