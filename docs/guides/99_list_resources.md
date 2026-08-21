---
page_title: "Relationships between resources"
subcategory: ""
---

# Relationships between resources

The diagrams below show the different resources available in the provider and their relationships to each other.

As the provider has grown, a single diagram became too large to read, so the resources are now broken down into four diagrams by area. Resources shown with a dashed grey border belong to a different diagram than the one they appear in and are included as a reference so you can see how the areas connect; click on a diagram to see a larger version of it.

## Account & access

Groups, permissions, users, service tokens, SCIM, and other account-level settings.

[![Terraform resources - Account and access](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_account_access.png?raw=true)](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_account_access.png)

## Connections & credentials

Global connections, warehouse credentials, semantic layer credentials, and platform metadata credentials.

[![Terraform resources - Connections and credentials](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_connections_credentials.png?raw=true)](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_connections_credentials.png)

## Project & environment setup

Projects, environments, repositories, profiles, and related project-level settings.

[![Terraform resources - Project and environment setup](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_project_environment.png?raw=true)](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_project_environment.png)

## Jobs & orchestration

Jobs, triggers, notifications, webhooks, and runs.

[![Terraform resources - Jobs and orchestration](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_jobs_orchestration.png?raw=true)](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources_jobs_orchestration.png)

## Previous versions

Older versions of the provider documented a single combined diagram, [still available here](https://github.com/dbt-labs/terraform-provider-dbtcloud/blob/main/terraform_resources.png), for reference.
