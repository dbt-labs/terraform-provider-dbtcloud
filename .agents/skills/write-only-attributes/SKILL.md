---
name: write-only-attributes
description: Use when adding, changing or fixing write-only (`_wo`) attributes, `*_wo_version` fields or any secret attribute (password, token, private key) on a resource, data source or semantic layer credential in this provider. Covers schema, Create/Update, validators and testing, and the bugs seen before.
---

# Write-only (`_wo`) attributes

Secrets can be set with a write-only attribute (`token_wo`, `password_wo`, `private_key_wo`, ...) so they never reach the Terraform state. Write-only attributes need Terraform >= 1.11. Most existing bugs in this area come from copying a neighbouring resource without knowing the points below. The reference implementation is `pkg/framework/objects/snowflake_credential`.

## Schema

- Declare the pair `<name>_wo` (`Optional`, `WriteOnly: true`) and `<name>_wo_version` (`Optional`, `Int64`). The version is how Terraform notices a rotation, because the value itself is never stored.
- On the plain attribute, add `stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("<name>_wo"))` and `helper.PreferWriteOnlyAttributeValidator{WriteOnlyAttributeName: "<name>_wo"}`.
- Always use relative paths (`MatchRelative().AtParent()`), never `path.MatchRoot`. The credential schemas are reused nested under `credential` by the semantic layer resources, and an absolute path then fails with `Invalid Path Expression for Schema` (#712, #710). The same goes for `stringvalidator.PreferWriteOnlyAttribute`, use the helper in `pkg/helper` instead.
- A computed `id` needs `UseStateForUnknown()`, otherwise every plan shows an in-place update (#719).
- If the data source shares its model with the resource, the `_wo` attributes have to exist in the data source schema too (as `Computed`), otherwise reading the config or setting the state fails on the missing field.
- Any validator that requires the plain attribute (for example "`private_key` must be set") has to accept the write-only one as well.

## Create and Update

- Write-only values are **null in the plan and the state**. Read them from `req.Config`, not `req.Plan`, and resolve them with `helper.ResolveWriteOnlyString(config.<Name>Wo, plan.<Name>)`. Only reading the plan sends an empty secret to the API, and the apply still succeeds (#732).
- Decide whether to send the secret with `<name>_wo_version` (and the plain attribute) changing between state and plan, never with the `_wo` value.
- Set the state from the plan, not from the API response. The API never echoes secrets, so writing the response back into the secret fields gives `inconsistent values for sensitive attribute` or an empty value (#732). Keep `<name>_wo` null in the state and carry `<name>_wo_version` over from the plan.
- Do not send a partial update that blanks the secret. Some endpoints replace the whole record, so an empty secret clears what is stored and can drop related links such as service token mappings.
- After import the version is null, so the first plan with `<name>_wo_version` set shows an update. That is expected, add `<name>_wo_version` to `ImportStateVerifyIgnore`.

## Testing

- Acceptance tests that use `_wo` attributes must be guarded with `TerraformVersionChecks: []tfversion.TerraformVersionCheck{tfversion.SkipBelow(tfversion.Version1_11_0)}`. If the Terraform on your `PATH` is older, install a newer one (for example `tenv tf install 1.12.2`) and run with `TF_ACC_TERRAFORM_PATH=<path to the binary>`. Otherwise the tests are skipped and prove nothing.
- Add an offline test that sets the **plain** secret and runs `PlanOnly` against a mock server, to catch validator path errors. The validators attached to the plain attribute only resolve their `_wo` sibling when the plain attribute has a value. See `pkg/framework/objects/semantic_layer_credential/snowflake_sl_credential_unit_test.go`.
- Add a test with `testhelpers.SetupMockServer` that applies with the `_wo` attribute and asserts on the **captured request payload** (create, then update after bumping the version). Checking the state is not enough: the state is empty either way, which is how #732 went unnoticed for days. Break the resolver temporarily to confirm that the test fails.
- The mock responses must return nullable fields as explicit `null`, like the real API, otherwise `nullable.MustGet()` panics when the response is read.
- Regenerate the docs with `go generate .` and add a `.changes/unreleased` entry.
