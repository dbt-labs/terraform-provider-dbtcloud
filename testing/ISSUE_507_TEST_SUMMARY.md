# Issue #507 Test Configuration - Summary

## Overview

This document summarizes the test configuration created for reproducing issue #507 (global connection with OAuth configuration) in the terraform-provider-dbtcloud.

## Changes Made

### 1. Created GitHub Actions Workflow

**File:** `.github/workflows/issue_507_reproduction.yml`

**Description:**
- Focused workflow for testing issue #507 only
- Creates test configuration with:
  - Global connection with Snowflake native OAuth
  - Test project and environment
- Tests for drift and read-only attribute issues on subsequent applies

**Required GitHub Secrets:**
- `TEST_DBT_CLOUD_ACCOUNT_ID`
- `TEST_DBT_CLOUD_TOKEN`
- `TEST_DBT_CLOUD_HOST_URL`
- `TEST_SNOWFLAKE_ACCOUNT`
- `TEST_SNOWFLAKE_DATABASE` (optional)
- `TEST_SNOWFLAKE_WAREHOUSE` (optional)
- `TEST_SNOWFLAKE_OAUTH_CLIENT_ID`
- `TEST_SNOWFLAKE_OAUTH_CLIENT_SECRET`

### 2. Created Documentation

**File:** `.github/workflows/README_ISSUE_TESTS.md`

Comprehensive documentation including:
- How to run the workflow
- Required GitHub secrets
- What the workflow does
- Expected behavior
- Troubleshooting tips

### 3. Created Local Test Configuration

**Files:**
- `testing/local_test_issue_507/main.tf` - Terraform configuration
- `testing/local_test_issue_507/README.md` - Usage documentation
- `testing/local_test_issue_507/.gitignore` - Git ignore rules

Allows developers to reproduce issue #507 locally before running CI/CD tests.

## How Issue #507 is Reproduced

The test creates a `dbtcloud_global_connection` resource with:
1. Snowflake connection details with native OAuth credentials (`allow_sso`, `oauth_client_id`, `oauth_client_secret`)
2. Read-only attributes (like `adapter_version`) that are exposed in outputs

The test then:
1. Applies the configuration successfully
2. Runs a second plan/apply to check for drift
3. Should detect if read-only attributes cause issues

## Key Features

1. **Focused Testing:** Dedicated workflow for issue #507
2. **Comprehensive Secrets:** All required credentials passed as environment variables
3. **Local Testing:** Developers can test locally before CI/CD
4. **Automatic Cleanup:** Resources are always destroyed, even on failure
5. **Drift Detection:** Second apply specifically tests for read-only attribute issues

## Testing Issue #507

### In CI/CD:
```bash
# From GitHub Actions UI
1. Go to Actions tab
2. Select "Issue #507 Reproduction - Global Connection OAuth"
3. Click "Run workflow"
4. (Optional) Toggle "run_second_apply" setting
5. Click "Run workflow"
```

### Locally:
```bash
cd testing/local_test_issue_507
# Create terraform.tfvars with your credentials
terraform init
terraform plan
terraform apply
# Run again to check for drift
terraform plan
terraform apply
# Cleanup
terraform destroy
```

## Expected Outcomes

### If Issue #507 Exists:
- ❌ Second terraform plan shows unexpected changes
- ❌ Second terraform apply fails or modifies resources
- ❌ Read-only attributes cause drift

### If Issue #507 is Fixed:
- ✅ Second terraform plan shows no changes
- ✅ Second terraform apply succeeds with no changes
- ✅ Read-only attributes are properly ignored

## Related Issues

- **Issue #507:** Global connection OAuth configuration issues

## Next Steps

1. **Configure GitHub Secrets:** Add the required Snowflake OAuth secrets to your repository
2. **Run the Workflow:** Test issue #507 in the CI/CD pipeline
3. **Verify Results:** Check if the issue is reproduced
4. **Fix the Issue:** If reproduced, work on a fix
5. **Re-test:** Run the workflow again to verify the fix

## Files Created

```
.github/workflows/
├── issue_507_reproduction.yml (new)
└── README_ISSUE_TESTS.md (new)

testing/
├── local_test_issue_507/
│   ├── main.tf (new)
│   ├── README.md (new)
│   └── .gitignore (new)
└── ISSUE_507_TEST_SUMMARY.md (new)
```

