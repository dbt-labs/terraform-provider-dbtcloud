# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_semantic_layer_credential_service_token_mapping.example
  id = "id"
}

import {
  to = dbtcloud_semantic_layer_credential_service_token_mapping.example
  id = "12345"
}

# using the older import command
terraform import dbtcloud_semantic_layer_credential_service_token_mapping.example "id"
terraform import dbtcloud_semantic_layer_credential_service_token_mapping.example 12345 