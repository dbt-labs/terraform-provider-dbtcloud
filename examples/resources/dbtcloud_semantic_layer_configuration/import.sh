# using import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_semantic_layer_configuration.example
  id = "project_id:id"
}

import {
  to = dbtcloud_semantic_layer_configuration.example
  id = "12345:5678"
}

# using the older import command
terraform import dbtcloud_semantic_layer_configuration.example "project_id:id"
terraform import dbtcloud_semantic_layer_configuration.example 12345:5678 