# Fix Service Token Permission Inconsistency Bug

## Summary

Fixes two bugs in the Terraform provider:

1. **Service Token Permission Inconsistency**: "Provider produced inconsistent result after apply" errors when creating service tokens with empty `writable_environment_categories`.

2. **Service Token Permission Assignment 404s**: Service token permissions cannot be changed after creation, but the provider was attempting to update them after creation, causing 404 errors. Permissions must be set during token creation.

## Problems

### Problem 1: Permission Inconsistency

When creating service tokens with `writable_environment_categories = []` (empty set), Terraform reports an inconsistency error:

```
Error: Provider produced inconsistent result after apply
.service_token_permissions: planned set element does not correlate with any element in actual.
```

### Root Cause

The bug is in `pkg/framework/objects/service_token/model.go` in the `ConvertServiceTokenPermissionDataToModel` function:

1. **Plan**: User specifies `writable_environment_categories = []` (empty set)
2. **API Request**: Empty array `[]` is sent to the API (correct)
3. **API Response**: API returns empty array `[]` or omits the field
4. **Model Conversion** (lines 93-95): Code automatically converts empty to `["all"]`:
   ```go
   if len(permission.WritableEnvs) == 0 {
       permission.WritableEnvs = []dbt_cloud.EnvironmentCategory{dbt_cloud.EnvironmentCategory_All}
   }
   ```
5. **Result**: Plan has `[]`, actual state has `["all"]` → **Mismatch Error**

### Problem 2: Permission Assignment 404s

When creating service tokens with permissions, the provider:
1. Creates the token without permissions
2. Attempts to update permissions via `UpdateServiceTokenPermissions` endpoint
3. Receives 404 errors because **service token permissions cannot be changed after creation**

The dbt Cloud API requires permissions to be set **during token creation**, not afterward.

## Solutions

### Solution 1: Remove Automatic Normalization

Removed the automatic normalization that converts empty `WritableEnvs` to `["all"]`. This preserves the plan value when the API returns empty, ensuring consistency between plan and actual state.

### Solution 2: Set Permissions During Creation

Modified the provider to include permissions in the service token creation request instead of attempting to update them afterward.

**Changes:**

**File**: `pkg/dbt_cloud/service_token.go`
- Modified `CreateServiceToken()` to accept `permissions []ServiceTokenPermission` parameter
- Permissions are now included in the creation request payload

**File**: `pkg/framework/objects/service_token/model.go`
- Added `ConvertServiceTokenPermissionModelToDataForCreation()` function
- This function converts permissions for creation (without `ServiceTokenID` - API assigns it)
- Original `ConvertServiceTokenPermissionModelToData()` remains for update operations

**File**: `pkg/framework/objects/service_token/resource.go`
- Modified `Create()` function to:
  1. Convert permissions for creation (without ServiceTokenID)
  2. Pass permissions to `CreateServiceToken()` during creation
  3. Read back permissions from the created token (or fetch if not in response)
  4. Removed `UpdateServiceTokenPermissions()` call (permissions are immutable after creation)

### Changes Summary

**File**: `pkg/framework/objects/service_token/model.go`

**Removed lines 93-95** (automatic normalization):
```go
// REMOVED: Automatic normalization that caused inconsistency
if len(permission.WritableEnvs) == 0 {
    permission.WritableEnvs = []dbt_cloud.EnvironmentCategory{dbt_cloud.EnvironmentCategory_All}
}
```

**Added** `ConvertServiceTokenPermissionModelToDataForCreation()` function for creation-time permission conversion.

## Impact

### Before Fix
- Users setting `writable_environment_categories = []` get inconsistency errors
- Terraform state becomes inconsistent
- Subsequent applies may fail or attempt to recreate permissions

### After Fix
- Empty `writable_environment_categories` are preserved as empty (matching plan)
- Schema default `["all"]` still applies when users don't specify the attribute
- Users who explicitly set `[]` get `[]` back (matching their plan)
- No more inconsistency errors
- Service token permissions are set during creation (not updated afterward)
- No more 404 errors when creating service tokens with permissions

## Backward Compatibility

✅ **Fully backward compatible**

- Users who don't set `writable_environment_categories` will still get the schema default `["all"]` (via schema default, not normalization)
- Users who explicitly set `["all"]` will get `["all"]` back
- Users who set `[]` will now get `[]` back (this was broken before, now fixed)

## Testing

### Manual Testing

1. **Build the provider**:
   ```bash
   go build -o terraform-provider-dbtcloud
   ```

2. **Configure Terraform dev_overrides** in `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "dbt-labs/dbtcloud" = "/path/to/terraform-provider-dbtcloud"
     }
     direct {}
   }
   ```

3. **Test with empty writable_environment_categories**:
   ```terraform
   resource "dbtcloud_service_token" "test" {
     name = "test-token"
     
     service_token_permissions {
       permission_set = "account_admin"
       all_projects   = true
       writable_environment_categories = []  # Empty set
     }
   }
   ```

4. **Verify**:
   - `terraform plan` should succeed
   - `terraform apply` should succeed without inconsistency errors
   - `terraform show` should show `writable_environment_categories = []`

### Automated Testing

The fix has been tested with the terraform-dbtcloud-yaml E2E test suite:
- Service tokens with empty `writable_environment_categories` create successfully
- No "inconsistent result" errors observed
- State matches plan correctly
- Service tokens with permissions create successfully (permissions set during creation)
- No 404 errors when assigning permissions (permissions included in creation request)

## Related Issues

This fixes the issues described in:
- terraform-dbtcloud-yaml: `dev_support/KNOWN_ISSUES.md` - Service Token Permission Inconsistency
- terraform-dbtcloud-yaml: `dev_support/APPLY_ERRORS_ANALYSIS.md` - Error Category 1 (Inconsistency) and Category 4 (Permission Assignment 404s)

## Checklist

- [x] Code changes implemented
- [x] Backward compatibility verified
- [x] Manual testing completed
- [x] E2E testing completed
- [x] Documentation updated (KNOWN_ISSUES.md)
- [ ] Unit tests added (if applicable)
- [ ] Integration tests updated (if applicable)

## Notes

- The schema default (`["all"]`) is still applied via Terraform's schema default mechanism, not via normalization
- This change only affects the conversion from API response to Terraform model
- The "hack" in `ConvertServiceTokenPermissionModelToData` (lines 63-66) remains unchanged as it correctly handles the API request side

