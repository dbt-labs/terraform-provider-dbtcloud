---
page_title: "2. Leveraging the Hashicorp HTTP provider"
subcategory: ""
---

# 2. Leveraging the HTTP provider

While this provider supports different resources and data sources from dbt Cloud, it doesn't cover all the API endpoints provided by the platform.

In the case that the current version of the provider doesn't support your use case, you can:

- raise an issue in [the provider GitHub repository](https://github.com/dbt-labs/terraform-provider-dbtcloud) explaining what resource/data source you would need to solve your use case
- and/or leverage the Hashicorp HTTP provider to directly call the dbt Cloud API from Terraform

## Using the HTTP provider to retrieve data

The [Hashicorp HTTP provider](https://registry.terraform.io/providers/hashicorp/http/latest/docs) can be used to query the dbt Cloud API manually.

The list of endpoints available in dbt Cloud can be found [at this page for v2](https://docs.getdbt.com/dbt-cloud/api-v2#/) and [at that page for v3](https://docs.getdbt.com/dbt-cloud/api-v3#/). Please note that v2 and v3 have different endpoints, and depending on your use case you might want to use one or the other.

In the example below:

- we query the `GET /groups/` endpoint to retrieve the list of groups
- we parse the results sent back from the API
- we filter the groups to retrieve only the one with the name "Owner"
- we store its id in `local.owner_group_id`

```terraform
data "http" "example" {
  url = "${var.dbt_host_url}/v3/accounts/${var.dbt_account_id}/groups/"
  request_headers = {
    Authorization = "Token ${var.dbt_token}"
  }
}


locals {
  owner_group_name = "Owner"
  groups = jsondecode(data.http.example.response_body)
  owner_groups = [for group in local.groups.data: group.id if group.name == local.owner_group_name]
  owner_group_id = length(local.owner_groups) > 0 ? local.owner_groups[0] : "No owner group found"
}
```

This same technique can be used with any of the `GET` endpoints of the dbt Cloud API.
