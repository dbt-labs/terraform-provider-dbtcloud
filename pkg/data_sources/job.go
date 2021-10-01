package data_sources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var jobSchema = map[string]*schema.Schema{
	"project_id": &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	},
	"environment_id": &schema.Schema{
		Type:     schema.TypeInt,
		Computed: true,
	},
	"name": &schema.Schema{
		Type:     schema.TypeString,
		Computed: true,
	},
	"job_id": &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	},
}

func DatasourceJob() *schema.Resource {
	return &schema.Resource{
		Read:   dataSourceGithubUserRead,
		Schema: jobSchema,
	}
}

func datasourceJobRead(d *schema.ResourceData, m interface{}) diag.Diagnostics {
	token := d.Get("token").(string)
	account_id := d.Get("account_id").(int)
	job_id := d.Get("job_id").(int)

	url := fmt.Sprintf("https://cloud.getdbt.com/api/v2/accounts/%s/jobs/%s", account_id, job_id)
	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)

	if err != nil {
		log.Printf("Error reading job %s", job_id)
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	job := new(dbt_cloud.Job)
	err = json.NewDecoder(resp.Body).Decode(&job)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("job", job); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(job_id, 10))

	return diags
}
