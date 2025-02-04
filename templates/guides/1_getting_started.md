---
page_title: "1. Getting started"
subcategory: ""
---

# 1. Getting started

The example below shows a simple set up with:

- 1 dbt Cloud project, connected to:
  - a Snowflake Data Warehouse
  - a GitHub repository, via the native GitHub integration
- 2 dbt Cloud environments
  - Dev: For developers, with access to the IDE
  - Prod: For running jobs. Authentication in this example is with user/password

-> This example is not "production ready" and working with secrets like the data warehouse credentials need to be done in a more robust way but this should provide an overview of some of the key resources of the provider.

## Writing the configuration

In a file called `terraform.tfvars`, add the following variables.

Those will be used in the main configuration file

```terraform
dbt_account_id = your_account_id
dbt_token      = "your_token"
// for the dbt_host_url, the default is "https://cloud.getdbt.com/api" but it can be updated
dbt_host_url   = "https://emea.dbt.com/api"
```

In a file called `main.tf`, add the following content:

```terraform
// define the variables we will use
variable "dbt_account_id" {
  type = number
}

variable "dbt_token" {
  type = string
}

variable "dbt_host_url" {
  type = string
}


// initialize the provider and set the settings
terraform {
  required_providers {
    dbtcloud = {
      source  = "dbt-labs/dbtcloud"
      version = "0.3.23"
    }
  }
}

provider "dbtcloud" {
  account_id = var.dbt_account_id
  token      = var.dbt_token
  host_url   = var.dbt_host_url
}


// create a project
resource "dbtcloud_project" "my_project" {
  name = "My dbt project"
}


// create a global connection
resource "dbtcloud_global_connection" "my_connection" {
  name       = "My Snowflake warehouse"
  snowflake  = {  
    account    = "my-snowflake-account"
    database   = "MY_DATABASE"
    role       = "MY_ROLE"
    warehouse  = "MY_WAREHOUSE"
  }
}

// link a repository to the dbt Cloud project
// this example adds a github repo for which we know the installation_id but the resource docs have other examples
resource "dbtcloud_repository" "my_repository" {
  project_id             = dbtcloud_project.my_project.id
  remote_url             = "git@github.com:<github_org>/<github_repo>.git"
  github_installation_id = 9876
  git_clone_strategy     = "github_app"
}

resource "dbtcloud_project_repository" "my_project_repository" {
  project_id    = dbtcloud_project.my_project.id
  repository_id = dbtcloud_repository.my_repository.repository_id
}


// create 2 environments, one for Dev and one for Prod
// here both are linked to the same Data Warehouse connection
// for Prod, we need to create a credential as well
resource "dbtcloud_environment" "my_dev" {
  dbt_version     = "latest"
  name            = "Dev"
  project_id      = dbtcloud_project.my_project.id
  type            = "development"
  connection_id   = dbtcloud_global_connection.my_connection.id
}

resource "dbtcloud_environment" "my_prod" {
  dbt_version     = "latest"
  name            = "Prod"
  project_id      = dbtcloud_project.my_project.id
  type            = "deployment"
  deployment_type = "production"
  credential_id   = dbtcloud_snowflake_credential.prod_credential.credential_id
  connection_id   = dbtcloud_global_connection.my_connection.id
}

// we use user/password but there are other options on the resource docs
resource "dbtcloud_snowflake_credential" "prod_credential" {
  project_id  = dbtcloud_project.my_project.id
  auth_type   = "password"
  num_threads = 16
  schema      = "analytics"
  user        = "my_snowflake_user"
  // note, this is a simple example to get Terraform and dbt Cloud working, but do not store passwords in the config for a real productive use case
  // there are different strategies available to protect sensitive input: https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables
  password    = "my_snowflake_password"
}
```

## Running Terraform

Install Terraform on your machine if it is not done.

Once installed, open a terminal in the same folder as the 2 files created previously. We can then run the following commands

- `terraform init`
  - this will install the dbt Cloud provider
  - if you want to upgrade from a previous version, use `terraform init -upgrade`
- `terraform plan`
  - this command won't change any config in dbt Cloud but will let you know what changes are going to happen
- `terraform apply`
  - this command will compare the configuration and the actual configuration and apply changes to dbt Cloud. It will ask for approval before doing so
  - to skip the approval, use `terraform apply -auto-approve`

After those commands, the dbt Cloud config should be updated with your new project created as per the configuration

If you want to delete the project and all the configuration attached from dbt Cloud, you can use the command `dbt destroy`
