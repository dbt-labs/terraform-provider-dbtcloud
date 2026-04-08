# Manual test — SAML auth provider (cases 1 & 2)

## Prerequisites

- dbt Cloud account on an **Enterprise plan** with SSO enabled
- User API token from an **Account Admin** or **Security Admin**
- A SAML 2.0 IdP app (Okta developer account recommended — free at developer.okta.com)
- Terraform >= 1.11 (required for `cert_wo` write-only attribute)

## Setup

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with real values
```

## Running the tests

Only one auth provider may exist per account. Run cases separately.

### Case 1 — SAML with write-only cert

```bash
TF_VAR_run_case=1 terraform init
TF_VAR_run_case=1 terraform apply

# Verify:
# - login_url output is non-empty
# - slug output is auto-generated (not empty)
# - cert_expiry output is populated
# - terraform.tfstate does NOT contain the cert value

# Re-apply should show no changes:
TF_VAR_run_case=1 terraform plan  # expect: No changes

# Rotate cert: bump saml_cert_version in tfvars, then:
TF_VAR_run_case=1 terraform apply  # expect: in-place update, no destroy

# Clean up before running case 2:
TF_VAR_run_case=1 terraform destroy
```

### Case 2 — SAML with custom slug and all optional fields

```bash
TF_VAR_run_case=2 terraform apply

# Verify:
# - login_url contains the custom slug (var.saml_slug)
# - slug output matches var.saml_slug exactly
# - cert value appears as (sensitive) in plan, not plaintext
# - allow_password_backdoor=false is accepted without error
# - Re-apply shows no changes

TF_VAR_run_case=2 terraform plan  # expect: No changes
TF_VAR_run_case=2 terraform destroy
```

## Checking state

```bash
# Confirm cert_wo is absent from state (case 1):
cat terraform.tfstate | grep -i cert  # should show cert_expiry_date only, no cert value

# Confirm cert is marked sensitive in state (case 2):
cat terraform.tfstate | grep -i cert  # value will be present but Terraform marks it sensitive
```
