package data_sources

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
)

var jobSchema = map[string]*schema.Schema{
	"project_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the project the job is in",
	},
	"environment_id": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the environment the job is in",
	},
	"name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Given name for the job",
	},
	"description": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Long description for the job",
	},
	"job_id": {
		Type:        schema.TypeInt,
		Required:    true,
		Description: "ID of the job",
	},
	"deferring_job_id": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the job this job defers to",
	},
	"deferring_environment_id": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "ID of the environment this job defers to",
	},
	"self_deferring": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether this job defers on a previous run of itself (overrides value in deferring_job_id)",
	},
	"triggers": {
		Type:     schema.TypeMap,
		Computed: true,
		Elem: &schema.Schema{
			Type:     schema.TypeBool,
			Optional: false,
			Default:  false,
		},
		Description: "Flags for which types of triggers to use, keys of github_webhook, git_provider_webhook, schedule, on_merge",
	},
	"timeout_seconds": {
		Type:        schema.TypeInt,
		Computed:    true,
		Description: "Number of seconds before the job times out",
	},
	"triggers_on_draft_pr": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the CI job should be automatically triggered on draft PRs",
	},
	"job_completion_trigger_condition": {
		Type:     schema.TypeSet,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"job_id": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The ID of the job that would trigger this job after completion.",
				},
				"project_id": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The ID of the project where the trigger job is running in.",
				},
				"statuses": {
					Type:        schema.TypeSet,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Computed:    true,
					Description: "List of statuses to trigger the job on.",
				},
			},
		},
		Description: "Which other job should trigger this job when it finishes, and on which conditions.",
	},
	"run_compare_changes": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "Whether the CI job should compare data changes introduced by the code change in the PR.",
	},
}

func DatasourceJob() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceJobRead,
		Schema:      jobSchema,
	}
}

func datasourceJobRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	c := m.(*dbt_cloud.Client)

	var diags diag.Diagnostics

	jobId := strconv.Itoa(d.Get("job_id").(int))

	job, err := c.GetJob(jobId)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project_id", job.ProjectId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment_id", job.EnvironmentId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", job.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", job.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("job_id", job.ID); err != nil {
		return diag.FromErr(err)
	}
	selfDeferring := job.DeferringJobId != nil && *job.DeferringJobId == *job.ID
	if !selfDeferring {
		if err := d.Set("deferring_job_id", job.DeferringJobId); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("deferring_environment_id", job.DeferringEnvironmentId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("self_deferring", selfDeferring); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout_seconds", job.Execution.TimeoutSeconds); err != nil {
		return diag.FromErr(err)
	}
	var triggers map[string]interface{}
	triggersInput, _ := json.Marshal(job.Triggers)
	json.Unmarshal(triggersInput, &triggers)
	if err := d.Set("triggers", triggers); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("triggers_on_draft_pr", job.TriggersOnDraftPR); err != nil {
		return diag.FromErr(err)
	}

	if job.JobCompletionTrigger == nil {
		if err := d.Set("job_completion_trigger_condition", nil); err != nil {
			return diag.FromErr(err)
		}
	} else {
		triggerCondition := job.JobCompletionTrigger.Condition
		// we convert the statuses from ID to human-readable strings
		statusesNames := lo.Map(triggerCondition.Statuses, func(status int, idx int) any {
			return utils.JobCompletionTriggerConditionsMappingCodeHuman[status]
		})
		triggerConditionMap := map[string]any{
			"job_id":     triggerCondition.JobID,
			"project_id": triggerCondition.ProjectID,
			"statuses":   statusesNames,
		}
		triggerConditionSet := utils.JobConditionMapToSet(triggerConditionMap)

		if err := d.Set("job_completion_trigger_condition", triggerConditionSet); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("run_compare_changes", job.RunCompareChanges); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(jobId)

	return diags
}
