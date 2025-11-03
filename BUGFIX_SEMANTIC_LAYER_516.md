# Bug Fix: Semantic Layer Resources Inconsistent Behavior (Issue #516)

## Issue Summary

When using semantic layer resources (`dbtcloud_semantic_layer_configuration`, `dbtcloud_semantic_layer_credential_service_token_mapping`, etc.), users experienced inconsistent behavior:

1. **First run**: Resource created successfully
2. **Post-creation**: Terraform reported the resource "could not be found" and removed it from state
3. **Second run**: Terraform attempted to recreate the resource, resulting in duplicate key constraint errors

## Root Cause

The issue was in `/pkg/dbt_cloud/client.go` in the `doRequestWithRetry()` function. The error handling code incorrectly treated **all HTTP 400 (Bad Request) errors as "resource-not-found"** errors:

```go
if res.StatusCode == 400 {
    return nil, fmt.Errorf("resource-not-found: %s", body)
}
```

### Why This Was Wrong

HTTP status codes have specific meanings:
- **400 Bad Request**: Validation errors, constraint violations, malformed requests, duplicate keys, etc.
- **404 Not Found**: The requested resource does not exist

By treating 400 as "resource-not-found", the code caused semantic layer resources to be incorrectly removed from Terraform state when they encountered validation or constraint errors.

### The Bug Sequence

1. User creates semantic layer configuration (succeeds)
2. Terraform immediately calls Read() to refresh state
3. If any transient issue occurs (timing, eventual consistency, etc.) and the API returns a 400 error
4. The 400 error gets incorrectly labeled as "resource-not-found"
5. The resource Read method sees "resource-not-found" prefix and removes the resource from state
6. On next run, Terraform tries to create it again
7. API returns 400 with duplicate key constraint error: `duplicate key value violates unique constraint "sl_creds_project_name_unique"`
8. This 400 also gets labeled as "resource-not-found", compounding the confusion

## The Fix

Changed the error handling in `/pkg/dbt_cloud/client.go` line 226-228:

**Before:**
```go
if res.StatusCode == 400 {
    return nil, fmt.Errorf("resource-not-found: %s", body)
}
```

**After:**
```go
// Handle 400 Bad Request errors - these are validation/constraint errors, NOT "not found" errors
if res.StatusCode == 400 {
    return nil, fmt.Errorf("bad-request: %s", body)
}
```

## Impact Analysis

### What This Fixes
- Semantic layer resources will no longer be removed from state when encountering 400 errors
- Duplicate key constraint errors will be reported correctly as validation errors, not "resource not found"
- Resources that are successfully created will remain in Terraform state

### What This Doesn't Break
- All existing code checks for `strings.HasPrefix(err.Error(), "resource-not-found")` or `strings.Contains(err.Error(), "resource-not-found")`
- No code currently checks for "bad-request" prefix
- 404 errors are still correctly handled as "resource-not-found"
- Resource Read methods will still correctly handle actual "resource not found" scenarios (404s)

### Testing
- Reviewed all usages of "resource-not-found" error handling in the codebase
- Verified no existing code depends on 400 being treated as "resource-not-found"
- Confirmed the fix aligns with HTTP status code standards
- No linter errors introduced

## Resources Affected

This fix affects all resources that use the dbt Cloud API client, but most notably:
- `dbtcloud_semantic_layer_configuration`
- `dbtcloud_bigquery_semantic_layer_credential`
- `dbtcloud_databricks_semantic_layer_credential`
- `dbtcloud_redshift_semantic_layer_credential`
- `dbtcloud_snowflake_semantic_layer_credential`
- `dbtcloud_postgres_semantic_layer_credential`
- `dbtcloud_semantic_layer_credential_service_token_mapping`

And potentially any other resource that could encounter 400 errors during normal operation.

## Related Issue

GitHub Issue: https://github.com/dbt-labs/terraform-provider-dbtcloud/issues/516

## Recommendations

Users who experienced this issue should:
1. Update to the provider version containing this fix
2. If resources were removed from state incorrectly, they can:
   - Use `terraform import` to re-import existing resources
   - Or allow Terraform to create them (should now work correctly on subsequent runs)

## Additional Notes

The issue reporter noted this only occurred in CI/CD environments, not local development. This is likely because:
- CI/CD environments may have different network characteristics (latency, timing)
- The race condition was timing-dependent
- Local runs might complete fast enough to avoid the timing window where the bug manifested


