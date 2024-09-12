---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "dbtcloud_projects Data Source - dbtcloud"
subcategory: ""
description: |-
  Retrieve all the projects created in dbt Cloud with an optional filter on parts of the project name.
---

# dbtcloud_projects (Data Source)

Retrieve all the projects created in dbt Cloud with an optional filter on parts of the project name.

## Example Usage

```terraform
// can be filtered by parts of the project name
data dbtcloud_projects my_acme_projects {
  name_contains = "acme"
}

// or can return all projects
data dbtcloud_projects my_projects {
  name_contains = "acme"
}

// this can be used to make sure that there are no distinct projects with the same names for example

locals {
  name_occurrences = {
    for project in data.dbtcloud_projects.my_projects.projects : project.name => project.id ...
  }
  duplicates_with_id = [
    for name, project_id in local.name_occurrences : "'${name}':${join(",", project_id)}" if length(project_id) > 1
  ]
}

check "no_different_projects_with_same_name" {
  assert {
    condition = length(local.duplicates_with_id) == 0
    error_message = "There are duplicate project names: ${join(" ; ", local.duplicates_with_id)}"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `name_contains` (String) Used to filter projects by name, Optional

### Read-Only

- `projects` (Attributes Set) Set of projects with their details (see [below for nested schema](#nestedatt--projects))

<a id="nestedatt--projects"></a>
### Nested Schema for `projects`

Read-Only:

- `connection` (Attributes) Details for the connection linked to the project (see [below for nested schema](#nestedatt--projects--connection))
- `created_at` (String) When the project was created
- `dbt_project_subdirectory` (String) Subdirectory for the dbt project inside the git repo
- `description` (String) Project description
- `id` (Number) Project ID
- `name` (String) Project name
- `repository` (Attributes) Details for the repository linked to the project (see [below for nested schema](#nestedatt--projects--repository))
- `semantic_layer_config_id` (Number) Semantic layer config ID
- `updated_at` (String) When the project was last updated

<a id="nestedatt--projects--connection"></a>
### Nested Schema for `projects.connection`

Read-Only:

- `adapter_version` (String) Version of the adapter for the connection. Will tell what connection type it is
- `id` (Number) Connection ID
- `name` (String) Connection name


<a id="nestedatt--projects--repository"></a>
### Nested Schema for `projects.repository`

Read-Only:

- `id` (Number) Repository ID
- `pull_request_url_template` (String) URL template for PRs
- `remote_url` (String) URL of the git repo remote