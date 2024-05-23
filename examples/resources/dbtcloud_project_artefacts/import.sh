# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_project_artefacts.my_artefacts
  id = "project_id"
}

import {
  to = dbtcloud_project_artefacts.my_artefacts
  id = "12345"
}

# using the older import command
terraform import dbtcloud_project_artefacts.my_artefacts "project_id"
terraform import dbtcloud_project_artefacts.my_artefacts 12345
