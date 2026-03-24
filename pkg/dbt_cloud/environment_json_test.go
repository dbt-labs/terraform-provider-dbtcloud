package dbt_cloud

import (
	"encoding/json"
	"strings"
	"testing"
)

// Regression for https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/623 :
// the API rejects "deployment_type":""; clearing the field must use a nil pointer so
// omitempty drops the key (do not use a non-nil pointer to "").
func TestEnvironmentMarshal_deploymentTypeNilDoesNotEmitEmptyString(t *testing.T) {
	env := Environment{
		Account_Id:     1,
		Project_Id:     1,
		Name:           "n",
		Dbt_Version:    "latest",
		Type:           "deployment",
		DeploymentType: nil,
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"deployment_type":""`) {
		t.Fatalf("nil DeploymentType must not marshal as empty string: %s", b)
	}
}

func TestEnvironmentMarshal_deploymentTypePointerToEmptyString(t *testing.T) {
	empty := ""
	env := Environment{
		Account_Id:     1,
		Project_Id:     1,
		Name:           "n",
		Dbt_Version:    "latest",
		Type:           "deployment",
		DeploymentType: &empty,
	}
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"deployment_type":""`) {
		t.Fatalf("pointer to empty string encodes as empty JSON string (invalid for API): %s", b)
	}
}
