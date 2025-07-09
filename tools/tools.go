// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build generate

package tools

import (
	_ "github.com/hashicorp/copywrite"
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

// Generate copyright headers
//go:generate go run github.com/hashicorp/copywrite headers -d .. --config ../.copywrite.hcl

// Format OpenTofu code for use in documentation.
// If you do not have OpenTofu installed, you can remove the formatting command, but it is suggested
// to ensure the documentation is formatted properly.
//go:generate tofu fmt -recursive ../examples/

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir .. -provider-name appleappstoreconnect --tf-version 1.6.0
