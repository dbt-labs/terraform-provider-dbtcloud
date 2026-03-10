package codegen

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Generate reads a YAML resource definition and produces a Go command file.
func Generate(yamlPath, outputDir string) error {
	res, err := LoadYAMLResourceDef(yamlPath)
	if err != nil {
		return err
	}
	return GenerateFromDef(res, outputDir)
}

// LoadYAMLResourceDef reads a YAML file and returns the parsed ResourceDef.
func LoadYAMLResourceDef(yamlPath string) (*ResourceDef, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", yamlPath, err)
	}

	var res ResourceDef
	if err := yaml.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", yamlPath, err)
	}

	return &res, nil
}

// GenerateFromDef produces a Go command file from a ResourceDef.
func GenerateFromDef(res *ResourceDef, outputDir string) error {
	if res.ListType == "" {
		res.ListType = res.GoType
	}

	funcMap := template.FuncMap{
		"title": func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToUpper(s[:1]) + s[1:]
		},
		"lower":          strings.ToLower,
		"flagGetter":     flagGetter,
		"flagType":       flagType,
		"flagDefault":    flagDefault,
		"fieldFormatExpr": func(f Field, varName string, goType string) string {
			return fieldFormatExpr(f, varName, goType)
		},
		"fieldTypeFor": func(structName, fieldName string) string {
			key := structName + "." + fieldName
			if t, ok := FieldTypeMap[key]; ok {
				return t
			}
			return "string"
		},
		"needsStrconv": func(res ResourceDef) bool {
			// Check if any field formatting needs strconv
			for _, f := range append(res.Columns, res.Details...) {
				goType := FieldTypeMap[res.GoType+"."+f.GoField]
				if goType == "" {
					goType = FieldTypeMap[res.ListType+"."+f.GoField]
				}
				if goType == "*int" || goType == "*int64" || goType == "*bool" || goType == "int" || goType == "int64" || goType == "bool" {
					return true
				}
			}
			return false
		},
		"needsStrings": func(res ResourceDef) bool {
			for _, f := range append(res.Columns, res.Details...) {
				if f.Format == "join" {
					return true
				}
			}
			return false
		},
		"needsJSON": func(res ResourceDef) bool {
			if res.List != nil && res.List.Mode == "paginate" {
				return true
			}
			for _, f := range append(res.Columns, res.Details...) {
				if f.Format == "json" {
					return true
				}
			}
			return false
		},
		"needsFmt": func(_ ResourceDef) bool {
			return true
		},
		"trimTrailingNewline": func(s string) string {
			return strings.TrimRight(s, "\n")
		},
		"indent": func(n int, s string) string {
			pad := strings.Repeat("\t", n)
			lines := strings.Split(s, "\n")
			for i, line := range lines {
				if line != "" {
					lines[i] = pad + line
				}
			}
			return strings.Join(lines, "\n")
		},
		"hasSliceFlag": func(flags []Flag) bool {
			for _, f := range flags {
				if f.GoType == "string_slice" {
					return true
				}
			}
			return false
		},
		"callSetup": func(s string) string {
			s = strings.TrimRight(s, "\n")
			lines := strings.Split(s, "\n")
			lastClientIdx := -1
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "client.") {
					lastClientIdx = i
				}
			}
			if lastClientIdx <= 0 {
				return ""
			}
			return strings.Join(lines[:lastClientIdx], "\n")
		},
		"callExpr": func(s string) string {
			s = strings.TrimRight(s, "\n")
			lines := strings.Split(s, "\n")
			lastClientIdx := -1
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "client.") {
					lastClientIdx = i
				}
			}
			if lastClientIdx < 0 {
				return s
			}
			return strings.Join(lines[lastClientIdx:], "\n")
		},
	}

	tmplContent, err := os.ReadFile("internal/codegen/resource.go.tmpl")
	if err != nil {
		return fmt.Errorf("reading template: %w", err)
	}

	tmpl, err := template.New("resource").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, res); err != nil {
		return fmt.Errorf("executing template for %s: %w", res.Name, err)
	}

	fileBase := res.Name
	if res.FuncPrefix != "" {
		fileBase = camelToSnake(res.FuncPrefix)
	}
	outPath := filepath.Join(outputDir, fileBase+"_gen.go")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("creating output dir: %w", err)
	}
	if err := os.WriteFile(outPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", outPath, err)
	}

	// Run goimports to fix imports and formatting
	if goimports, err := exec.LookPath("goimports"); err == nil {
		cmd := exec.Command(goimports, "-w", outPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("goimports failed on %s: %s\n%w", outPath, out, err)
		}
	} else {
		// Fall back to gofmt
		cmd := exec.Command("gofmt", "-w", outPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("gofmt failed on %s: %s\n%w", outPath, out, err)
		}
	}

	return nil
}

// flagGetter returns the urfave/cli getter expression for a flag type.
func flagGetter(f Flag) string {
	switch f.GoType {
	case "int":
		return fmt.Sprintf(`cmd.Int("%s")`, f.Name)
	case "bool":
		return fmt.Sprintf(`cmd.Bool("%s")`, f.Name)
	case "string_slice":
		return fmt.Sprintf(`cmd.StringSlice("%s")`, f.Name)
	default:
		return fmt.Sprintf(`cmd.String("%s")`, f.Name)
	}
}

// flagType returns the urfave/cli flag type name.
func flagType(f Flag) string {
	switch f.GoType {
	case "int":
		return "IntFlag"
	case "bool":
		return "BoolFlag"
	case "string_slice":
		return "StringSliceFlag"
	default:
		return "StringFlag"
	}
}

// flagDefault returns the Go default value expression for a flag.
func flagDefault(f Flag) string {
	if f.Default == "" {
		return ""
	}
	switch f.GoType {
	case "int":
		return f.Default
	case "bool":
		return f.Default
	default:
		return fmt.Sprintf(`"%s"`, f.Default)
	}
}

// fieldFormatExpr returns a Go expression that converts a struct field to string for display.
func fieldFormatExpr(f Field, varName string, goType string) string {
	accessor := varName + "." + f.GoField

	// Special formats
	switch f.Format {
	case "secret":
		return `"*****"`
	case "state":
		return fmt.Sprintf("stateLabel(%s)", accessor)
	case "bool":
		if goType == "bool" {
			return fmt.Sprintf("strconv.FormatBool(%s)", accessor)
		}
	case "join":
		return fmt.Sprintf(`strings.Join(%s, ", ")`, accessor)
	case "json":
		// Handled via pre-computation in template (jsonMarshal variable)
		return fmt.Sprintf("%sJSON", strings.ToLower(f.GoField))
	}

	// Auto-detect from Go type
	switch goType {
	case "*int":
		// Need nil check - handled in template
		return fmt.Sprintf("strconv.Itoa(*%s)", accessor)
	case "*string":
		// Need nil check - handled in template
		return fmt.Sprintf("*%s", accessor)
	case "int":
		return fmt.Sprintf("strconv.Itoa(%s)", accessor)
	case "int64":
		return fmt.Sprintf("strconv.FormatInt(%s, 10)", accessor)
	case "bool":
		return fmt.Sprintf("strconv.FormatBool(%s)", accessor)
	case "[]string":
		return fmt.Sprintf(`strings.Join(%s, ", ")`, accessor)
	default:
		return accessor
	}
}
