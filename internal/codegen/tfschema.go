package codegen

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasource_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resource_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// TFAttrInfo holds metadata about a single TF schema attribute.
type TFAttrInfo struct {
	Name        string       // snake_case attribute name (e.g., "dbt_project_subdirectory")
	TFType      string       // "string", "int64", "bool", "list_string", "nested", "block"
	Required    bool
	Optional    bool
	Computed    bool
	Sensitive   bool
	Description string
	SubAttrs    []TFAttrInfo // populated for nested types (SingleNestedAttribute)
}

// ExtractResourceAttrs extracts top-level attribute info from a TF resource schema.
func ExtractResourceAttrs(r resource.Resource) ([]TFAttrInfo, error) {
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)

	var attrs []TFAttrInfo
	for name, attr := range resp.Schema.Attributes {
		info := extractResourceAttr(name, attr)
		if info != nil {
			attrs = append(attrs, *info)
		}
	}
	for name, block := range resp.Schema.Blocks {
		info := extractResourceBlock(name, block)
		if info != nil {
			attrs = append(attrs, *info)
		}
	}
	return attrs, nil
}

// extractResourceAttr extracts info from a single resource schema attribute.
func extractResourceAttr(name string, attr resource_schema.Attribute) *TFAttrInfo {
	switch a := attr.(type) {
	case resource_schema.StringAttribute:
		return &TFAttrInfo{
			Name: name, TFType: "string",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Sensitive: a.Sensitive,
			Description: a.Description,
		}
	case resource_schema.Int64Attribute:
		return &TFAttrInfo{
			Name: name, TFType: "int64",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	case resource_schema.BoolAttribute:
		return &TFAttrInfo{
			Name: name, TFType: "bool",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	case resource_schema.ListAttribute:
		return &TFAttrInfo{
			Name: name, TFType: "list_string",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	case resource_schema.SingleNestedAttribute:
		var subAttrs []TFAttrInfo
		for subName, subAttr := range a.Attributes {
			if info := extractResourceAttr(subName, subAttr); info != nil {
				subAttrs = append(subAttrs, *info)
			}
		}
		return &TFAttrInfo{
			Name: name, TFType: "nested",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
			SubAttrs:    subAttrs,
		}
	case resource_schema.SetNestedAttribute:
		return &TFAttrInfo{
			Name: name, TFType: "nested",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	default:
		return nil
	}
}

// extractResourceBlock extracts info from a resource schema block (e.g., ListNestedBlock).
func extractResourceBlock(name string, block resource_schema.Block) *TFAttrInfo {
	switch b := block.(type) {
	case resource_schema.ListNestedBlock:
		return &TFAttrInfo{
			Name: name, TFType: "block",
			Description: b.Description,
		}
	default:
		return nil
	}
}

// ExtractDatasourceFilterAttrs extracts non-nested, non-Computed-only attributes
// from a plural datasource schema. These are the filter flags for list operations.
func ExtractDatasourceFilterAttrs(ds datasource.DataSource) ([]TFAttrInfo, error) {
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), req, resp)

	var attrs []TFAttrInfo
	for name, attr := range resp.Schema.Attributes {
		info := extractDatasourceFilterAttr(name, attr)
		if info != nil {
			attrs = append(attrs, *info)
		}
	}
	return attrs, nil
}

// extractDatasourceFilterAttr extracts filter-relevant attrs from a datasource schema.
// Only returns simple (non-nested) attributes that are Optional (i.e., filter params).
func extractDatasourceFilterAttr(name string, attr datasource_schema.Attribute) *TFAttrInfo {
	switch a := attr.(type) {
	case datasource_schema.StringAttribute:
		if !a.Optional {
			return nil
		}
		return &TFAttrInfo{
			Name: name, TFType: "string",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	case datasource_schema.Int64Attribute:
		if !a.Optional {
			return nil
		}
		return &TFAttrInfo{
			Name: name, TFType: "int64",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	case datasource_schema.BoolAttribute:
		if !a.Optional {
			return nil
		}
		return &TFAttrInfo{
			Name: name, TFType: "bool",
			Required: a.Required, Optional: a.Optional, Computed: a.Computed,
			Description: a.Description,
		}
	default:
		return nil
	}
}

// GetResourceDescription extracts the top-level description from a TF resource schema.
// Truncates to the first sentence and escapes quotes for use in Go string literals.
func GetResourceDescription(r resource.Resource) string {
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)
	desc := resp.Schema.Description
	// Truncate to first sentence
	if idx := strings.Index(desc, "."); idx >= 0 {
		desc = desc[:idx+1]
	}
	// Escape double quotes
	desc = strings.ReplaceAll(desc, `"`, `\"`)
	return desc
}

// snakeToLabel converts a snake_case attribute name to a human-readable label.
// Examples: "id" → "ID", "project_id" → "Project ID", "dbt_version" → "Dbt Version"
func snakeToLabel(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = capitalizeWord(p)
	}
	return strings.Join(parts, " ")
}

// tfTypeToFlagType converts a TF attribute type to a CLI flag type string.
func tfTypeToFlagType(tfType string) string {
	switch tfType {
	case "int64":
		return "int"
	case "bool":
		return "bool"
	case "list_string":
		return "string_slice"
	default:
		return "string"
	}
}
