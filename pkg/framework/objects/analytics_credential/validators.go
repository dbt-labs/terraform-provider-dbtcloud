package analytics_credential

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// snowflakeAuthValidator checks that the secrets of the snowflake block match its auth type
// and that every secret comes with its version. It reads the whole block, so it works
// without absolute path expressions into the nested attributes.
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
	var block types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root(blockSnowflake), &block)...)
	if resp.Diagnostics.HasError() || block.IsNull() || block.IsUnknown() {
		return
	}

	var snowflake SnowflakeModel
	resp.Diagnostics.Append(block.As(ctx, &snowflake, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}
	if snowflake.AuthType.IsNull() || snowflake.AuthType.IsUnknown() {
		return
	}
	authType := snowflake.AuthType.ValueString()

	attribute := func(name string) path.Path { return path.Root(blockSnowflake).AtName(name) }
	require := func(name string, value types.String) {
		if value.IsNull() {
			resp.Diagnostics.AddAttributeError(
				attribute(name),
				fmt.Sprintf("Missing required attribute `%s`", name),
				fmt.Sprintf("`%s` is required when `auth_type` is `%s`.", name, authType),
			)
		}
	}
	forbid := func(name string, value types.String) {
		if !value.IsNull() {
			resp.Diagnostics.AddAttributeError(
				attribute(name),
				fmt.Sprintf("Attribute `%s` not allowed", name),
				fmt.Sprintf("`%s` cannot be set when `auth_type` is `%s`.", name, authType),
			)
		}
	}

	switch authType {
	case authTypePassword:
		require("password_wo", snowflake.PasswordWo)
		forbid("private_key_wo", snowflake.PrivateKeyWo)
		forbid("private_key_passphrase_wo", snowflake.PrivateKeyPassphraseWo)
	case authTypeKeypair:
		require("private_key_wo", snowflake.PrivateKeyWo)
		forbid("password_wo", snowflake.PasswordWo)
	}

	for _, secret := range []struct {
		name    string
		value   types.String
		version types.Int64
	}{
		{"password_wo", snowflake.PasswordWo, snowflake.PasswordWoVersion},
		{"private_key_wo", snowflake.PrivateKeyWo, snowflake.PrivateKeyWoVersion},
		{"private_key_passphrase_wo", snowflake.PrivateKeyPassphraseWo, snowflake.PrivateKeyPassphraseWoVersion},
	} {
		if !secret.value.IsNull() && secret.version.IsNull() {
			resp.Diagnostics.AddAttributeError(
				attribute(secret.name+"_version"),
				fmt.Sprintf("Missing required attribute `%s_version`", secret.name),
				fmt.Sprintf("`%s_version` is required together with `%s`, it triggers the rotation of the secret.", secret.name, secret.name),
			)
		}
	}
}
