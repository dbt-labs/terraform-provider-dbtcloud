# Local Testing Scenarios

This directory contains test scenarios for validating provider fixes.

## Testing the Permission Error Fix

### Scenario: Job Environment Change with Limited Permissions

This tests the fix for issue #537 where the provider would panic when trying to move a job to an environment where the token lacks write permissions.

### Prerequisites

1. Two dbt Cloud environments in the same project:
   - Environment A (e.g., "Production") - token has write access
   - Environment B (e.g., "Staging") - token does NOT have write access

2. Service token with:
   - ACCOUNT VIEWER permissions
   - TEAM ADMIN permissions
   - Environment-level write scope ONLY for Environment A

### Expected Behavior

**Before the fix:**
- Terraform would crash with: `panic: runtime error: invalid memory address or nil pointer dereference`
- The job would be destroyed but not recreated, leaving you with no job

**After the fix:**
- Terraform fails gracefully with a clear error message:
  - "forbidden: The API token does not have permission to perform this action. This may be due to environment-level permissions..."
- The job remains in the original environment (or if destroyed, you get a clear error before state corruption)

### Test Steps

1. Create a job in Environment A (where you have permissions)
2. Apply successfully
3. Change the job's `environment_id` to Environment B (where you DON'T have permissions)
4. Run `terraform apply`
5. Verify you get a clear permission error instead of a panic

