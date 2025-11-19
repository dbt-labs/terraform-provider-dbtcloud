# Local Test for Issue #507

This directory contains a test configuration to reproduce issue #507 locally.

## Issue Description

Issue #507 relates to problems with `dbtcloud_global_connection` resource when using OAuth configuration, specifically:
- Read-only attributes (like `adapter_version`) causing issues on subsequent applies
- OAuth configuration not being properly handled during updates
- Drift detection showing false positives

## Prerequisites

1. dbt Cloud account with appropriate permissions
2. Snowflake account with OAuth configured
3. Snowflake OAuth client credentials

## Setup

1. Create a `terraform.tfvars` file with your credentials:

```hcl
dbt_cloud_account_id = 12345
dbt_cloud_token      = "your-dbt-cloud-token"
dbt_cloud_host_url   = "https://cloud.getdbt.com/api"

snowflake_account            = "your-snowflake-account"
snowflake_database           = "YOUR_DATABASE"
snowflake_warehouse          = "YOUR_WAREHOUSE"
snowflake_oauth_client_id    = "your-oauth-client-id"
snowflake_oauth_client_secret = "your-oauth-client-secret"
```

**Note:** Make sure to add `terraform.tfvars` to `.gitignore` to avoid committing sensitive data!

2. Initialize Terraform:

```bash
terraform init
```

## Reproducing the Issue

1. **First Apply** - Create the resources:

```bash
terraform plan
terraform apply
```

This should succeed and create:
- OAuth configuration
- Global connection with OAuth
- Test project
- Test environment using the connection

2. **Second Apply** - Check for drift:

```bash
terraform plan
```

If issue #507 is present, you may see:
- Unexpected changes detected (drift)
- Errors related to read-only attributes
- OAuth configuration issues

3. **Second Apply** - Attempt to apply changes:

```bash
terraform apply
```

This may fail or show unexpected behavior if the issue is present.

## Expected Behavior

**Without the bug fix:**
- Second `terraform plan` may show unexpected drift
- Second `terraform apply` may fail with errors about read-only attributes
- OAuth configuration may be incorrectly handled

**With the bug fix:**
- Second `terraform plan` should show no changes
- Second `terraform apply` should succeed with no changes
- OAuth configuration should be properly maintained

## Cleanup

To clean up all created resources:

```bash
terraform destroy
```

## Debugging

If you encounter issues:

1. Enable Terraform debug logging:
```bash
export TF_LOG=DEBUG
terraform plan 2>&1 | tee terraform-debug.log
```

2. Check the state file for unexpected values:
```bash
terraform show
```

3. Check specific resource details:
```bash
terraform state show dbtcloud_global_connection.test_snowflake_oauth
terraform state show dbtcloud_oauth_configuration.test_oauth
```

## Related Files

- CI/CD test: `.github/workflows/semantic_layer_issue_516.yml`
- Documentation: `.github/workflows/README_ISSUE_TESTS.md`

