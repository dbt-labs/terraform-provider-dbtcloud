# `dbtp` CLI Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a CLI tool (`dbtp`) that reuses `pkg/dbt_cloud/` to provide imperative CRUD commands and declarative YAML-based config management for the dbt platform, with commands auto-generated from annotated Go source.

**Architecture:** The CLI lives at `cmd/dbtp/main.go`, uses `urfave/cli/v3` for command routing, Charmbracelet libraries for UX, and delegates all API calls to the existing `pkg/dbt_cloud.Client`. A code generator in `internal/codegen/` reads `@cli:` annotations from `pkg/dbt_cloud/` source and emits command files into `pkg/cli/generated/`.

**Tech Stack:** Go 1.24, `urfave/cli/v3`, `charmbracelet/lipgloss`, `charmbracelet/huh`, `charmbracelet/table`, `gopkg.in/yaml.v3`

**Module:** `github.com/dbt-labs/terraform-provider-dbtcloud`

---

## Phase 1: Foundation

### Task 1: Add Dependencies

**Files:**
- Modify: `go.mod`

**Step 1: Add CLI and UI dependencies**

Run:
```bash
cd /Users/bper/dev/terraform-provider-dbtcloud
go get github.com/urfave/cli/v3@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/huh@latest
go get github.com/charmbracelet/table@latest
go get gopkg.in/yaml.v3@latest
go mod tidy
```

**Step 2: Verify dependencies resolve**

Run: `go build ./...`
Expected: No errors

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "feat(dbtp): add urfave/cli and charmbracelet dependencies"
```

---

### Task 2: Create CLI Entrypoint and Auth

**Files:**
- Create: `cmd/dbtp/main.go`
- Create: `pkg/cli/auth.go`

**Step 1: Create the auth helper**

Create `pkg/cli/auth.go`:

```go
package cli

import (
	"fmt"
	"os"
	"strconv"

	dbtcloud "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

type AuthConfig struct {
	Token     string
	AccountID int64
	HostURL   string
}

func LoadAuthConfig() (*AuthConfig, error) {
	token := os.Getenv("DBT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("DBT_TOKEN environment variable is required")
	}

	accountIDStr := os.Getenv("DBT_ACCOUNT_ID")
	if accountIDStr == "" {
		return nil, fmt.Errorf("DBT_ACCOUNT_ID environment variable is required")
	}

	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("DBT_ACCOUNT_ID must be a valid integer: %w", err)
	}

	hostURL := os.Getenv("DBT_HOST_URL")
	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	return &AuthConfig{
		Token:     token,
		AccountID: accountID,
		HostURL:   hostURL,
	}, nil
}

func NewClientFromAuth(cfg *AuthConfig) (*dbtcloud.Client, error) {
	return dbtcloud.NewClient(
		&cfg.AccountID,
		&cfg.Token,
		&cfg.HostURL,
		nil, // maxRetries
		nil, // retryIntervalSeconds
		nil, // retriableStatusCodes
		false, // skipCredentialsValidation
		nil, // timeoutSeconds
	)
}
```

**Step 2: Create the CLI entrypoint**

Create `cmd/dbtp/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"os"

	dbtCli "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/cli"
	"github.com/urfave/cli/v3"
)

var version = "dev"

func main() {
	app := &cli.Command{
		Name:    "dbtp",
		Usage:   "CLI for the dbt platform",
		Version: version,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "Output format: table, json, yaml",
				Value:   "table",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "auth",
				Usage: "Authentication commands",
				Commands: []*cli.Command{
					{
						Name:  "status",
						Usage: "Check authentication status",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							cfg, err := dbtCli.LoadAuthConfig()
							if err != nil {
								return fmt.Errorf("not authenticated: %w", err)
							}
							client, err := dbtCli.NewClientFromAuth(cfg)
							if err != nil {
								return fmt.Errorf("authentication failed: %w", err)
							}
							_ = client
							fmt.Printf("Authenticated to account %d at %s\n", cfg.AccountID, cfg.HostURL)
							return nil
						},
					},
				},
			},
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 3: Verify it builds and runs**

Run:
```bash
go build -o dbtp ./cmd/dbtp/
./dbtp --help
./dbtp --version
```
Expected: Help text showing `dbtp` with `auth` subcommand and `--format` flag.

**Step 4: Commit**

```bash
git add cmd/dbtp/main.go pkg/cli/auth.go
git commit -m "feat(dbtp): add CLI entrypoint with auth support"
```

---

### Task 3: Output Formatting Module

**Files:**
- Create: `pkg/cli/output.go`

**Step 1: Create the output formatter**

Create `pkg/cli/output.go`:

```go
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/table"
	"gopkg.in/yaml.v3"
)

var (
	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	ErrorStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	SuccessStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	MutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func noColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}

// PrintTable renders a list of items as a styled table.
// headers are the column names, rows is a slice of string slices.
func PrintTable(headers []string, rows [][]string) {
	t := table.New().
		Headers(headers...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return HeaderStyle
			}
			return lipgloss.NewStyle()
		})

	for _, row := range rows {
		t.Row(row...)
	}

	fmt.Println(t.Render())
}

// PrintKeyValue renders a single item as key-value pairs.
func PrintKeyValue(pairs []KeyValue) {
	maxKeyLen := 0
	for _, kv := range pairs {
		if len(kv.Key) > maxKeyLen {
			maxKeyLen = len(kv.Key)
		}
	}

	for _, kv := range pairs {
		key := HeaderStyle.Render(kv.Key + ":" + strings.Repeat(" ", maxKeyLen-len(kv.Key)))
		fmt.Printf("%s %s\n", key, kv.Value)
	}
}

type KeyValue struct {
	Key   string
	Value string
}

// PrintJSON outputs data as formatted JSON.
func PrintJSON(data any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintYAML outputs data as YAML.
func PrintYAML(data any) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	return enc.Encode(data)
}

// FormatOutput routes to the correct output method based on format flag.
func FormatOutput(format string, data any, tableFunc func()) error {
	switch format {
	case "json":
		return PrintJSON(data)
	case "yaml":
		return PrintYAML(data)
	default:
		tableFunc()
		return nil
	}
}
```

**Step 2: Verify it compiles**

Run: `go build ./pkg/cli/`
Expected: No errors

**Step 3: Commit**

```bash
git add pkg/cli/output.go
git commit -m "feat(dbtp): add output formatting with lipgloss and table"
```

---

### Task 4: Project Commands (Handwritten)

This is the first resource implementation. It sets the pattern all generated commands will follow.

**Files:**
- Create: `pkg/cli/commands/project.go`
- Modify: `cmd/dbtp/main.go` (register project commands)

**Step 1: Create project commands**

Create `pkg/cli/commands/project.go`:

```go
package commands

import (
	"context"
	"fmt"
	"strconv"

	dbtCli "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/cli"
	dbtcloud "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/urfave/cli/v3"
)

func getClient() (*dbtcloud.Client, error) {
	cfg, err := dbtCli.LoadAuthConfig()
	if err != nil {
		return nil, err
	}
	return dbtCli.NewClientFromAuth(cfg)
}

func getFormat(cmd *cli.Command) string {
	return cmd.Root().String("format")
}

func ProjectCommands() *cli.Command {
	return &cli.Command{
		Name:  "project",
		Usage: "Manage dbt projects",
		Commands: []*cli.Command{
			projectGetCmd(),
			projectListCmd(),
			projectCreateCmd(),
			projectUpdateCmd(),
		},
	}
}

func projectGetCmd() *cli.Command {
	return &cli.Command{
		Name:  "get",
		Usage: "Get a project by ID",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "id",
				Usage:    "Project ID",
				Required: true,
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			project, err := client.GetProject(strconv.FormatInt(cmd.Int("id"), 10))
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}

			return dbtCli.FormatOutput(getFormat(cmd), project, func() {
				dbtCli.PrintKeyValue(projectToKeyValue(project))
			})
		},
	}
}

func projectListCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List projects",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "name",
				Usage: "Filter by project name",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			name := cmd.String("name")
			if name != "" {
				project, err := client.GetProjectByName(name)
				if err != nil {
					return fmt.Errorf("failed to find project: %w", err)
				}
				return dbtCli.FormatOutput(getFormat(cmd), []dbtcloud.Project{*project}, func() {
					dbtCli.PrintTable(
						[]string{"ID", "Name", "Description", "State"},
						[][]string{projectToRow(project)},
					)
				})
			}

			// No filter - get all projects via pagination
			data := client.GetData(client.AccountURL + "/projects/")
			var projects []dbtcloud.Project
			for _, raw := range data {
				if m, ok := raw.(map[string]any); ok {
					id := int(m["id"].(float64))
					name := fmt.Sprintf("%v", m["name"])
					desc := fmt.Sprintf("%v", m["description"])
					state := int(m["state"].(float64))
					projects = append(projects, dbtcloud.Project{
						ID: &id, Name: name, Description: desc, State: state,
					})
				}
			}

			return dbtCli.FormatOutput(getFormat(cmd), projects, func() {
				rows := make([][]string, 0, len(projects))
				for i := range projects {
					rows = append(rows, projectToRow(&projects[i]))
				}
				dbtCli.PrintTable([]string{"ID", "Name", "Description", "State"}, rows)
			})
		},
	}
}

func projectCreateCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Project name",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "Project description",
			},
			&cli.StringFlag{
				Name:  "dbt-project-subdirectory",
				Usage: "dbt project subdirectory",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			project, err := client.CreateProject(
				cmd.String("name"),
				cmd.String("description"),
				cmd.String("dbt-project-subdirectory"),
				1, // dbt project type
			)
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			format := getFormat(cmd)
			if format == "table" {
				fmt.Println(dbtCli.SuccessStyle.Render(fmt.Sprintf("Created project %q (ID: %d)", project.Name, *project.ID)))
			}
			return dbtCli.FormatOutput(format, project, func() {
				dbtCli.PrintKeyValue(projectToKeyValue(project))
			})
		},
	}
}

func projectUpdateCmd() *cli.Command {
	return &cli.Command{
		Name:  "update",
		Usage: "Update a project",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "id",
				Usage:    "Project ID",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "name",
				Usage: "New project name",
			},
			&cli.StringFlag{
				Name:  "description",
				Usage: "New project description",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			client, err := getClient()
			if err != nil {
				return err
			}

			id := strconv.FormatInt(cmd.Int("id"), 10)
			project, err := client.GetProject(id)
			if err != nil {
				return fmt.Errorf("failed to get project: %w", err)
			}

			if name := cmd.String("name"); name != "" {
				project.Name = name
			}
			if desc := cmd.String("description"); desc != "" {
				project.Description = desc
			}

			updated, err := client.UpdateProject(id, *project)
			if err != nil {
				return fmt.Errorf("failed to update project: %w", err)
			}

			format := getFormat(cmd)
			if format == "table" {
				fmt.Println(dbtCli.SuccessStyle.Render(fmt.Sprintf("Updated project %q (ID: %d)", updated.Name, *updated.ID)))
			}
			return dbtCli.FormatOutput(format, updated, func() {
				dbtCli.PrintKeyValue(projectToKeyValue(updated))
			})
		},
	}
}

func projectToRow(p *dbtcloud.Project) []string {
	state := "active"
	if p.State == dbtcloud.STATE_DELETED {
		state = "deleted"
	}
	return []string{
		strconv.Itoa(*p.ID),
		p.Name,
		p.Description,
		state,
	}
}

func projectToKeyValue(p *dbtcloud.Project) []dbtCli.KeyValue {
	state := "active"
	if p.State == dbtcloud.STATE_DELETED {
		state = "deleted"
	}
	pairs := []dbtCli.KeyValue{
		{Key: "ID", Value: strconv.Itoa(*p.ID)},
		{Key: "Name", Value: p.Name},
		{Key: "Description", Value: p.Description},
		{Key: "State", Value: state},
		{Key: "Account ID", Value: strconv.FormatInt(p.AccountID, 10)},
	}
	if p.DbtProjectSubdirectory != nil {
		pairs = append(pairs, dbtCli.KeyValue{Key: "Subdirectory", Value: *p.DbtProjectSubdirectory})
	}
	return pairs
}
```

**Step 2: Register project commands in main.go**

Update `cmd/dbtp/main.go` to import `pkg/cli/commands` and add `commands.ProjectCommands()` to the app's `Commands` slice.

**Step 3: Build and test manually**

Run:
```bash
go build -o dbtp ./cmd/dbtp/
./dbtp project --help
./dbtp project list --help
./dbtp project get --help
./dbtp project create --help
```
Expected: Help text for each subcommand with correct flags.

**Step 4: Commit**

```bash
git add pkg/cli/commands/project.go cmd/dbtp/main.go
git commit -m "feat(dbtp): add project CRUD commands"
```

---

### Task 5: Environment Commands (Handwritten)

**Files:**
- Create: `pkg/cli/commands/environment.go`
- Modify: `cmd/dbtp/main.go` (register environment commands)

**Step 1: Create environment commands**

Create `pkg/cli/commands/environment.go` following the same pattern as project commands:

- `dbtp environment get --project-id <N> --id <N>`
- `dbtp environment list --project-id <N>`
- `dbtp environment create --project-id <N> --name <S> --dbt-version <S> --type <S> [--custom-branch <S>] [--deployment-type <S>]`
- `dbtp environment update --project-id <N> --id <N> [--name <S>] [--dbt-version <S>]`
- `dbtp environment delete --project-id <N> --id <N>`

Uses `client.GetEnvironment()`, `client.GetAllEnvironments()`, `client.CreateEnvironment()`, `client.UpdateEnvironment()`, `client.DeleteEnvironment()`.

For list, use `client.GetAllEnvironments(projectID)` which returns `[]Environment`.

Helper functions: `environmentToRow()` and `environmentToKeyValue()` mapping Environment fields to display.

Table columns: `ID`, `Name`, `Type`, `dbt Version`, `State`.

**Step 2: Register in main.go**

Add `commands.EnvironmentCommands()` to the app's `Commands` slice.

**Step 3: Build and verify**

Run:
```bash
go build -o dbtp ./cmd/dbtp/
./dbtp environment --help
```

**Step 4: Commit**

```bash
git add pkg/cli/commands/environment.go cmd/dbtp/main.go
git commit -m "feat(dbtp): add environment CRUD commands"
```

---

### Task 6: Job Commands (Handwritten)

**Files:**
- Create: `pkg/cli/commands/job.go`
- Modify: `cmd/dbtp/main.go` (register job commands)

**Step 1: Create job commands**

Create `pkg/cli/commands/job.go`:

- `dbtp job get --id <N>`
- `dbtp job list [--project-id <N>] [--environment-id <N>]`
- `dbtp job create --name <S> --project-id <N> --environment-id <N> --execute-steps <S>... [many optional flags]`
- `dbtp job update --id <N> [--name <S>] [--description <S>] ...`

Uses `client.GetJob()`, `client.GetAllJobs()`, `client.CreateJob()`, `client.UpdateJob()`.

For `create`, expose the most common flags: `--name`, `--project-id`, `--environment-id`, `--execute-steps` (string slice), `--description`, `--schedule-cron`, `--num-threads`, `--target-name`, `--generate-docs`, `--run-generate-sources`. The remaining parameters get sensible defaults (e.g., `triggers` defaults to `schedule: true`, `timeoutSeconds` defaults to `0`).

Table columns: `ID`, `Name`, `Project ID`, `Environment ID`, `State`.

**Step 2: Register in main.go**

Add `commands.JobCommands()` to the app's `Commands` slice.

**Step 3: Build and verify**

Run:
```bash
go build -o dbtp ./cmd/dbtp/
./dbtp job --help
```

**Step 4: Commit**

```bash
git add pkg/cli/commands/job.go cmd/dbtp/main.go
git commit -m "feat(dbtp): add job CRUD commands"
```

---

### Task 7: Add Makefile Targets

**Files:**
- Modify: `Makefile`

**Step 1: Add CLI build targets**

Add to `Makefile`:

```makefile
build-cli:
	go build -ldflags "-w -s" -o dbtp ./cmd/dbtp/

install-cli: build-cli
	mv ./dbtp $(HOME)/.local/bin/dbtp
```

**Step 2: Verify**

Run: `make build-cli`
Expected: `dbtp` binary produced.

**Step 3: Commit**

```bash
git add Makefile
git commit -m "feat(dbtp): add Makefile targets for CLI build"
```

---

## Phase 2: Code Generator

### Task 8: Annotation Parser

**Files:**
- Create: `internal/codegen/parser.go`
- Create: `internal/codegen/parser_test.go`

**Step 1: Write tests for the annotation parser**

Create `internal/codegen/parser_test.go` that tests:

- Parsing `// @cli:resource Project` from a struct comment → extracts resource name `"Project"` and subcommand `"project"`
- Parsing `// @cli:description "Manage dbt projects"` → extracts description
- Parsing `// @cli:hidden` on a struct field → marks it hidden
- Parsing `// @cli:operation create` on a method → maps method to `create` verb
- Parsing `// @cli:required name` → extracts required fields
- Parsing `// @cli:optional description,subdirectory` → extracts optional fields
- Struct field parsing → extracts field name, Go type, JSON tag, whether pointer

Test data: create a `testdata/project.go` file with annotated code.

Run: `go test ./internal/codegen/ -v`
Expected: All fail (parser not yet written)

**Step 2: Implement the parser**

Create `internal/codegen/parser.go`:

- Uses `go/ast`, `go/parser`, `go/token` to parse Go source files
- `ParseFile(path string) (*FileInfo, error)` → returns parsed resources and operations
- `FileInfo` contains `[]Resource` and `[]Operation`
- `Resource` has `Name`, `Description`, `SubCommand`, `Fields []Field`
- `Field` has `Name`, `GoType`, `JSONTag`, `IsPointer`, `IsHidden`
- `Operation` has `Verb`, `MethodName`, `RequiredFields`, `OptionalFields`, `ReceiverType`

Annotation parsing: scan comment groups for lines matching `@cli:<directive>` regex, extract directive and value.

**Step 3: Run tests**

Run: `go test ./internal/codegen/ -v`
Expected: All pass

**Step 4: Commit**

```bash
git add internal/codegen/parser.go internal/codegen/parser_test.go internal/codegen/testdata/
git commit -m "feat(dbtp): add Go AST annotation parser for codegen"
```

---

### Task 9: Code Generator Templates and Runner

**Files:**
- Create: `internal/codegen/generator.go`
- Create: `internal/codegen/generator_test.go`
- Create: `internal/codegen/templates/resource.go.tmpl`
- Create: `internal/codegen/main.go`

**Step 1: Write the Go template**

Create `internal/codegen/templates/resource.go.tmpl` that generates a file matching the pattern from Task 4 (project.go). The template should produce:

- Package declaration (`package commands`)
- Import block
- `<Resource>Commands() *cli.Command` function returning command group
- One subcommand function per operation (`get`, `list`, `create`, `update`, `delete`)
- `<resource>ToRow()` and `<resource>ToKeyValue()` helpers
- Flag definitions derived from struct fields

**Step 2: Write generator tests**

Test that given a parsed `Resource` and `[]Operation`, the generator produces valid Go code that compiles.

**Step 3: Implement the generator**

Create `internal/codegen/generator.go`:
- `Generate(resources []Resource, operations []Operation, outputDir string) error`
- Loads template, executes per resource, writes to `<outputDir>/<resource>.go`
- Runs `goimports` on output for clean imports

**Step 4: Create the codegen CLI entrypoint**

Create `internal/codegen/main.go` (or `cmd/codegen/main.go`):
- Reads all `.go` files in `pkg/dbt_cloud/`
- Parses each for annotations
- Runs generator, outputs to `pkg/cli/generated/`

**Step 5: Run tests**

Run: `go test ./internal/codegen/ -v`
Expected: All pass

**Step 6: Commit**

```bash
git add internal/codegen/
git commit -m "feat(dbtp): add code generator with templates"
```

---

### Task 10: Annotate Resources and Generate

**Files:**
- Modify: `pkg/dbt_cloud/project.go` (add annotations)
- Modify: `pkg/dbt_cloud/environment.go` (add annotations)
- Modify: `pkg/dbt_cloud/job.go` (add annotations)

**Step 1: Add annotations to project.go**

Add `// @cli:resource`, `// @cli:description`, `// @cli:hidden`, `// @cli:operation`, `// @cli:required`, `// @cli:optional` comments to the `Project` struct and its CRUD methods in `pkg/dbt_cloud/project.go`.

**Step 2: Add annotations to environment.go and job.go**

Same pattern.

**Step 3: Run the generator**

Run:
```bash
go run ./internal/codegen/ -input=pkg/dbt_cloud/ -output=pkg/cli/generated/
```
Expected: Files generated in `pkg/cli/generated/`

**Step 4: Verify generated code compiles**

Run: `go build ./pkg/cli/generated/`

**Step 5: Replace handwritten commands with generated ones**

Update `cmd/dbtp/main.go` to import from `pkg/cli/generated/` instead of `pkg/cli/commands/` for the 3 annotated resources. Keep `pkg/cli/commands/` for handwritten commands that don't map to simple CRUD.

**Step 6: Build and verify behavior matches**

Run:
```bash
go build -o dbtp ./cmd/dbtp/
./dbtp project --help
./dbtp environment --help
./dbtp job --help
```

**Step 7: Add Makefile target**

Add to `Makefile`:
```makefile
generate-cli:
	go run ./internal/codegen/ -input=pkg/dbt_cloud/ -output=pkg/cli/generated/
```

**Step 8: Commit**

```bash
git add pkg/dbt_cloud/project.go pkg/dbt_cloud/environment.go pkg/dbt_cloud/job.go \
       pkg/cli/generated/ cmd/dbtp/main.go Makefile
git commit -m "feat(dbtp): annotate resources and generate CLI commands"
```

---

## Phase 3: Annotate Remaining Resources

### Task 11: Annotate All Resources in pkg/dbt_cloud/

Work through remaining resources one batch at a time, running the generator and verifying after each batch:

**Batch 1 - Credentials:**
`snowflake_credential.go`, `bigquery_credential.go`, `databricks_credential.go`, `postgres_credential.go`

**Batch 2 - Access Control:**
`group.go`, `service_token.go`, `license_map.go`, `ip_restrictions_rule.go`

**Batch 3 - Config & Integration:**
`webhook.go`, `notification.go`, `repository.go`, `global_connection.go`

**Batch 4 - Remaining:**
All other resource files in `pkg/dbt_cloud/`

For each batch:
1. Add annotations
2. Run `make generate-cli`
3. Run `go build ./cmd/dbtp/`
4. Verify with `./dbtp <resource> --help`
5. Commit

---

## Phase 4: Declarative Mode

### Task 12: YAML Schema and Parser

**Files:**
- Create: `pkg/cli/declarative/types.go`
- Create: `pkg/cli/declarative/parser.go`
- Create: `pkg/cli/declarative/parser_test.go`

Define the YAML config types:

```go
type Config struct {
	AccountID int64             `yaml:"account_id"`
	Projects  []ProjectConfig   `yaml:"projects"`
}

type ProjectConfig struct {
	ID          *int              `yaml:"id,omitempty"`
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Delete      bool              `yaml:"_delete,omitempty"`
	Environments []EnvironmentConfig `yaml:"environments,omitempty"`
	Jobs        []JobConfig       `yaml:"jobs,omitempty"`
}
// ... similar for EnvironmentConfig, JobConfig
```

Test: parse a sample YAML file, verify all fields populated correctly.

---

### Task 13: Diff Engine

**Files:**
- Create: `pkg/cli/declarative/diff.go`
- Create: `pkg/cli/declarative/diff_test.go`

Implement:
- `ComputeDiff(desired Config, current Config) []Change`
- `Change` struct with `Type` (create/update/delete), `ResourceType`, `ResourceName`, `Fields []FieldChange`
- `FieldChange` with `Field`, `OldValue`, `NewValue`

Match resources by ID first, then by name within parent scope.
Items with `_delete: true` produce delete changes.

---

### Task 14: Apply Command

**Files:**
- Create: `pkg/cli/declarative/apply.go`
- Modify: `cmd/dbtp/main.go` (register apply/export commands)

Implement:
- `dbtp apply --file <path> [--dry-run] [--auto-approve]`
- `dbtp export --project-id <N>`

`--dry-run`: compute and display diff, exit.
Default: compute diff, display, prompt for confirmation.
`--auto-approve`: compute diff, apply without prompting.

Colorized diff output using lipgloss: green for creates, yellow for updates, red for deletes.

---

## Phase 5: Polish

### Task 15: Interactive Forms

Add `charmbracelet/huh` integration to `create` commands: when required flags are missing, launch an interactive form instead of erroring.

### Task 16: Shell Completions

Add `dbtp completion bash|zsh|fish` command using urfave/cli's built-in completion support.

### Task 17: CI Check

Add a GitHub Actions workflow that runs `make generate-cli` and checks `git diff --exit-code` to ensure generated code is up to date.
