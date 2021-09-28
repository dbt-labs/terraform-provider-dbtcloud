package resources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/dbt_cloud"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var jobSchema = map[string]*schema.Schema{
	"id": &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	},
	"account_id": &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	},
	"project_id": &schema.Schema{
		Type:     schema.TypeInt,
		Required: true,
	},
	"name": &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
	},
}

func ResourceJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceJobCreate,
		Read:   resourceJobRead,

		Schema: jobSchema,
	}
}

func resourceJobRead(d *schema.ResourceData, m interface{}) error {
	token := d.Get("token").(string)

	account_id := d.Get("account_id").(int)
	job_id := d.Get("job_id").(int)

	if token != "" {
		url := fmt.Sprintf("https://cloud.getdbt.com/api/v2/accounts/%s/jobs/%s", account_id, job_id)
		req, err := http.NewRequest("GET", url, nil)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)

		if err != nil {
			log.Printf("Error reading job %s", job_id)
			return err
		}
		defer resp.Body.Close()

		job := new(dbt_cloud.Job)
		err = json.NewDecoder(resp.Body).Decode(&job)
		if err != nil {
			return err
		}

		return err
	}

	return nil
}

func resourceJobCreate(d *schema.ResourceData, m interface{}) error {
	token := d.Get("token").(string)

	account_id := d.Get("account_id").(int)
	project_id := d.Get("project_id").(int)
	environment_id := d.Get("environment_id").(int)
	name := d.Get("name").(string)
	execute_steps := d.Get("execute_steps").([]string)
	dbt_version := d.Get("dbt_version").(string)
	triggers := d.Get("triggers").(dbt_cloud.JobTrigger)
	settings := d.Get("settings").(dbt_cloud.JobSettings)
	state := d.Get("state").(int)
	generate_docs := d.Get("generate_docs").(bool)
	schedule := d.Get("schedule").(dbt_cloud.JobSchedule)

	if token != "" {
		newJob := dbt_cloud.JobData{
			Account_Id:     account_id,
			Project_Id:     project_id,
			Environment_Id: environment_id,
			Name:           name,
			Execute_Steps:  execute_steps,
			Dbt_Version:    dbt_version,
			Triggers:       triggers,
			Settings:       settings,
			State:          state,
			Generate_Docs:  generate_docs,
			Schedule:       schedule,
		}
		url := fmt.Sprintf("https://cloud.getdbt.com/api/v2/accounts/%s/jobs/", account_id)
		newJobData, err := json.Marshal(newJob)
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(newJobData))
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)

		if err != nil {
			log.Printf("Error creating job")
			return err
		}
		defer resp.Body.Close()

		job := new(dbt_cloud.Job)
		err = json.NewDecoder(resp.Body).Decode(&job)
		if err != nil {
			return err
		}

		return err
	}

	return nil
}
