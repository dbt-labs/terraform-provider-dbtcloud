# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_postgres_semantic_layer_credential.example
  id = "credential_id"
}

import {
  to = dbtcloud_postgres_semantic_layer_credential.example
  id = "12345"
}

# using the older import command
terraform import dbtcloud_postgres_semantic_layer_credential.example "credential_id"
terraform import dbtcloud_postgres_semantic_layer_credential.example 12345 