package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"
)

var (
	JobCompletionTriggerConditionsMappingCodeHuman = map[int]any{
		10: "success",
		20: "error",
		30: "canceled",
	}
)

var JobCompletionTriggerConditionsMappingHumanCode = lo.Invert(
	JobCompletionTriggerConditionsMappingCodeHuman,
)

var objectSchema = map[string]*schema.Schema{
	"job_id": {
		Type: schema.TypeInt,
	},
	"project_id": {
		Type: schema.TypeInt,
	},
	"statuses": {
		// we use TypeList here, just for moving from Map to Set
		// the resource parameter itself is a TypeSet so that duplicates are removed and order doesn't matter
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
}

func JobConditionMapToSet(item map[string]any) *schema.Set {
	// The hash function helps identify unique items in the set
	hashFunc := schema.HashResource(&schema.Resource{Schema: objectSchema})

	// Create a slice of maps as required by schema.NewSet
	items := []interface{}{item}

	return schema.NewSet(hashFunc, items)
}

func ExtractJobConditionSet(
	d *schema.ResourceData,
) (empty bool, jobID, projectID int, statuses []int) {

	if d.Get("job_completion_trigger_condition").(*schema.Set).Len() == 0 {
		return true, 0, 0, []int{}
	} else {
		// this is a set but we only allow 1 item
		jobCompletionTrigger := d.Get("job_completion_trigger_condition").(*schema.Set).List()[0].(map[string]any)

		jobCompletionStatuses := lo.Map(
			jobCompletionTrigger["statuses"].(*schema.Set).List(),
			func(status interface{}, idx int) int {
				return JobCompletionTriggerConditionsMappingHumanCode[status.(string)]
			},
		)
		return false, jobCompletionTrigger["job_id"].(int), jobCompletionTrigger["project_id"].(int), jobCompletionStatuses
	}
}
