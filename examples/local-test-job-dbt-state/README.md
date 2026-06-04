# Local test: dbt State / cost_optimization_features

Manually exercises `cost_optimization_features` on `dbtcloud_job` against the
locally built provider.

## 1. Build and point Terraform at the local provider

From the repo root:

```bash
go build -o terraform-provider-dbtcloud .
```

Create a `~/.terraformrc` (or `$TF_CLI_CONFIG_FILE`) with a dev override so
Terraform uses your local binary instead of the registry:

```hcl
provider_installation {
  dev_overrides {
    "dbt-labs/dbtcloud" = "/absolute/path/to/terraform-provider-dbtcloud"
  }
  direct {}
}
```

Point it at the directory containing the built binary. With `dev_overrides` in
effect you can skip `terraform init` and run `terraform plan`/`apply` directly —
Terraform prints a warning that the override is active, which is expected. (Do
not run `terraform init`: it would try to install the provider from the registry
and fail, since this build isn't published there.)

## 2. Set credentials

```bash
export DBT_CLOUD_ACCOUNT_ID=...
export DBT_CLOUD_TOKEN=...
export DBT_CLOUD_HOST_URL=https://cloud.getdbt.com/api   # adjust if needed
```

## 3. Account requirements

- dbt State: account entitled to dbt State (non-Team plan) **and** the
  `ORC-3638-enable-dbt-state` LaunchDarkly flag on.
- SAO: `orc-2664-sao-beta` flag on, plus a Fusion `dbt_version` and a
  production/staging environment (or the `orc-2714-allow-sao-for-all-environments`
  flag). The config defaults `dbt_version = latest-fusion` and uses a
  `production` environment.

## 4. Run

```bash
terraform plan
terraform apply
```

Check the outputs: both jobs should report `force_node_selection = false`,
derived by the API from the set you sent. Then try Case C (mixed set, should
fail at plan time) and Case D (empty set, `force_node_selection = true`) from
`main.tf`.

```bash
terraform destroy
```
