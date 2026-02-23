package codegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"
)

// MethodInfo holds metadata about a *Client method.
type MethodInfo struct {
	Name       string      // Method name (e.g., "GetProject")
	Params     []ParamInfo // Parameters (excluding receiver)
	ReturnType string      // Primary return type (e.g., "*Project", "[]Environment")
	Kind       string      // "get", "list", "create", "update", "delete"
}

// ParamInfo holds metadata about a method parameter.
type ParamInfo struct {
	Name   string // Parameter name (e.g., "projectID")
	GoType string // Go type string (e.g., "string", "int", "*bool")
}

// ScanMethods parses all Go files in dir and returns methods grouped by
// resource name. Only methods on *Client that match known CRUD patterns
// are included.
func ScanMethods(dir string) (map[string][]*MethodInfo, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, 0)
	if err != nil {
		return nil, err
	}

	result := map[string][]*MethodInfo{}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if !ok || funcDecl.Recv == nil {
					continue
				}

				// Check for *Client receiver.
				if !isClientReceiver(funcDecl.Recv) {
					continue
				}

				name := funcDecl.Name.Name
				resourceName, kind := classifyMethod(name)
				if resourceName == "" {
					continue
				}

				mi := &MethodInfo{
					Name: name,
					Kind: kind,
				}

				// Extract parameters.
				if funcDecl.Type.Params != nil {
					for _, field := range funcDecl.Type.Params.List {
						goType := exprToString(field.Type)
						for _, ident := range field.Names {
							mi.Params = append(mi.Params, ParamInfo{
								Name:   ident.Name,
								GoType: goType,
							})
						}
					}
				}

				// Extract primary return type.
				if funcDecl.Type.Results != nil && len(funcDecl.Type.Results.List) > 0 {
					mi.ReturnType = exprToString(funcDecl.Type.Results.List[0].Type)
				}

				result[resourceName] = append(result[resourceName], mi)
			}
		}
	}

	return result, nil
}

// isClientReceiver checks if the receiver is *Client.
func isClientReceiver(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	recvType := recv.List[0].Type
	star, ok := recvType.(*ast.StarExpr)
	if !ok {
		return false
	}
	ident, ok := star.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "Client"
}

// classifyMethod returns the resource name and kind for a method name.
// Returns ("", "") if the method doesn't match any known pattern.
func classifyMethod(name string) (resource string, kind string) {
	if strings.HasPrefix(name, "GetAll") {
		// GetAll{Name}s → list
		trimmed := strings.TrimPrefix(name, "GetAll")
		trimmed = strings.TrimSuffix(trimmed, "s")
		if trimmed == "" {
			return "", ""
		}
		return trimmed, "list"
	}
	if strings.HasPrefix(name, "Get") {
		trimmed := strings.TrimPrefix(name, "Get")
		if trimmed == "" || trimmed == "Endpoint" || trimmed == "Data" || trimmed == "RawData" {
			return "", ""
		}
		// GetProjectByName → not a standard get, skip
		if strings.Contains(trimmed, "By") {
			return "", ""
		}
		return trimmed, "get"
	}
	if strings.HasPrefix(name, "Create") {
		trimmed := strings.TrimPrefix(name, "Create")
		if trimmed == "" {
			return "", ""
		}
		return trimmed, "create"
	}
	if strings.HasPrefix(name, "Update") {
		trimmed := strings.TrimPrefix(name, "Update")
		if trimmed == "" {
			return "", ""
		}
		return trimmed, "update"
	}
	if strings.HasPrefix(name, "Delete") {
		trimmed := strings.TrimPrefix(name, "Delete")
		if trimmed == "" {
			return "", ""
		}
		return trimmed, "delete"
	}
	return "", ""
}

// deriveGetOp creates a Get operation from a method like GetProject(projectID string).
func deriveGetOp(m *MethodInfo, resourceName string) (*Operation, error) {
	flags, err := paramsToFlags(m.Params, resourceName, true)
	if err != nil {
		return nil, err
	}
	call := buildCallExpr("client."+m.Name, m.Params, resourceName, true)
	return &Operation{
		Flags:     flags,
		Call:      call,
		ResultVar: strings.ToLower(resourceName),
	}, nil
}

// deriveDeleteOp creates a Delete operation.
func deriveDeleteOp(m *MethodInfo, resourceName string) (*Operation, error) {
	for _, p := range m.Params {
		if !isSimpleParamType(p.GoType) {
			return nil, fmt.Errorf("complex param type %s in delete method", p.GoType)
		}
	}
	flags, err := paramsToFlags(m.Params, resourceName, true)
	if err != nil {
		return nil, err
	}
	call := buildCallExpr("client."+m.Name, m.Params, resourceName, true)
	return &Operation{
		Flags:     flags,
		Call:      call,
		ResultVar: strings.ToLower(resourceName),
	}, nil
}

// isSimpleParamType returns true if the type can be auto-converted to a CLI flag.
func isSimpleParamType(goType string) bool {
	switch goType {
	case "string", "int", "int64", "bool":
		return true
	case "[]string":
		return true
	}
	return false
}

// paramsToFlags converts method parameters to CLI flags.
func paramsToFlags(params []ParamInfo, resourceName string, required bool) ([]Flag, error) {
	var flags []Flag
	for _, p := range params {
		f, err := paramToFlag(p, resourceName, required)
		if err != nil {
			return nil, err
		}
		flags = append(flags, *f)
	}
	return flags, nil
}

// paramToFlag converts a single method parameter to a CLI flag.
func paramToFlag(p ParamInfo, resourceName string, required bool) (*Flag, error) {
	flagName := paramToFlagName(p.Name, resourceName)
	usage := fieldNameToLabel(p.Name)

	// String ID params: param named with "id"/"ID" and type is string → IntFlag.
	flagType := goTypeToFlagType(p.GoType)
	if p.GoType == "string" && isIDParam(p.Name) {
		flagType = "int"
	}

	return &Flag{
		Name:     flagName,
		GoType:   flagType,
		Required: required,
		Usage:    usage,
	}, nil
}

// paramToFlagName converts a parameter name to a CLI flag name.
// It applies camelToKebab and renames the resource's own ID param to just "id".
func paramToFlagName(paramName, resourceName string) string {
	// Check if this param is the resource's own ID.
	lower := strings.ToLower(paramName)
	resLower := strings.ToLower(resourceName)
	if lower == resLower+"id" {
		return "id"
	}

	// Check if the param matches the suffix of the resource name + "id".
	// e.g., for "SnowflakeCredential", suffix "Credential" → param "credentialId" → "id"
	if strings.HasSuffix(lower, "id") {
		paramRoot := strings.TrimSuffix(lower, "id")
		if strings.HasSuffix(resLower, paramRoot) {
			return "id"
		}
	}

	return camelToKebab(paramName)
}

// goTypeToFlagType converts a Go type to a flag type string.
func goTypeToFlagType(goType string) string {
	switch goType {
	case "int", "int64":
		return "int"
	case "bool":
		return "bool"
	case "[]string":
		return "string_slice"
	default:
		return "string"
	}
}

// buildCallExpr generates the Go call expression for an API method.
func buildCallExpr(method string, params []ParamInfo, resourceName string, _ bool) string {
	args := make([]string, len(params))
	for i, p := range params {
		args[i] = paramToCallArg(p, resourceName)
	}
	return method + "(" + strings.Join(args, ", ") + ")"
}

// paramToCallArg generates the Go expression for passing a flag value as a method argument.
func paramToCallArg(p ParamInfo, resourceName string) string {
	flagName := paramToFlagName(p.Name, resourceName)

	switch p.GoType {
	case "string":
		if isIDParam(p.Name) {
			// String ID param → read as int, convert to string.
			return fmt.Sprintf(`strconv.Itoa(int(cmd.Int("%s")))`, flagName)
		}
		return fmt.Sprintf(`cmd.String("%s")`, flagName)
	case "int":
		return fmt.Sprintf(`int(cmd.Int("%s"))`, flagName)
	case "int64":
		return fmt.Sprintf(`int64(cmd.Int("%s"))`, flagName)
	case "bool":
		return fmt.Sprintf(`cmd.Bool("%s")`, flagName)
	case "[]string":
		return fmt.Sprintf(`cmd.StringSlice("%s")`, flagName)
	default:
		return fmt.Sprintf(`cmd.String("%s")`, flagName)
	}
}

// buildIDExpr generates the Go expression for computing the ID variable in update operations.
func buildIDExpr(idParams []ParamInfo, resourceName string) string {
	if len(idParams) == 0 {
		return `""`
	}
	// Use the last ID param as the primary ID.
	p := idParams[len(idParams)-1]
	flagName := paramToFlagName(p.Name, resourceName)
	if p.GoType == "string" && isIDParam(p.Name) {
		return fmt.Sprintf(`strconv.Itoa(int(cmd.Int("%s")))`, flagName)
	}
	if p.GoType == "int" || p.GoType == "int64" {
		return fmt.Sprintf(`int(cmd.Int("%s"))`, flagName)
	}
	return fmt.Sprintf(`cmd.String("%s")`, flagName)
}

// buildGetCallExpr generates the get call for the update fetch step.
func buildGetCallExpr(method string, idParams []ParamInfo, resourceName string) string {
	if len(idParams) == 1 {
		return method + "(id)"
	}
	// Multiple ID params: all but last use cmd.Int directly, last uses `id`.
	args := make([]string, len(idParams))
	for i, p := range idParams {
		if i == len(idParams)-1 {
			args[i] = "id"
		} else {
			flagName := paramToFlagName(p.Name, resourceName)
			args[i] = fmt.Sprintf(`int(cmd.Int("%s"))`, flagName)
		}
	}
	return method + "(" + strings.Join(args, ", ") + ")"
}

// buildUpdateCallExpr generates the update call for the update step.
func buildUpdateCallExpr(method string, idParams []ParamInfo, resourceName string) string {
	if len(idParams) == 1 {
		return method + "(id, *existing)"
	}
	args := make([]string, len(idParams)+1)
	for i, p := range idParams {
		if i == len(idParams)-1 {
			args[i] = "id"
		} else {
			flagName := paramToFlagName(p.Name, resourceName)
			args[i] = fmt.Sprintf(`int(cmd.Int("%s"))`, flagName)
		}
	}
	args[len(idParams)] = "*existing"
	return method + "(" + strings.Join(args, ", ") + ")"
}

// isIDParam returns true if the parameter name suggests it's an ID parameter.
func isIDParam(name string) bool {
	lower := strings.ToLower(name)
	return strings.HasSuffix(lower, "id")
}

// camelToKebab converts a camelCase or PascalCase string to kebab-case.
// Examples: "projectId" → "project-id", "dbtVersion" → "dbt-version",
// "ID" → "id", "projectID" → "project-id", "type_" → "type"
func camelToKebab(name string) string {
	// Strip trailing underscores (Go convention for reserved words like type_)
	name = strings.TrimRight(name, "_")

	// Handle underscore-separated names first.
	if strings.Contains(name, "_") {
		parts := strings.Split(name, "_")
		for i, p := range parts {
			parts[i] = strings.ToLower(p)
		}
		return strings.Join(parts, "-")
	}

	var words []string
	runes := []rune(name)
	start := 0

	for i := 1; i < len(runes); i++ {
		if unicode.IsLower(runes[i-1]) && unicode.IsUpper(runes[i]) {
			words = append(words, string(runes[start:i]))
			start = i
		} else if i+1 < len(runes) && unicode.IsUpper(runes[i-1]) && unicode.IsUpper(runes[i]) && unicode.IsLower(runes[i+1]) {
			words = append(words, string(runes[start:i]))
			start = i
		}
	}
	words = append(words, string(runes[start:]))

	for i, w := range words {
		words[i] = strings.ToLower(w)
	}

	return strings.Join(words, "-")
}
