# CLI Code Generator

`cmd/codegen/` generates the Go command files that power `dbtp`. It derives
CLI commands from the **same Terraform schemas and Go client code** used by the
Terraform provider, so flags, types, descriptions, and validations stay in sync
without manual duplication.

```bash
go run ./cmd/codegen/   # regenerate all commands into pkg/cli/commands/
```

## Architecture overview

```
                        ┌──────────────────────────┐
                        │   TF Resource Schema      │  flag names, types,
                        │   (pkg/framework/objects/) │  required/optional,
                        └────────────┬─────────────┘  descriptions, sensitive
                                     │
                        ┌────────────▼─────────────┐
                        │   Go Client Methods       │  get/list/create/update/
                        │   (pkg/dbt_cloud/)        │  delete signatures,
                        └────────────┬─────────────┘  param types
                                     │
              ┌──────────────────────▼──────────────────────┐
              │              Derivation pipeline             │
              │  1. ExtractResourceAttrs (TF schema)        │
              │  2. ScanMethods (*Client methods)           │
              │  3. ParseStructTags (json tags → Go fields) │
              │  4. DeriveTFResourceDef (combine all)       │
              └──────────────────────┬─────────────────────┘
                                     │
              ┌──────────────────────▼──────────────────────┐
              │          YAML Override (optional)            │
              │       (cli-resources/*.yaml)                 │
              │                                             │
              │  Partial: replace specific operations only   │
              │  Full: override entire resource definition   │
              └──────────────────────┬─────────────────────┘
                                     │
              ┌──────────────────────▼──────────────────────┐
              │         Go Template (resource.go.tmpl)       │
              │              ↓                               │
              │    pkg/cli/commands/<name>_gen.go            │
              └─────────────────────────────────────────────┘
```

## Resource registry

Resources are defined in `cmd/codegen/main.go` as a flat list:

```go
var cliResources = []cliResource{
    {
        Name:       "project",
        GoType:     "Project",
        TFResource: project.ProjectResource(),
        TFPluralDS: project.ProjectsDataSource(),
    },
    // ...
}
```

Each entry needs:

| Field | Required | Description |
|-------|----------|-------------|
| `Name` | yes | CLI command name (e.g., `"project"`, `"snowflake"`) |
| `GoType` | yes | Go struct in `pkg/dbt_cloud/` (e.g., `"Project"`) |
| `TFResource` | yes | Terraform resource constructor for schema extraction |
| `TFPluralDS` | no | Terraform plural datasource (for list filter attrs). Nil if no list endpoint. |
| `FuncPrefix` | no | Prefix for generated function names when it differs from `Name`. See [FuncPrefix](#funcprefix). |
| `ListType` | no | Separate Go struct for list view (e.g., `"JobWithEnvironment"`). Defaults to `GoType`. |
| `DefaultDelete` | no | Fallback delete operation injected when none is auto-derived. See [DefaultDelete](#defaultdelete). |

**Adding a new resource** is usually a one-line addition to this list.

## Derivation pipeline

The pipeline runs for each resource in the registry:

### Step 1: Extract TF schema attributes

`ExtractResourceAttrs()` reads the Terraform resource schema and returns
attribute metadata:

- **Name** -- snake_case attribute name (e.g., `"dbt_project_subdirectory"`)
- **Type** -- `string`, `int64`, `bool`, `list_string`, `nested`, `block`
- **Required / Optional / Computed** -- used to determine which fields are
  user-settable vs. server-computed
- **Sensitive** -- marks fields for masked display (`*****`)
- **Description** -- reused as `--flag` help text
- **SubAttrs** -- for `SingleNestedAttribute`, contains child attributes (used
  for flattening nested structs in update operations)

### Step 2: Scan `*Client` methods

`ScanMethods()` parses all Go files in `pkg/dbt_cloud/` and extracts methods on
`*Client` that match CRUD patterns:

| Pattern | Kind | Example |
|---------|------|---------|
| `Get<Name>` | get | `GetProject(projectID string)` |
| `GetAll<Name>s` | list | `GetAllEnvironments(projectID string)` |
| `Create<Name>` | create | `CreateSnowflakeCredential(projectId int, ...)` |
| `Update<Name>` | update | `UpdateProject(projectID string, project Project)` |
| `Delete<Name>` | delete | `DeleteCredential(credentialId, projectId string)` |

Methods are grouped by the `<Name>` portion and matched to registry entries via
`GoType`.

### Step 3: Parse struct tags

`ParseStructTags()` and `ParseFieldTypes()` read all Go structs in
`pkg/dbt_cloud/` to build two maps:

- **json tag -> Go field name + type** (e.g., `"dbt_version"` -> `DbtVersion *string`)
- **struct.field -> Go type** (for template type lookups in display formatting)

These maps bridge TF attribute names (snake_case) to Go struct fields
(PascalCase).

### Step 4: Derive ResourceDef

`DeriveTFResourceDef()` combines all the above into a `ResourceDef`:

- **Display fields** (columns/details) -- TF attributes matched to Go struct
  fields, with automatic format detection (`state`, `bool`, `join`, `json`,
  `secret`)
- **Get operation** -- flags and API call from the `Get<Name>` method signature
- **List operation** -- flags from the `GetAll<Name>s` method, enriched with TF
  datasource filter descriptions
- **Create operation** -- flags from the `Create<Name>` method, enriched with
  TF schema descriptions and required/optional markers
- **Update operation** -- fetch-then-update pattern with editable fields derived
  from TF schema (Optional, non-Computed-only attributes)
- **Delete operation** -- flags and API call from the `Delete<Name>` method

### Step 5: Merge with YAML override

If a file `cli-resources/<name>.yaml` exists, it is merged with the auto-derived
definition. See [YAML overrides](#yaml-overrides).

### Step 6: Generate Go code

The merged `ResourceDef` is passed to the Go template
(`internal/codegen/resource.go.tmpl`) which produces the final
`pkg/cli/commands/<name>_gen.go` file, then runs `goimports` for formatting.

## Auto-derivation details

### Flag name mapping

Parameter names are converted to CLI flag names:

| Parameter | Resource | Flag |
|-----------|----------|------|
| `projectID` | `Project` | `--id` (resource's own ID) |
| `projectId` | `SnowflakeCredential` | `--project-id` |
| `credentialId` | `SnowflakeCredential` | `--id` (suffix match) |
| `isActive` | any | `--is-active` (camelToKebab) |

The `paramToFlagName` function applies two rules:
1. If the param name equals the resource name + `"id"`, rename to `--id`
2. If the param name suffix matches the resource name suffix + `"id"`, rename to
   `--id` (e.g., `credentialId` on `SnowflakeCredential`)

### String ID parameters

Many API methods accept IDs as `string` but the CLI exposes them as `--id int`.
When a parameter name ends in `id`/`ID` and has Go type `string`, the generated
code reads it as `cmd.Int("id")` and converts via `strconv.Itoa()`.

### Update editable fields

The update command uses a fetch-then-update pattern:

```
1. GET the existing resource
2. Apply flag overrides to the fetched struct
3. PUT the modified struct back
```

Editable fields are derived from TF schema attributes that are:
- **Optional** or **Required** (user-settable)
- **Not Computed-only** (not server-generated)

Supported Go types for editable fields:

| Go type | CLI flag type | Update behavior |
|---------|--------------|----------------|
| `string` | `StringFlag` | Set if non-empty |
| `*string` | `StringFlag` | Set pointer if `--flag` is provided |
| `int` | `IntFlag` | Set if non-zero |
| `*int` | `IntFlag` | Set pointer if `--flag` is provided |
| `bool` | `BoolFlag` | Set if `--flag` is explicitly passed |
| `[]string` | `StringSliceFlag` | Replace slice if `--flag` is provided |

### Nested struct flattening

`SingleNestedAttribute` types (e.g., `triggers`, `execution`) are flattened
into individual editable flags:

```
TF schema:                          CLI flag:
triggers.schedule        (bool) --> --triggers-schedule
triggers.on_merge        (bool) --> --triggers-on-merge
execution.timeout_seconds (int) --> --execution-timeout-seconds
```

The generated code uses dotted Go field paths: `existing.Triggers.Schedule = ...`

Pointer-to-struct types (e.g., `*JobCompletionTrigger`) are **not** flattened
because the parent might be `nil` at runtime, which would cause a panic.

### Sensitive fields

TF attributes with `Sensitive: true` (e.g., passwords, private keys) are
assigned `Format: "secret"` during derivation. The template renders these as
`"*****"` in both table and detail views. The actual value is never displayed
in get/list output.

### Display field formats

Fields are auto-formatted for display based on their Go type:

| Format | Trigger | Display |
|--------|---------|---------|
| (auto) | `string` | raw value |
| (auto) | `int`, `int64` | `strconv.Itoa` / `FormatInt` |
| `bool` | `bool` | `"true"` / `"false"` |
| `state` | field named `State` with int type | `"active"` / `"deleted"` |
| `join` | `[]string` | comma-separated |
| `json` | nested/block TF types | JSON-marshaled |
| `secret` | `Sensitive: true` in TF schema | `"*****"` |

Pointer types (`*int`, `*string`, `*bool`) get automatic nil-check handling in
the template.

## YAML overrides

YAML files in `cli-resources/` override or supplement auto-derived operations.
The file name (without extension) is matched to the resource name.

### Partial override

When the YAML does **not** set `go_type`, it merges with the auto-derived
definition. Only operations present in the YAML replace the derived ones:

```yaml
# cli-resources/project.yaml
name: project

# Override display labels
label_overrides:
  dbt_project_subdirectory: "Subdirectory"

# Add columns that exist in the Go struct but not in the TF schema
extra_columns:
  - {json_tag: connection_id, label: "Connection ID"}
  - {json_tag: repository_id, label: "Repository ID"}
```

### Full override

When the YAML sets `go_type`, the entire auto-derived definition is replaced.
This is useful for resources whose methods are too complex for auto-derivation.

### Operation fields

#### `get` / `create` / `delete`

```yaml
get:
  flags:
    - {name: id, type: int, required: true, usage: "Resource ID"}
  call: 'client.GetFoo(strconv.Itoa(int(cmd.Int("id"))))'
  result_var: foo
```

#### `list`

```yaml
list:
  mode: method    # or "paginate" for raw URL-based pagination
  call: 'client.GetAllFoos(int(cmd.Int("project-id")))'
  flags:
    - {name: project-id, type: int, usage: "Filter by project ID"}
  validation: |   # optional Go code run before the API call
    if int(cmd.Int("project-id")) == 0 {
      return fmt.Errorf("--project-id is required")
    }
```

#### `update`

Not commonly overridden since the fetch-then-update pattern is auto-derived.

#### `create` with setup code

The `call` field can contain multiple lines. Everything before the last
`client.` line becomes setup code:

```yaml
create:
  call: |
    triggers := map[string]any{"schedule": false}
    client.CreateJob(int(cmd.Int("project-id")), triggers)
  result_var: job
```

### Flag types

| Type | CLI type | Go getter |
|------|----------|-----------|
| `string` | `StringFlag` | `cmd.String("name")` |
| `int` | `IntFlag` | `cmd.Int("name")` |
| `bool` | `BoolFlag` | `cmd.Bool("name")` |
| `string_slice` | `StringSliceFlag` | `cmd.StringSlice("name")` |

### Merge behavior

| YAML field | Behavior |
|-----------|----------|
| `get`, `list`, `create`, `update`, `delete` | Replaces the auto-derived operation |
| `name`, `description` | Overrides the derived value |
| `list_type` | Overrides the list struct type |
| `label_overrides` | Renames display labels by json tag |
| `extra_columns` | Appends additional display columns |

## FuncPrefix

For resources that live as subcommands under a parent group (like credentials),
`FuncPrefix` separates Go function names from the CLI command name:

```go
{
    Name:       "snowflake",           // CLI: dbtp credential snowflake
    FuncPrefix: "credentialSnowflake", // Go:  CredentialSnowflakeCommands()
    GoType:     "SnowflakeCredential",
    TFResource: snowflake_credential.SnowflakeCredentialResource(),
}
```

- **CLI command name** uses `Name`: `dbtp credential snowflake get`
- **Go functions** use `FuncPrefix`: `CredentialSnowflakeCommands()`,
  `credentialSnowflakeToRow()`, etc.
- **Output filename** uses `FuncPrefix` (camelToSnake): `credential_snowflake_gen.go`
- When `FuncPrefix` is empty, `Name` is used for everything (backwards
  compatible).

The parent command group is assembled manually in `cmd/dbtp/main.go`:

```go
{
    Name:  "credential",
    Usage: "Manage warehouse credentials",
    Commands: []*cli.Command{
        commands.CredentialSnowflakeCommands(),
        commands.CredentialPostgresCommands(),
        // ...
    },
}
```

## DefaultDelete

Some resources share a delete endpoint. For example, all credential types use
`client.DeleteCredential(credentialId, projectId)`, but the method scanner
classifies this under resource name `"Credential"` -- not `"SnowflakeCredential"`.

`DefaultDelete` provides a fallback delete operation that is injected after
derivation if no delete was auto-derived or overridden by YAML:

```go
{
    Name:          "snowflake",
    GoType:        "SnowflakeCredential",
    DefaultDelete: credentialDelete(), // shared across all credential types
}
```

## File map

```
cmd/codegen/
  main.go                    Resource registry, pipeline orchestration

cmd/dbtp/
  main.go                    CLI entry point, command wiring

internal/codegen/
  schema.go                  ResourceDef and related types (the intermediate representation)
  generator.go               Template execution, file output, goimports
  resource.go.tmpl           Go template that produces *_gen.go files
  tfschema.go                TF schema extraction (attributes, descriptions, sensitive)
  tfderive.go                Derivation logic (TF + methods + struct tags → ResourceDef)
  merge.go                   YAML override merging
  methods.go                 *Client method scanning, flag/call expression builders
  structtags.go              Go struct json tag and field type parsing
  annotations.go             Label formatting, auto-format detection
  fieldinfo.go               Struct field type map (for template type lookups)

cli-resources/
  project.yaml               Partial override: extra columns, label overrides
  job.yaml                   Partial override: custom list/create operations

pkg/cli/
  auth.go                    Authentication (env vars → API client)
  output.go                  Table, key-value, JSON, YAML formatting

pkg/cli/commands/
  common.go                  Shared helpers (getClient, getFormat, stateLabel)
  project_gen.go             Generated -- do not edit
  environment_gen.go         Generated -- do not edit
  job_gen.go                 Generated -- do not edit
  credential_snowflake_gen.go  Generated -- do not edit
  credential_postgres_gen.go   Generated -- do not edit
  credential_redshift_gen.go   Generated -- do not edit
  credential_bigquery_gen.go   Generated -- do not edit
```

## Adding a new resource

### Simple case (all methods auto-derivable)

1. Add one entry to `cliResources` in `cmd/codegen/main.go`:

   ```go
   {
       Name:       "environment",
       GoType:     "Environment",
       TFResource: environment.EnvironmentResource(),
       TFPluralDS: environment.EnvironmentsDataSource(),
   },
   ```

2. Run `go run ./cmd/codegen/`

3. Wire the command in `cmd/dbtp/main.go`:

   ```go
   commands.EnvironmentCommands(),
   ```

### With YAML overrides

If a method can't be auto-derived (complex params, custom logic), add a YAML
file:

1. Add the registry entry as above
2. Create `cli-resources/<name>.yaml` with the operations that need overriding
3. Run codegen -- the YAML operations replace the auto-derived ones

### As a nested subcommand

For resources under a parent group (like credentials):

1. Add the registry entry with `FuncPrefix`:

   ```go
   {
       Name:          "mydb",
       FuncPrefix:    "credentialMydb",
       GoType:        "MydbCredential",
       TFResource:    mydb_credential.MydbCredentialResource(),
       DefaultDelete: credentialDelete(), // if shared delete
   },
   ```

2. Run codegen

3. Add to the parent group in `cmd/dbtp/main.go`:

   ```go
   commands.CredentialMydbCommands(),
   ```
