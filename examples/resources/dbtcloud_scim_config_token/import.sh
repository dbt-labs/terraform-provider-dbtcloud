# Import a SCIM config token by its numeric ID.
# Note: token_string will be empty after import — the API never returns the
# token value after creation. This is expected and documented behaviour.
terraform import dbtcloud_scim_config_token.okta 12345
