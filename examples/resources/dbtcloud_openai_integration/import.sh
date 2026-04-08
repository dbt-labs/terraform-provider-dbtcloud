# Import an existing OpenAI integration by its numeric ID.
# Note: key_value will be absent after import — the API never returns it.
# Use key_value_wo or key_value_wo_version to manage the key going forward.
terraform import dbtcloud_openai_integration.openai 12345
