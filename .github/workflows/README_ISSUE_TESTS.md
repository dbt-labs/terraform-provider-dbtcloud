# Issue Reproduction Tests

This directory contains GitHub Actions workflows for reproducing and testing specific issues in the terraform-provider-dbtcloud.

## Issue #507 Reproduction - Global Connection OAuth

**Workflow File:** `issue_507_reproduction.yml`

This workflow tests issue #507: Global connection with OAuth configuration and read-only attributes.

### How to Run

1. Go to the **Actions** tab in GitHub
2. Select **"Issue #507 Reproduction - Global Connection OAuth"**
3. Click **"Run workflow"**
4. Choose your options:
   - **run_second_apply**: Whether to run a second terraform apply to reproduce drift issues (default: true)

### Required GitHub Secrets

The following secrets must be configured in your repository settings:

#### Common Secrets
- `TEST_DBT_CLOUD_ACCOUNT_ID` - dbt Cloud Account ID
- `TEST_DBT_CLOUD_TOKEN` - dbt Cloud API Token
- `TEST_DBT_CLOUD_HOST_URL` - dbt Cloud Host URL (e.g., https://cloud.getdbt.com/api)

#### Snowflake OAuth Secrets
- `TEST_SNOWFLAKE_ACCOUNT` - Snowflake Account Identifier
- `TEST_SNOWFLAKE_DATABASE` - Snowflake Database (optional, defaults to TEST_DATABASE)
- `TEST_SNOWFLAKE_WAREHOUSE` - Snowflake Warehouse (optional, defaults to TEST_WAREHOUSE)
- `TEST_SNOWFLAKE_OAUTH_CLIENT_ID` - Snowflake OAuth Client ID
- `TEST_SNOWFLAKE_OAUTH_CLIENT_SECRET` - Snowflake OAuth Client Secret

### What the Workflow Does

1. **Setup Phase**
   - Checks out the code
   - Sets up Go and Terraform
   - Builds and installs the provider

2. **Configuration Phase**
   - Creates test configuration file: `main.tf`
   - Creates global connection with Snowflake native OAuth (global connection, project, environment)

3. **First Apply**
   - Runs `terraform init`
   - Runs `terraform plan` (first run)
   - Runs `terraform apply` (should succeed)

4. **Second Apply (Optional)**
   - Runs `terraform plan` again to check for drift
   - Runs `terraform apply` again (may reproduce the issue)
   - Reports whether the issue was reproduced

5. **Cleanup**
   - Runs `terraform destroy` to clean up all created resources
   - Uploads Terraform logs as artifacts

### Expected Behavior

- First apply should succeed
- Second apply may fail when trying to update global connection with OAuth configuration
- This tests the issue where read-only attributes (like `adapter_version`) cause problems on subsequent applies

### Artifacts

The workflow uploads the following artifacts:
- `.terraform/` directory
- `terraform.tfstate` and `terraform.tfstate.backup`
- `*.tfplan` files

These artifacts are retained for 7 days and can be downloaded for debugging.

### Notes

- The workflow uses `continue-on-error: true` for the second apply steps to ensure cleanup happens even if the issue is reproduced
- All resources are automatically destroyed at the end of the workflow, even if steps fail
- The workflow runs in the `./testing/cicd` directory

