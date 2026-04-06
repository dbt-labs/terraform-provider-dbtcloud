# Import an existing auth provider by its numeric ID.
# The ID can be found via the dbt Cloud API:
#   GET /api/v3/accounts/{account_id}/auth-provider/

terraform import dbtcloud_auth_provider.example 12345
