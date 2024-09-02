
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