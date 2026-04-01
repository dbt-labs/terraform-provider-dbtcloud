package utils

import (
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
