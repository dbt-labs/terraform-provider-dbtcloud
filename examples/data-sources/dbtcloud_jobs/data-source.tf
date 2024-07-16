// we can search all jobs by project
data dbtcloud_jobs test_all_jobs_in_project {
  project_id = 1234
}

// or by environment
data dbtcloud_jobs test_all_jobs_in_environment {
  environment_id = 1234
}

// we can then retrieve all the jobs from the environment flagged as production
// this would include the jobs created by Terraform and the jobs created from the dbt Cloud UI
locals {
  my_jobs_prod = [for job in data.dbtcloud_jobs.test_all_jobs_in_project.jobs : job if job.environment.deployment_type == "production"]
}
