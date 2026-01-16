# PR: Fix CI/Merge Job Deferral & Add cost_optimization_features Support

## Summary

This PR addresses critical bugs affecting CI and Merge jobs, and introduces the new `cost_optimization_features` attribute for better State-Aware Orchestration (SAO) control.

## Bug Fixes

### 1. CI/Merge Jobs Can Now Use `deferring_environment_id`

**Problem:** The provider was incorrectly dropping `deferring_environment_id` for CI and Merge jobs, preventing users from configuring artifact deferral on these job types.

**Root Cause:** The `CreateJob` function in `pkg/dbt_cloud/job.go` contained logic that explicitly zeroed out deferral settings for jobs with `github_webhook`, `git_provider_webhook`, or `on_merge` triggers:

```go
// REMOVED: This was incorrectly dropping deferral for CI/Merge jobs
if isGithubWebhook || isOnMerge {
    deferringJobId = 0
    deferringEnvironmentID = 0
}
```

**Fix:** Removed this incorrect logic. CI/Merge jobs CAN have `deferring_environment_id` for artifact deferral - this is separate from SAO.

### 2. CI/Merge Jobs No Longer Fail with SAO Errors

**Problem:** CI and Merge jobs would fail with API error `405: State aware orchestration is not available for CI or Merge jobs` when creating jobs, even without explicitly setting `force_node_selection`.

**Root Cause:** The API rejects `force_node_selection=false` (or any explicit value) for CI/Merge jobs. The provider needed to completely omit this field for these job types.

**Fix:** The provider now correctly handles `force_node_selection` for different job types:
- CI/Merge jobs: Field is omitted entirely
- Non-Fusion jobs: Field is set to `true` (SAO requires Fusion)
- Fusion jobs: Uses the configured value or omits if null

## New Feature: `cost_optimization_features`

Added the `cost_optimization_features` attribute as the preferred way to control SAO features.

### Usage

```hcl
resource "dbtcloud_job" "production_job" {
  name           = "Production Daily Run"
  project_id     = dbtcloud_project.example.id
  environment_id = dbtcloud_environment.production.environment_id
  dbt_version    = "latest-fusion"  # Required for SAO
  
  execute_steps = ["dbt build"]
  
  # Enable SAO using the new attribute
  cost_optimization_features = ["state_aware_orchestration"]
  
  triggers = {
    schedule = true
  }
}
```

### Valid Values

| Value | Description |
|-------|-------------|
| `state_aware_orchestration` | Enables SAO for optimized job execution |

### Relationship with `force_node_selection`

| `cost_optimization_features` | Equivalent `force_node_selection` | SAO Status |
|------------------------------|-----------------------------------|------------|
| `["state_aware_orchestration"]` | `false` | Enabled |
| `[]` or not set | `true` | Disabled |

### Deprecation Notice

`force_node_selection` is now deprecated and will be removed in a future major version. Users should migrate to `cost_optimization_features`.

## Files Changed

### Core API Client
- `pkg/dbt_cloud/job.go`
  - Added `CostOptimizationFeatures` field to `Job` struct
  - Added `costOptimizationFeatures` parameter to `CreateJob` function
  - **Removed** incorrect deferral-dropping logic for CI/Merge jobs

### Resource Implementation
- `pkg/framework/objects/job/resource.go`
  - Added handling for `cost_optimization_features` in Create, Read, and Update
  - Properly populates the field from API responses

### Schema
- `pkg/framework/objects/job/schema.go`
  - Added `cost_optimization_features` as a Set attribute with validation
  - Added deprecation message to `force_node_selection`

### Model
- `pkg/framework/objects/job/model.go`
  - Added `CostOptimizationFeatures types.Set` to all job models

### Tests
- `pkg/framework/objects/job/resource_acceptance_cost_optimization_test.go` (NEW)
  - `TestAccDbtCloudJobResourceCIWithDeferral` - Validates CI jobs with deferral
  - `TestAccDbtCloudJobResourceMergeWithDeferral` - Validates Merge jobs with deferral
  - `TestAccDbtCloudJobResourceCINoForceNodeSelection` - Validates CI jobs work without force_node_selection
  - `TestAccDbtCloudJobResourceMergeNoForceNodeSelection` - Validates Merge jobs work without force_node_selection
  - `TestAccDbtCloudJobResourceCostOptimizationFeatures` - Validates new feature

### Documentation
- `docs/resources/job.md`
  - Added SAO section with comprehensive guidance
  - Added migration guide from `force_node_selection` to `cost_optimization_features`
  - Clarified deferral vs SAO distinction

## Test Results

All new acceptance tests pass:

```
--- PASS: TestAccDbtCloudJobResourceMergeWithDeferral (4.30s)
--- PASS: TestAccDbtCloudJobResourceCostOptimizationFeatures (4.53s)
--- PASS: TestAccDbtCloudJobResourceMergeNoForceNodeSelection (4.65s)
--- PASS: TestAccDbtCloudJobResourceCINoForceNodeSelection (5.26s)
--- PASS: TestAccDbtCloudJobResourceCIWithDeferral (5.66s)
PASS
```

## Migration Guide

### From `force_node_selection` to `cost_optimization_features`

**Before (deprecated):**
```hcl
resource "dbtcloud_job" "example" {
  # ... other config ...
  force_node_selection = false  # Enable SAO
}
```

**After (recommended):**
```hcl
resource "dbtcloud_job" "example" {
  # ... other config ...
  cost_optimization_features = ["state_aware_orchestration"]  # Enable SAO
}
```

## Breaking Changes

None. All changes are backward compatible:
- `force_node_selection` continues to work (with deprecation warning)
- Existing CI/Merge jobs will now correctly preserve `deferring_environment_id`

## Related Issues

- Fixes deferral being dropped for CI/Merge jobs
- Fixes SAO validation errors for CI/Merge jobs
- Implements `cost_optimization_features` for better SAO control

## Checklist

- [x] Code changes
- [x] Tests added/updated
- [x] Documentation updated
- [x] Changelog updated
- [x] Backward compatible
