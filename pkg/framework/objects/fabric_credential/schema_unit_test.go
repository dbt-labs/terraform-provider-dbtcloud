package fabric_credential

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// configWith builds a Fabric credential configuration where every attribute is null
// except the string attributes named in set.
func configWith(t *testing.T, set map[string]string) tfsdk.Config {
	t.Helper()
	objectType := resourceSchema.Type().TerraformType(context.Background()).(tftypes.Object)
	values := map[string]tftypes.Value{}
	for name, attributeType := range objectType.AttributeTypes {
		values[name] = tftypes.NewValue(attributeType, nil)
	}
	for name, value := range set {
		values[name] = tftypes.NewValue(tftypes.String, value)
	}
	return tfsdk.Config{
		Schema: resourceSchema,
		Raw:    tftypes.NewValue(objectType, values),
	}
}

func validateAuthentication(t *testing.T, authentication string, set map[string]string) (hasError bool, paths []path.Path) {
	t.Helper()
	attribute := resourceSchema.Attributes["authentication"]
	stringAttribute, ok := attribute.(interface {
		StringValidators() []validator.String
	})
	if !ok {
		t.Fatalf("authentication is not a string attribute with validators")
	}
	request := validator.StringRequest{
		Path:           path.Root("authentication"),
		PathExpression: path.MatchRoot("authentication"),
		Config:         configWith(t, set),
		ConfigValue:    types.StringValue(authentication),
	}
	for _, v := range stringAttribute.StringValidators() {
		response := &validator.StringResponse{}
		v.ValidateString(context.Background(), request, response)
		if response.Diagnostics.HasError() {
			hasError = true
			for _, d := range response.Diagnostics.Errors() {
				if withPath, ok := d.(interface{ Path() path.Path }); ok {
					paths = append(paths, withPath.Path())
				}
			}
		}
	}
	return hasError, paths
}

func TestAuthenticationValidator_RejectsEachConflictingFieldWithExternalOAuth(t *testing.T) {
	// Rule: Each password or service principal field is rejected when the authentication is external_oauth
	//
	// given: the authentication external_oauth and one password or service principal field set
	// when: the configuration is validated
	// then: it fails and the error points at that field
	for _, field := range []string{
		"user", "password", "password_wo", "tenant_id", "client_id", "client_secret", "client_secret_wo",
	} {
		t.Run(field, func(t *testing.T) {
			hasError, paths := validateAuthentication(t, "external_oauth", map[string]string{field: "value"})
			if !hasError {
				t.Fatalf("expected an error when %s is set", field)
			}
			want := path.Root(field)
			for _, got := range paths {
				if got.Equal(want) {
					return
				}
			}
			t.Errorf("no error points at %s, got %v", want, paths)
		})
	}
}

func TestAuthenticationValidator_AcceptsExternalOAuthWithoutThoseFields(t *testing.T) {
	// Rule: An external_oauth credential with only the schema set is valid
	//
	// given: the authentication external_oauth and only the schema set
	// when: the configuration is validated
	// then: it passes
	if hasError, _ := validateAuthentication(t, "external_oauth", map[string]string{"schema": "dbt"}); hasError {
		t.Errorf("expected no error")
	}
}

func TestAuthenticationValidator_LeavesOtherAuthenticationsAlone(t *testing.T) {
	// Rule: The older authentication values keep accepting their own fields
	//
	// given: the authentication ServicePrincipal with a tenant id set
	// when: the configuration is validated
	// then: it passes
	if hasError, _ := validateAuthentication(t, "ServicePrincipal", map[string]string{"tenant_id": "tenant"}); hasError {
		t.Errorf("expected no error")
	}
}

func TestAuthenticationValidator_RejectsUnknownValues(t *testing.T) {
	// Rule: An authentication value outside the three options is rejected
	//
	// given: the authentication other
	// when: the configuration is validated
	// then: it fails
	if hasError, _ := validateAuthentication(t, "other", nil); !hasError {
		t.Errorf("expected an error")
	}
}
