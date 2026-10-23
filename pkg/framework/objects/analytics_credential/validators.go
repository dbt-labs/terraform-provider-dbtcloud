package analytics_credential

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// blockChecker reports the missing, forbidden and unversioned attributes of an adapter
// block, according to its auth type. It reads the whole block, so it works without
// absolute path expressions into the nested attributes.
type blockChecker struct {
	diags    *diag.Diagnostics
	block    string
	authType string
}

func (c blockChecker) attribute(name string) path.Path { return path.Root(c.block).AtName(name) }

func (c blockChecker) require(name string, value types.String) {
	if value.IsNull() {
		c.diags.AddAttributeError(
			c.attribute(name),
			fmt.Sprintf("Missing required attribute `%s`", name),
			fmt.Sprintf("`%s` is required when `auth_type` is `%s`.", name, c.authType),
		)
	}
}

func (c blockChecker) forbid(name string, value types.String) {
	if !value.IsNull() {
		c.diags.AddAttributeError(
			c.attribute(name),
			fmt.Sprintf("Attribute `%s` not allowed", name),
			fmt.Sprintf("`%s` cannot be set when `auth_type` is `%s`.", name, c.authType),
		)
	}
}

// version requires the version of a write-only secret that is set
func (c blockChecker) version(name string, value types.String, version types.Int64) {
	if !value.IsNull() && version.IsNull() {
		c.diags.AddAttributeError(
			c.attribute(name+"_version"),
			fmt.Sprintf("Missing required attribute `%s_version`", name),
			fmt.Sprintf("`%s_version` is required together with `%s`, it triggers the rotation of the secret.", name, name),
		)
	}
}

// blockOf reads the adapter block from the configuration; ok is false when it is null or unknown
func blockOf(ctx context.Context, config resource.ValidateConfigRequest, name string, target any, diags *diag.Diagnostics) bool {
	var block types.Object
	diags.Append(config.Config.GetAttribute(ctx, path.Root(name), &block)...)
	if diags.HasError() || block.IsNull() || block.IsUnknown() {
		return false
	}
	diags.Append(block.As(ctx, target, basetypes.ObjectAsOptions{})...)
	return !diags.HasError()
}

// snowflakeAuthValidator checks that the secrets of the snowflake block match its auth type
// and that every secret comes with its version.
type snowflakeAuthValidator struct{}

func (snowflakeAuthValidator) Description(_ context.Context) string {
	return "password auth requires password_wo and forbids the key fields; keypair auth requires private_key_wo and forbids password_wo; every secret needs its version"
}

func (v snowflakeAuthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (snowflakeAuthValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var snowflake SnowflakeModel
	if !blockOf(ctx, req, blockSnowflake, &snowflake, &resp.Diagnostics) {
		return
	}
	if snowflake.AuthType.IsNull() || snowflake.AuthType.IsUnknown() {
		return
	}
	check := blockChecker{diags: &resp.Diagnostics, block: blockSnowflake, authType: snowflake.AuthType.ValueString()}

	switch check.authType {
	case authTypePassword:
		check.require("password_wo", snowflake.PasswordWo)
		check.forbid("private_key_wo", snowflake.PrivateKeyWo)
		check.forbid("private_key_passphrase_wo", snowflake.PrivateKeyPassphraseWo)
	case authTypeKeypair:
		check.require("private_key_wo", snowflake.PrivateKeyWo)
		check.forbid("password_wo", snowflake.PasswordWo)
	}

	check.version("password_wo", snowflake.PasswordWo, snowflake.PasswordWoVersion)
	check.version("private_key_wo", snowflake.PrivateKeyWo, snowflake.PrivateKeyWoVersion)
	check.version("private_key_passphrase_wo", snowflake.PrivateKeyPassphraseWo, snowflake.PrivateKeyPassphraseWoVersion)
}

// bigqueryAuthValidator checks that the fields of the bigquery block match its auth type:
// a service account (the default) needs the key fields, workload identity federation needs
// the provider path and forbids the key fields.
type bigqueryAuthValidator struct{}

func (bigqueryAuthValidator) Description(_ context.Context) string {
	return "service account auth requires the key fields and forbids the workload identity federation fields; external-oauth-wif requires workload_pool_provider_path and forbids the key fields"
}

func (v bigqueryAuthValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (bigqueryAuthValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var bigquery BigQueryModel
	if !blockOf(ctx, req, blockBigQuery, &bigquery, &resp.Diagnostics) {
		return
	}
	if bigquery.AuthType.IsUnknown() {
		return
	}
	authType := authTypeServiceAccount
	if !bigquery.AuthType.IsNull() {
		authType = bigquery.AuthType.ValueString()
	}
	check := blockChecker{diags: &resp.Diagnostics, block: blockBigQuery, authType: authType}

	switch authType {
	case authTypeServiceAccount:
		check.require("private_key_id", bigquery.PrivateKeyID)
		check.require("private_key_wo", bigquery.PrivateKeyWo)
		check.require("client_email", bigquery.ClientEmail)
		check.require("client_id", bigquery.ClientID)
		check.require("auth_uri", bigquery.AuthURI)
		check.require("token_uri", bigquery.TokenURI)
		check.require("auth_provider_x509_cert_url", bigquery.AuthProviderX509CertURL)
		check.require("client_x509_cert_url", bigquery.ClientX509CertURL)
		check.forbid("workload_pool_provider_path", bigquery.WorkloadPoolProviderPath)
		check.forbid("service_account_impersonation_url", bigquery.ServiceAccountImpersonationURL)
	case authTypeWIF:
		check.require("workload_pool_provider_path", bigquery.WorkloadPoolProviderPath)
		check.forbid("private_key_id", bigquery.PrivateKeyID)
		check.forbid("private_key_wo", bigquery.PrivateKeyWo)
		check.forbid("client_email", bigquery.ClientEmail)
		check.forbid("client_id", bigquery.ClientID)
		check.forbid("auth_uri", bigquery.AuthURI)
		check.forbid("token_uri", bigquery.TokenURI)
		check.forbid("auth_provider_x509_cert_url", bigquery.AuthProviderX509CertURL)
		check.forbid("client_x509_cert_url", bigquery.ClientX509CertURL)
		if !bigquery.PrivateKeyWoVersion.IsNull() {
			resp.Diagnostics.AddAttributeError(
				check.attribute("private_key_wo_version"),
				"Attribute `private_key_wo_version` not allowed",
				fmt.Sprintf("`private_key_wo_version` cannot be set when `auth_type` is `%s`.", authType),
			)
		}
	}

	check.version("private_key_wo", bigquery.PrivateKeyWo, bigquery.PrivateKeyWoVersion)
}
