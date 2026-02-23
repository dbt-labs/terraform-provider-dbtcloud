package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"reflect"
	"strings"
)

// GoFieldInfo holds the Go field name and type for a struct field.
type GoFieldInfo struct {
	GoField string // Go field name (e.g., "DbtProjectSubdirectory")
	GoType  string // Go type string (e.g., "*string", "int64")
}

// ParseStructTags parses all Go files in dir and returns a map of
// struct name → json tag → GoFieldInfo. Embedded structs are resolved
// so that e.g. JobWithEnvironment inherits Job's fields.
func ParseStructTags(dir string) (map[string]map[string]GoFieldInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}

	type structData struct {
		fields map[string]GoFieldInfo // json tag → GoFieldInfo
		embeds []string               // embedded struct names
	}

	structs := map[string]*structData{}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					continue
				}
				for _, spec := range genDecl.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}
					name := typeSpec.Name.Name
					sd := &structData{
						fields: map[string]GoFieldInfo{},
					}

					for _, field := range structType.Fields.List {
						if len(field.Names) == 0 {
							// Embedded field
							typeName := exprToString(field.Type)
							sd.embeds = append(sd.embeds, typeName)
							continue
						}

						goField := field.Names[0].Name
						goType := exprToString(field.Type)

						// Extract json tag
						jsonTag := extractJSONTag(field)
						if jsonTag == "" || jsonTag == "-" {
							continue
						}

						sd.fields[jsonTag] = GoFieldInfo{
							GoField: goField,
							GoType:  goType,
						}
					}

					structs[name] = sd
				}
			}
		}
	}

	// Resolve embeds: copy fields from embedded structs
	result := map[string]map[string]GoFieldInfo{}
	for name, sd := range structs {
		merged := map[string]GoFieldInfo{}
		// First add embedded fields
		for _, embed := range sd.embeds {
			if embedded, ok := structs[embed]; ok {
				maps.Copy(merged, embedded.fields)
			}
		}
		// Then add own fields (override embedded)
		maps.Copy(merged, sd.fields)
		result[name] = merged
	}

	return result, nil
}

// extractJSONTag extracts the json tag name from a struct field.
// Returns the tag name (before any comma), or "" if no json tag.
func extractJSONTag(field *ast.Field) string {
	if field.Tag == nil {
		return ""
	}
	// field.Tag.Value includes backticks: `json:"name,omitempty"`
	tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
	jsonTag := tag.Get("json")
	if jsonTag == "" {
		return ""
	}
	// Split on comma, take the name part
	if idx := strings.Index(jsonTag, ","); idx != -1 {
		return jsonTag[:idx]
	}
	return jsonTag
}
