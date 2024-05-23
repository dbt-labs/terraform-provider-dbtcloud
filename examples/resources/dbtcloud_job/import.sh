# using  import blocks (requires Terraform >= 1.5)
import {
  to = dbtcloud_job.my_job
  id = "job_id"
}

import {
  to = dbtcloud_job.my_job
  id = "12345"
}

# using the older import command
terraform import dbtcloud_job.my_job "job_id"
terraform import dbtcloud_job.my_job 12345
