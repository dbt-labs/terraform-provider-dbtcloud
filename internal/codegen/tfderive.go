package codegen

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// TFResourceConfig holds the inputs needed to derive a ResourceDef from TF schemas.
type TFResourceConfig struct {
	Name          string                            // resource name (e.g., "project")
	GoType        string                            // Go struct for detail (e.g., "Project")
	ListType      string                            // Go struct for list (e.g., "ProjectConnectionRepository"), defaults to GoType
	ResourceAttrs []TFAttrInfo                      // from ExtractResourceAttrs
	FilterAttrs   []TFAttrInfo                      // from ExtractDatasourceFilterAttrs
	GoTypeTags    map[string]GoFieldInfo             // json tag → GoFieldInfo for GoType
	ListTypeTags  map[string]GoFieldInfo             // json tag → GoFieldInfo for ListType
	AllStructTags map[string]map[string]GoFieldInfo  // all struct tags (for resolving nested fields)
	Methods       []*MethodInfo                      // scanned client methods
	Description   string                             // from TF resource schema
}

// DeriveTFResourceDef builds a ResourceDef from TF schemas and method scanning.
func DeriveTFResourceDef(cfg TFResourceConfig) (*ResourceDef, error) {
	if cfg.ListType == "" {
		cfg.ListType = cfg.GoType
	}

	res := &ResourceDef{
		Name:        cfg.Name,
		Description: cfg.Description,
		GoType:      cfg.GoType,
		ListType:    cfg.ListType,
	}

	// Derive display fields from TF resource attrs
	fields := deriveDisplayFields(cfg.ResourceAttrs, cfg.GoTypeTags)
	res.Columns = fields
	res.Details = fields

	// Process methods
	for _, m := range cfg.Methods {
		switch m.Kind {
		case "get":
			op, err := deriveGetOp(m, cfg.GoType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skipping %s %s: %v\n", m.Kind, m.Name, err)
				continue
			}
			res.Get = op
		case "list":
			listOp, listType, err := deriveTFListOp(m, cfg.GoType, cfg.FilterAttrs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skipping %s %s: %v\n", m.Kind, m.Name, err)
				continue
			}
			res.List = listOp
			if listType != "" {
				res.ListType = listType
			}
		case "create":
			op, err := deriveTFCreateOp(m, cfg.GoType, cfg.ResourceAttrs, cfg.GoTypeTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skipping %s %s: %v\n", m.Kind, m.Name, err)
				continue
			}
			res.Create = op
		case "update":
			updateOp, err := deriveTFUpdateOp(m, cfg.GoType, cfg.ResourceAttrs, cfg.GoTypeTags, cfg.AllStructTags)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skipping %s %s: %v\n", m.Kind, m.Name, err)
				continue
			}
			res.Update = updateOp
		case "delete":
			op, err := deriveDeleteOp(m, cfg.GoType)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  skipping %s %s: %v\n", m.Kind, m.Name, err)
				continue
			}
			res.Delete = op
		}
	}

	// Ensure list uses ListType for columns if different from GoType
	if res.ListType != res.GoType && cfg.ListTypeTags != nil {
		listFields := deriveDisplayFields(cfg.ResourceAttrs, cfg.ListTypeTags)
		res.Columns = listFields
	}

	return res, nil
}

// deriveDisplayFields builds Fields from TF resource attrs matched to Go struct json tags.
func deriveDisplayFields(attrs []TFAttrInfo, tags map[string]GoFieldInfo) []Field {
	// Sort attrs for deterministic output
	sorted := make([]TFAttrInfo, len(attrs))
	copy(sorted, attrs)
	sort.Slice(sorted, func(i, j int) bool {
		return attrSortKey(sorted[i]) < attrSortKey(sorted[j])
	})

	var fields []Field
	for _, attr := range sorted {
		goInfo, ok := tags[attr.Name]
		if !ok {
			continue
		}

		format := deriveFormat(attr, goInfo)
		label := snakeToLabel(attr.Name)

		fields = append(fields, Field{
			GoField: goInfo.GoField,
			JSONTag: attr.Name,
			Label:   label,
			Format:  format,
		})
	}
	return fields
}

// attrSortKey returns a sort key that puts "id" first, then "name",
// then other fields alphabetically.
func attrSortKey(a TFAttrInfo) string {
	switch a.Name {
	case "id":
		return "00_id"
	case "name":
		return "01_name"
	case "description":
		return "02_description"
	default:
		return "10_" + a.Name
	}
}

// deriveFormat determines the display format for a field based on TF attr type and Go type.
func deriveFormat(attr TFAttrInfo, goInfo GoFieldInfo) string {
	// Sensitive fields → masked display
	if attr.Sensitive {
		return "secret"
	}
	// Nested/block types → json format
	if attr.TFType == "nested" || attr.TFType == "block" {
		return "json"
	}
	// Use the existing autoFormat logic for simple types
	return autoFormat(goInfo.GoField, goInfo.GoType)
}

// deriveTFListOp creates a List operation using method scanner + TF filter attrs.
func deriveTFListOp(m *MethodInfo, resourceName string, filterAttrs []TFAttrInfo) (*ListOp, string, error) {
	for _, p := range m.Params {
		if !isSimpleParamType(p.GoType) {
			return nil, "", fmt.Errorf("complex param type %s in list method", p.GoType)
		}
	}

	// Build flags from method params, using TF descriptions when available
	var flags []Flag
	for _, p := range m.Params {
		f, err := paramToFlag(p, resourceName, false)
		if err != nil {
			return nil, "", err
		}
		// Try to find a matching TF filter attr for better usage text
		if desc := findFilterDescription(f.Name, filterAttrs); desc != "" {
			f.Usage = desc
		}
		flags = append(flags, *f)
	}

	call := buildCallExpr("client."+m.Name, m.Params, resourceName, false)

	listType := ""
	if after, ok := strings.CutPrefix(m.ReturnType, "[]"); ok {
		listType = after
	}

	return &ListOp{
		Mode:  "method",
		Flags: flags,
		Call:  call,
	}, listType, nil
}

// findFilterDescription looks for a TF filter attr matching the given flag name.
func findFilterDescription(flagName string, filterAttrs []TFAttrInfo) string {
	// Convert flag name (kebab) to snake for matching
	snake := strings.ReplaceAll(flagName, "-", "_")
	for _, attr := range filterAttrs {
		if attr.Name == snake {
			return attr.Description
		}
	}
	return ""
}

// deriveTFCreateOp creates a Create operation by matching method params to TF schema attrs
// via Go struct field json tags. This allows us to get descriptions and required/optional
// from the TF schema.
func deriveTFCreateOp(m *MethodInfo, resourceName string, tfAttrs []TFAttrInfo, structTags map[string]GoFieldInfo) (*Operation, error) {
	for _, p := range m.Params {
		if !isSimpleParamType(p.GoType) {
			return nil, fmt.Errorf("complex param type %s in create method %s", p.GoType, m.Name)
		}
	}

	// Build lookups for matching params to TF attrs:
	// 1. Go field name (lowered) → TF attr info
	goFieldToTF := map[string]TFAttrInfo{}
	for _, attr := range tfAttrs {
		if goInfo, ok := structTags[attr.Name]; ok {
			goFieldToTF[strings.ToLower(goInfo.GoField)] = attr
		}
	}
	// 2. TF attr name → TF attr info (for direct snake_case matching)
	tfAttrByName := map[string]TFAttrInfo{}
	for _, attr := range tfAttrs {
		tfAttrByName[attr.Name] = attr
	}

	var flags []Flag
	for _, p := range m.Params {
		f, err := paramToFlag(p, resourceName, true)
		if err != nil {
			return nil, err
		}

		// Try to match param to a TF attr:
		// 1. Via Go struct field name (case-insensitive)
		// 2. Via direct conversion of param name to snake_case
		matched := false
		if tfAttr, ok := goFieldToTF[strings.ToLower(p.Name)]; ok {
			f.Required = tfAttr.Required
			if tfAttr.Description != "" {
				f.Usage = tfAttr.Description
			}
			matched = true
		}
		if !matched {
			snake := camelToSnake(p.Name)
			if tfAttr, ok := tfAttrByName[snake]; ok {
				f.Required = tfAttr.Required
				if tfAttr.Description != "" {
					f.Usage = tfAttr.Description
				}
			}
		}

		flags = append(flags, *f)
	}

	call := buildCallExpr("client."+m.Name, m.Params, resourceName, true)
	return &Operation{
		Flags:     flags,
		Call:      call,
		ResultVar: strings.ToLower(resourceName),
	}, nil
}

// deriveTFUpdateOp creates an Update operation using TF schema to determine editable fields.
func deriveTFUpdateOp(m *MethodInfo, resourceName string, tfAttrs []TFAttrInfo, structTags map[string]GoFieldInfo, allStructTags map[string]map[string]GoFieldInfo) (*UpdateOp, error) {
	if len(m.Params) == 0 {
		return nil, fmt.Errorf("update method has no params")
	}

	lastParam := m.Params[len(m.Params)-1]
	if lastParam.GoType != resourceName {
		return nil, fmt.Errorf("update method last param is not the struct type")
	}

	idParams := m.Params[:len(m.Params)-1]
	for _, p := range idParams {
		if !isSimpleParamType(p.GoType) {
			return nil, fmt.Errorf("complex param type %s in update method", p.GoType)
		}
	}

	flags, err := paramsToFlags(idParams, resourceName, true)
	if err != nil {
		return nil, err
	}

	idExpr := buildIDExpr(idParams, resourceName)
	getMethodName := "Get" + resourceName
	getCall := buildGetCallExpr("client."+getMethodName, idParams, resourceName)
	updateCall := buildUpdateCallExpr("client."+m.Name, idParams, resourceName)

	// Derive editable fields from TF schema: Optional, non-Computed-only attrs
	editable := deriveTFEditableFields(tfAttrs, structTags, allStructTags)

	return &UpdateOp{
		Flags:      flags,
		IDExpr:     idExpr,
		GetCall:    getCall,
		UpdateCall: updateCall,
		Editable:   editable,
	}, nil
}

// camelToSnake converts a camelCase param name to snake_case.
// Examples: "isActive" → "is_active", "projectId" → "project_id", "type_" → "type"
func camelToSnake(name string) string {
	name = strings.TrimRight(name, "_")
	kebab := camelToKebab(name)
	return strings.ReplaceAll(kebab, "-", "_")
}

// deriveTFEditableFields builds editable fields for update from TF schema attrs.
func deriveTFEditableFields(tfAttrs []TFAttrInfo, structTags map[string]GoFieldInfo, allStructTags map[string]map[string]GoFieldInfo) []EditableField {
	var editable []EditableField

	// Sort for deterministic output
	sorted := make([]TFAttrInfo, len(tfAttrs))
	copy(sorted, tfAttrs)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, attr := range sorted {
		// Skip computed-only attrs (not user-settable)
		if attr.Computed && !attr.Optional && !attr.Required {
			continue
		}
		// Skip block types (e.g., ListNestedBlock — too complex)
		if attr.TFType == "block" {
			continue
		}
		// Skip id
		if attr.Name == "id" {
			continue
		}

		// Flatten nested attributes (SingleNestedAttribute with sub-attrs)
		if attr.TFType == "nested" && len(attr.SubAttrs) > 0 {
			editable = append(editable, flattenNestedEditable(attr, structTags, allStructTags)...)
			continue
		}

		goInfo, ok := structTags[attr.Name]
		if !ok {
			continue
		}

		flagType := goTypeToEditableType(goInfo.GoType)
		if flagType == "" {
			continue
		}

		editable = append(editable, EditableField{
			Flag:    camelToKebab(goInfo.GoField),
			GoField: goInfo.GoField,
			GoType:  flagType,
			Usage:   attr.Description,
		})
	}
	return editable
}

// goTypeToEditableType maps a Go type to an editable field type string.
// Returns "" for unsupported types.
func goTypeToEditableType(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "*string":
		return "*string"
	case "int":
		return "int"
	case "*int":
		return "*int"
	case "bool":
		return "bool"
	case "[]string":
		return "string_slice"
	default:
		return ""
	}
}

// flattenNestedEditable expands a nested TF attribute into individual editable fields
// using dotted GoField paths (e.g., "Triggers.Schedule").
func flattenNestedEditable(attr TFAttrInfo, structTags map[string]GoFieldInfo, allStructTags map[string]map[string]GoFieldInfo) []EditableField {
	// Look up the parent Go field to find the nested struct type
	parentInfo, ok := structTags[attr.Name]
	if !ok {
		return nil
	}

	// Strip pointer prefix to look up struct tags (e.g., "*JobCompletionTrigger" → "JobCompletionTrigger")
	nestedTypeName := strings.TrimPrefix(parentInfo.GoType, "*")

	// Skip pointer-to-struct types: setting fields on a nil pointer would panic at runtime.
	// Only flatten non-pointer struct types where the nested value is always populated.
	if nestedTypeName != parentInfo.GoType {
		return nil
	}

	// Find the nested struct's tags
	nestedTags := allStructTags[nestedTypeName]
	if nestedTags == nil {
		return nil
	}

	// Sort sub-attrs for deterministic output
	subSorted := make([]TFAttrInfo, len(attr.SubAttrs))
	copy(subSorted, attr.SubAttrs)
	sort.Slice(subSorted, func(i, j int) bool {
		return subSorted[i].Name < subSorted[j].Name
	})

	var editable []EditableField
	for _, sub := range subSorted {
		if sub.Computed && !sub.Optional && !sub.Required {
			continue
		}

		subInfo, ok := nestedTags[sub.Name]
		if !ok {
			continue
		}

		flagType := goTypeToEditableType(subInfo.GoType)
		if flagType == "" {
			continue
		}

		editable = append(editable, EditableField{
			Flag:    camelToKebab(parentInfo.GoField) + "-" + camelToKebab(subInfo.GoField),
			GoField: parentInfo.GoField + "." + subInfo.GoField,
			GoType:  flagType,
			Usage:   sub.Description,
		})
	}
	return editable
}

// ResolveExtraColumns resolves extra columns that use JSONTag into concrete GoField names
// for a specific struct. Fields that already have GoField set are returned as-is.
// Fields whose JSONTag doesn't match the struct are dropped.
func ResolveExtraColumns(extras []Field, tags map[string]GoFieldInfo) []Field {
	var resolved []Field
	for _, f := range extras {
		if f.JSONTag == "" {
			// Already has a GoField, use as-is
			resolved = append(resolved, f)
			continue
		}
		goInfo, ok := tags[f.JSONTag]
		if !ok {
			continue
		}
		resolved = append(resolved, Field{
			GoField: goInfo.GoField,
			Label:   f.Label,
			Format:  resolveExtraFormat(f, goInfo),
		})
	}
	return resolved
}

func resolveExtraFormat(f Field, goInfo GoFieldInfo) string {
	if f.Format != "" {
		return f.Format
	}
	return autoFormat(goInfo.GoField, goInfo.GoType)
}
