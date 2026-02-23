package codegen

// MergeResourceDef merges an auto-derived ResourceDef with an optional YAML override.
//
// Two modes:
//   - Full override: if yamlOverride has GoType set, the YAML definition is used as-is.
//   - Partial override: only specified operations from YAML replace auto-derived ones.
//     Struct display (columns/details) always comes from the derived def.
func MergeResourceDef(derived *ResourceDef, yamlOverride *ResourceDef) *ResourceDef {
	if yamlOverride == nil {
		return derived
	}

	// Full override: YAML specifies go_type → use YAML entirely.
	if yamlOverride.GoType != "" {
		return yamlOverride
	}

	// Partial override: start with derived, replace operations from YAML.
	result := *derived // shallow copy

	// Override individual operations if specified in YAML.
	if yamlOverride.Get != nil {
		result.Get = yamlOverride.Get
	}
	if yamlOverride.List != nil {
		result.List = yamlOverride.List
	}
	if yamlOverride.Create != nil {
		result.Create = yamlOverride.Create
	}
	if yamlOverride.Update != nil {
		result.Update = yamlOverride.Update
	}
	if yamlOverride.Delete != nil {
		result.Delete = yamlOverride.Delete
	}

	// Override name and description if provided in YAML.
	if yamlOverride.Name != "" {
		result.Name = yamlOverride.Name
	}
	if yamlOverride.Description != "" {
		result.Description = yamlOverride.Description
	}

	// Handle ListType: YAML explicit value takes priority, then auto-derived.
	// When YAML overrides the list operation but doesn't specify a ListType,
	// the auto-derived ListType may be stale (from a different list method
	// that returns a wrapper type). Reset to GoType in that case.
	if yamlOverride.ListType != "" {
		result.ListType = yamlOverride.ListType
	} else if yamlOverride.List != nil {
		result.ListType = result.GoType
	}

	// Pass through extra columns for later resolution per struct type.
	if len(yamlOverride.ExtraColumns) > 0 {
		result.ExtraColumns = yamlOverride.ExtraColumns
	}

	// Apply label overrides to auto-derived columns and details.
	if len(yamlOverride.LabelOverrides) > 0 {
		applyLabelOverrides(result.Columns, yamlOverride.LabelOverrides)
		applyLabelOverrides(result.Details, yamlOverride.LabelOverrides)
	}

	return &result
}

// applyLabelOverrides replaces labels on fields whose JSONTag matches an override key.
func applyLabelOverrides(fields []Field, overrides map[string]string) {
	for i := range fields {
		if fields[i].JSONTag == "" {
			continue
		}
		if label, ok := overrides[fields[i].JSONTag]; ok {
			fields[i].Label = label
		}
	}
}
