# `dbtp` CLI Design Document

**Date:** 2026-02-23
**Status:** Draft

## Overview

`dbtp` is a CLI tool for performing CRUD operations on the dbt platform without Terraform. It lives in the same repo as the Terraform provider and reuses the existing `pkg/dbt_cloud/` API client. A code generator reads annotated Go source and auto-generates CLI commands, so adding a new resource to the provider gives the CLI that resource for (near) free.

## Goals

- **Imperative commands**: Ad-hoc `dbtp <resource> <verb>` operations for quick tasks
- **Declarative config**: Define desired state in YAML and `apply` it
- **Auto-generated**: CLI commands are generated from annotated `pkg/dbt_cloud/` source
- **Polished UX**: Interactive forms when flags are missing, styled output, colorized diffs

## Non-Goals

- Replacing Terraform for full infrastructure management
- Managing all dbt platform object types declaratively
- Backward compatibility with any existing CLI tool

## Tech Stack

- **CLI framework**: [urfave/cli](https://github.com/urfave/cli)
- **UI**: [Charmbracelet](https://github.com/charmbracelet) libraries (`lipgloss` for styling, `huh` for interactive forms, `table` for tabular output)
- **API client**: Existing `pkg/dbt_cloud/` package (no Terraform dependencies)
- **Code generation**: Go AST parsing + Go templates

## Repository Structure

```
terraform-provider-dbtcloud/
├── cmd/
│   ├── terraform-provider-dbtcloud/   # existing
│   └── dbtp/                          # new CLI binary
│       └── main.go
├── pkg/
│   ├── dbt_cloud/                     # existing API client (annotated)
│   └── cli/                           # new - generated + handwritten CLI code
│       ├── generated/                 # auto-generated commands per resource
│       ├── declarative/               # YAML config parsing + sync/apply logic
│       └── ui/                        # charmbracelet-based formatting/interaction
├── internal/
│   └── codegen/                       # the generator tool
│       ├── main.go                    # reads pkg/dbt_cloud/, emits pkg/cli/generated/
│       ├── parser.go                  # Go AST parsing + annotation extraction
│       └── templates/                 # Go templates for command scaffolding
└── Makefile                           # `make generate-cli` target
```

Generated files in `pkg/cli/generated/` are committed to the repo so the CLI builds without running the generator. A CI check ensures they stay up to date.

## Annotations

Lightweight Go comments in `pkg/dbt_cloud/` guide the generator. Annotations are placed on structs and methods.

### Struct annotations

```go
// @cli:resource Project
// @cli:description "Manage dbt projects"
type Project struct {
    ID                     *int    `json:"id,omitempty"`
    Name                   string  `json:"name"`
    Description            string  `json:"description"`
    // @cli:hidden
    DbtProjectType         *int    `json:"dbt_project_type,omitempty"`
}
```

- `@cli:resource <Name>`: Marks a struct as a CLI resource. The name becomes the subcommand.
- `@cli:description "<text>"`: Help text for the subcommand group.
- `@cli:hidden`: Excludes a field from CLI flags.

### Method annotations

```go
// @cli:operation create
// @cli:required name
// @cli:optional description,dbt_project_subdirectory
func (c *Client) CreateProject(name string, description string, ...) (*Project, error)

// @cli:operation get
func (c *Client) GetProject(projectID string) (*Project, error)

// @cli:operation list
func (c *Client) GetProjectByName(name string) (*Project, error)

// @cli:operation update
func (c *Client) UpdateProject(projectID string, project Project) (*Project, error)

// @cli:operation delete
func (c *Client) DeleteProject(projectID string) (string, error)
```

- `@cli:operation <verb>`: Maps a method to a CLI verb (`create`, `get`, `list`, `update`, `delete`).
- `@cli:required <fields>`: Comma-separated list of required flags.
- `@cli:optional <fields>`: Comma-separated list of optional flags.

### Flag derivation

- Flag names are derived from `json` struct tags, converted from `snake_case` to `--kebab-case`
- Go types map to flag types: `string` → string flag, `int` → int flag, `bool` → bool flag, `*T` → optional
- `@cli:hidden` fields are excluded entirely

## CLI UX

### Authentication

```bash
# Environment variables
export DBT_TOKEN=your-token
export DBT_ACCOUNT_ID=12345

# Or interactive setup
dbtp auth login

# Or config file (~/.dbtp/config.yml)
```

### Imperative commands

Consistent `dbtp <resource> <verb>` pattern:

```bash
dbtp project list
dbtp project get --id 123
dbtp project create --name "Analytics" --description "Main project"
dbtp project update --id 123 --name "Analytics v2"
dbtp project delete --id 123

dbtp job list --project-id 123
dbtp job get --id 456
dbtp job create --name "Daily Run" --project-id 123 --environment-id 789

# Complex operations (handwritten, not generated)
dbtp job run --id 456 --wait
```

### Output formatting

- **Default**: Styled tables (list) or key-value (get) via `charmbracelet/table` and `lipgloss`
- **`--format json`**: JSON output for scripting
- **`--format yaml`**: YAML output for consistency with declarative mode
- Respects `NO_COLOR` environment variable

### Interactive fallback

When required flags are missing, the CLI launches an interactive form via `charmbracelet/huh` instead of erroring:

```bash
$ dbtp project create
# → Interactive form prompts for name, description, etc.
```

## Declarative Mode

### YAML structure

```yaml
account_id: 12345

projects:
  - name: "Analytics"
    description: "Main analytics project"

    environments:
      - name: "Production"
        dbt_version: "1.8.0"
        type: deployment
      - name: "Staging"
        dbt_version: "1.8.0"
        type: deployment

    jobs:
      - name: "Daily Run"
        environment: "Production"    # reference by name
        execute_steps:
          - "dbt build"
        schedule:
          cron: "0 8 * * *"
      - id: 789
        _delete: true
```

### Key behaviors

- **Matching**: Resources matched by `id` if present, or by `name` within their parent scope
- **References by name**: Related resources can be referenced by name (e.g., `environment: "Production"`); the CLI resolves to IDs
- **Export**: `dbtp export --project-id 123 > dbt-platform.yml` fetches current state as YAML
- **Idempotent**: Running `apply` twice with no changes produces no API calls

### Deletes

Resources are only deleted when explicitly marked with `_delete: true`. Absence from the YAML file means "don't touch", not "delete". This is a deliberate safety choice.

### Apply workflow

```bash
# Preview changes
dbtp apply --file dbt-platform.yml --dry-run

# Apply interactively (prompts before deletes)
dbtp apply --file dbt-platform.yml

# Apply non-interactively (for CI/CD)
dbtp apply --file dbt-platform.yml --auto-approve
```

### Diff output

`--dry-run` shows a colorized diff:

```
  + Create environment "Staging" in project "Analytics"
  ~ Update job "Daily Run" (id: 456)
    schedule.cron: "0 6 * * *" → "0 8 * * *"
  ✗ Delete job (id: 789)
    ⚠ This action is irreversible
```

Creates in green, updates in yellow, deletes in red.

## Implementation Phases

### Phase 1 - Foundation

- Set up `cmd/dbtp/main.go` with `urfave/cli`
- Auth handling (env vars, config file, `dbtp auth login`)
- Output formatting with `lipgloss`, `table`, `--format json/yaml`
- Interactive fallback with `huh`
- Handwrite 2-3 resource commands (project, job, environment) to nail the UX patterns

### Phase 2 - Code Generator

- Build `internal/codegen/` with Go AST parser and annotation extraction
- Create Go templates for `urfave/cli` command scaffolding
- Annotate the 2-3 resources from Phase 1
- Validate generator output matches handwritten code
- Replace handwritten commands with generated ones
- Add `make generate-cli` target

### Phase 3 - Annotate All Resources

- Work through remaining ~40 resources in `pkg/dbt_cloud/`, adding annotations
- Run generator, review output, fix edge cases in templates

### Phase 4 - Declarative Mode

- YAML parsing and schema definition
- Diff engine (current state vs desired state)
- `dbtp export` command
- `dbtp apply` with `--dry-run` and `--auto-approve`
- Colorized diff output
- `_delete: true` handling with confirmation prompts

### Phase 5 - Polish

- `dbtp auth login` interactive setup
- Config file support (`~/.dbtp/config.yml`)
- Shell completions (bash, zsh, fish)
- CI check to ensure generated code is up to date
- Documentation
