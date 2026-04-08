resource "dbtcloud_azure_ad_application" "this" {
  organization_name = "my-azure-devops-org"
  client_id         = "00000000-0000-0000-0000-000000000000"
  client_secret     = var.azure_client_secret
  tenant_id         = "00000000-0000-0000-0000-000000000001"

  # Optional: defaults to "service_user". Set to "service_principal" to use
  # service principal authentication instead.
  azure_service_authentication_method = "service_user"
}

# NOTE: destroying this resource calls the dbt Cloud DELETE endpoint, which
# marks the record as inactive but does not remove the underlying database row.
# Re-creating the resource against the same account after a destroy will fail
# with a unique-constraint error. To recover, ask dbt Cloud support to remove
# the orphaned record, or use `terraform import` to re-adopt it:
#
#   terraform import dbtcloud_azure_ad_application.this <id>
