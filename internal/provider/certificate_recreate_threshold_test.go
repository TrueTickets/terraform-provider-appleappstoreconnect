// Copyright (c) TrueTickets, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestCertificateRecreateThresholdPlanModifier exercises the plan modifier
// that drives certificate auto-renewal directly, independent of an Apple API
// call. It pins down the exact scenario from a production incident where a
// certificate expiring in 5 days, with a 21-day (1,814,400s) recreate
// threshold configured, was not being flagged for replacement.
func TestCertificateRecreateThresholdPlanModifier(t *testing.T) {
	t.Parallel()

	var schemaResp resource.SchemaResponse
	(&CertificateResource{}).Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("schema error: %v", schemaResp.Diagnostics)
	}
	sch := schemaResp.Schema

	const dateLayout = "2006-01-02T15:04:05Z"

	tests := []struct {
		name              string
		expiresIn         time.Duration
		recreateThreshold int64
		wantReplace       bool
	}{
		{
			// Mirrors the reported production incident: a cert 5 days from
			// expiry with a 21-day (1,814,400s) threshold configured.
			name:              "expiring within threshold requires replacement",
			expiresIn:         5 * 24 * time.Hour,
			recreateThreshold: 3600 * 24 * 21,
			wantReplace:       true,
		},
		{
			name:              "expiring outside threshold does not require replacement",
			expiresIn:         120 * 24 * time.Hour,
			recreateThreshold: 3600 * 24 * 21,
			wantReplace:       false,
		},
		{
			name:              "zero threshold disables auto-renewal even when imminent",
			expiresIn:         5 * 24 * time.Hour,
			recreateThreshold: 0,
			wantReplace:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			expirationDate := time.Now().Add(tt.expiresIn).UTC().Format(dateLayout)

			model := CertificateResourceModel{
				ID:                    types.StringValue("CERT123"),
				CertificateType:       types.StringValue("PASS_TYPE_ID"),
				CsrContent:            types.StringValue("-----BEGIN CERTIFICATE REQUEST-----\nMII\n-----END CERTIFICATE REQUEST-----\n"),
				PrivateKeyPEM:         types.StringNull(),
				CertificateContent:    types.StringValue("AAAA"),
				CertificateContentPEM: types.StringNull(),
				CertificateCAIssuers:  types.ListNull(types.StringType),
				DisplayName:           types.StringValue("Pass Type ID Cert"),
				Name:                  types.StringValue("Pass Type ID Cert"),
				Platform:              types.StringValue("IOS"),
				SerialNumber:          types.StringValue("1234567890"),
				ExpirationDate:        types.StringValue(expirationDate),
				RecreateThreshold:     types.Int64Value(tt.recreateThreshold),
				Relationships:         types.ObjectNull(map[string]attr.Type{"pass_type_id": types.StringType}),
				PKCS12BundlePassword:  types.StringNull(),
				PKCS12BundleContent:   types.StringNull(),
			}

			state := tfsdk.State{Schema: sch}
			if diags := state.Set(ctx, &model); diags.HasError() {
				t.Fatalf("state.Set error: %v", diags)
			}

			// Plan mirrors state exactly: nothing in config changed, this is
			// a pure refresh-only "terraform plan" pass.
			plan := tfsdk.Plan{Schema: sch}
			if diags := plan.Set(ctx, &model); diags.HasError() {
				t.Fatalf("plan.Set error: %v", diags)
			}

			req := planmodifier.StringRequest{
				Path:        path.Root("expiration_date"),
				State:       state,
				Plan:        plan,
				StateValue:  model.ExpirationDate,
				PlanValue:   model.ExpirationDate,
				ConfigValue: types.StringUnknown(), // expiration_date is Computed-only, not settable in config
			}
			resp := planmodifier.StringResponse{PlanValue: req.PlanValue}

			CertificateRecreateThresholdPlanModifier{}.PlanModifyString(ctx, req, &resp)

			if resp.Diagnostics.HasError() {
				t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
			}

			if resp.RequiresReplace != tt.wantReplace {
				t.Errorf("RequiresReplace = %v, want %v", resp.RequiresReplace, tt.wantReplace)
			}

			// RequiresReplace alone is not sufficient: Terraform only honors
			// it when the attribute's planned value actually differs from
			// state. A tripped threshold must also mark the plan value
			// unknown, or the whole resource stays byte-identical to state
			// and Terraform silently drops the replacement.
			if tt.wantReplace && !resp.PlanValue.IsUnknown() {
				t.Errorf("PlanValue = %v, want unknown", resp.PlanValue)
			}
			if !tt.wantReplace && !resp.PlanValue.Equal(model.ExpirationDate) {
				t.Errorf("PlanValue = %v, want unchanged (%v)", resp.PlanValue, model.ExpirationDate)
			}
		})
	}
}
