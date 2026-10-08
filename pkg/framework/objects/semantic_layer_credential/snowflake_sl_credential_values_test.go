package semantic_layer_credential

import (
	"testing"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// A write-only attribute is null in the plan and only present in the
// configuration, so the payload has to be built from both.
func TestCredentialValuesUsesWriteOnlySecrets(t *testing.T) {
	plan := SnowflakeSLCredentialModel{
		Credential: snowflake_credential.SnowflakeCredentialResourceModel{
			AuthType:             types.StringValue("keypair"),
			Role:                 types.StringValue("my_role"),
			Warehouse:            types.StringValue("my_warehouse"),
			User:                 types.StringValue("my_user"),
			Password:             types.StringNull(),
			PrivateKey:           types.StringNull(),
			PrivateKeyPassphrase: types.StringNull(),
		},
	}
	config := SnowflakeSLCredentialModel{
		Credential: snowflake_credential.SnowflakeCredentialResourceModel{
			PrivateKeyWo:           types.StringValue("the-private-key"),
			PrivateKeyPassphraseWo: types.StringValue("the-passphrase"),
			PasswordWo:             types.StringNull(),
		},
	}

	values := credentialValues(plan, config)

	for field, want := range map[string]string{
		"private_key":            "the-private-key",
		"private_key_passphrase": "the-passphrase",
		"role":                   "my_role",
		"warehouse":              "my_warehouse",
		"user":                   "my_user",
		"auth_type":              "keypair",
	} {
		if got := values[field]; got != want {
			t.Errorf("values[%q] = %q, want %q", field, got, want)
		}
	}
}

// A configuration that uses the plain attributes keeps working.
func TestCredentialValuesUsesPlainSecrets(t *testing.T) {
	plan := SnowflakeSLCredentialModel{
		Credential: snowflake_credential.SnowflakeCredentialResourceModel{
			Password:   types.StringValue("the-password"),
			PrivateKey: types.StringValue("the-plain-key"),
		},
	}
	config := SnowflakeSLCredentialModel{
		Credential: snowflake_credential.SnowflakeCredentialResourceModel{
			PasswordWo:   types.StringNull(),
			PrivateKeyWo: types.StringNull(),
		},
	}

	values := credentialValues(plan, config)

	if got := values["password"]; got != "the-password" {
		t.Errorf("values[\"password\"] = %q, want the-password", got)
	}
	if got := values["private_key"]; got != "the-plain-key" {
		t.Errorf("values[\"private_key\"] = %q, want the-plain-key", got)
	}
}
