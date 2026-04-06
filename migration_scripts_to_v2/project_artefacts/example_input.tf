resource "dbtcloud_project_artefacts" "prod_artefacts" {
  project_id       = 1234
  docs_job_id      = 5678
  freshness_job_id = 9012
}

resource "dbtcloud_project_artefacts" "staging_artefacts" {
  project_id  = 2345
  docs_job_id = 6789
}
