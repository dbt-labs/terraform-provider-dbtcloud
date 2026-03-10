package codegen

// ResourceDef is the top-level YAML schema for a CLI resource definition.
type ResourceDef struct {
	Name           string            `yaml:"name"`
	FuncPrefix     string            `yaml:"func_prefix"`    // prefix for generated function names; defaults to Name
	Description    string            `yaml:"description"`
	GoType         string            `yaml:"go_type"`
	ListType       string            `yaml:"list_type"` // defaults to GoType if empty
	Columns        []Field           `yaml:"columns"`
	Details        []Field           `yaml:"details"`
	ExtraColumns   []Field           `yaml:"extra_columns"`   // additional columns appended during merge
	LabelOverrides map[string]string `yaml:"label_overrides"` // json_tag → custom label for auto-derived fields
	Get            *Operation        `yaml:"get"`
	List           *ListOp           `yaml:"list"`
	Create         *Operation        `yaml:"create"`
	Update         *UpdateOp         `yaml:"update"`
	Delete         *Operation        `yaml:"delete"`
}

// Field describes a struct field for display.
type Field struct {
	GoField string `yaml:"field"`    // Go struct field name (e.g., "DbtProjectSubdirectory")
	JSONTag string `yaml:"json_tag"` // json tag name (e.g., "repository_id"); resolved to GoField per struct
	Label   string `yaml:"label"`    // Display label (e.g., "Subdirectory")
	Format  string `yaml:"format"`   // Special formatting: "state", "bool", "join", "json", or empty for auto
}

// Flag describes a CLI flag.
type Flag struct {
	Name     string `yaml:"name"`     // CLI flag name (e.g., "project-id")
	GoType   string `yaml:"type"`     // Go type: "string", "int", "bool", "string_slice"
	Required bool   `yaml:"required"` // Whether the flag is required
	Usage    string `yaml:"usage"`    // Help text
	Default  string `yaml:"default"`  // Default value (as string)
}

// Operation describes a simple CRUD operation (get, create, delete).
type Operation struct {
	Flags []Flag `yaml:"flags"`
	// Call is the literal Go expression for the API call.
	// Available variables: client, cmd, plus any flag values read above.
	Call string `yaml:"call"`
	// ResultVar is the variable name for the result (e.g., "project").
	ResultVar string `yaml:"result_var"`
}

// ListOp describes a list operation which can use either a method call or raw pagination.
type ListOp struct {
	Flags []Flag `yaml:"flags"`
	// Mode is either "method" (calls a Go method) or "paginate" (uses GetRawData).
	Mode string `yaml:"mode"`
	// Call is the Go expression for method mode (e.g., "client.GetAllEnvironments(projectID)").
	Call string `yaml:"call"`
	// URL is the URL pattern for paginate mode (e.g., "/v3/accounts/%d/projects/").
	URL string `yaml:"url"`
	// Filter describes an optional filter (e.g., filter by name for projects).
	Filter *ListFilter `yaml:"filter"`
	// Validation is optional Go code run before the API call (e.g., mutual exclusion checks).
	Validation string `yaml:"validation"`
}

// ListFilter describes a filter on a list operation.
type ListFilter struct {
	Flag string `yaml:"flag"` // Which flag triggers the filter
	Call string `yaml:"call"` // Go expression returning a single result
}

// UpdateOp describes an update operation using the fetch-then-update pattern.
type UpdateOp struct {
	Flags []Flag `yaml:"flags"`
	// IDExpr is the Go expression to compute the ID (e.g., `strconv.Itoa(int(cmd.Int("id")))`).
	IDExpr string `yaml:"id_expr"`
	// GetCall is the Go expression to fetch the existing resource.
	GetCall string `yaml:"get_call"`
	// UpdateCall is the Go expression to perform the update.
	UpdateCall string `yaml:"update_call"`
	// Editable lists the fields that can be updated via flags.
	Editable []EditableField `yaml:"editable"`
}

// EditableField maps a CLI flag to a Go struct field for updates.
type EditableField struct {
	Flag    string `yaml:"flag"`    // CLI flag name
	GoField string `yaml:"field"`   // Go struct field name
	GoType  string `yaml:"type"`    // "string", "int", "bool"
	Usage   string `yaml:"usage"`   // Help text
}
