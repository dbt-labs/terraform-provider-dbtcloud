// use dbt_cloud_project_artefacts instead of dbtcloud_project_artefacts for the legacy resource names
// legacy names will be removed from 0.3 onwards

resource "dbtcloud_project_artefacts" "my_project_artefacts" {
  project_id       = dbtcloud_project.dbt_project.id
  docs_job_id      = dbtcloud_job.prod_job.id
  freshness_job_id = dbtcloud_job.prod_job.id
}