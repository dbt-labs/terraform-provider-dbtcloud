# dbtp -- CLI for the dbt platform

`dbtp` is a command-line tool for managing dbt Cloud resources. It is built from
the same Go client library and Terraform schema definitions used by the
[dbt Cloud Terraform provider](https://registry.terraform.io/providers/dbt-labs/dbtcloud/latest/docs),
so its flags, descriptions, and validations stay in sync automatically.

## Quick start

```bash
# Build
make build-cli          # produces ./dbtp

# Install to ~/.local/bin
make install-cli

# Or build directly
go run ./cmd/codegen/   # generate command files first
go build -o dbtp ./cmd/dbtp/
```

## Authentication

`dbtp` reads credentials from environment variables:

| Variable | Required | Description |
|----------|----------|-------------|
| `DBT_TOKEN` | yes | API token (service token or personal access token) |
| `DBT_ACCOUNT_ID` | yes | dbt Cloud account ID |
| `DBT_HOST_URL` | no | API base URL (default: `https://cloud.getdbt.com/api`) |

```bash
export DBT_TOKEN="your-token"
export DBT_ACCOUNT_ID="12345"
dbtp auth status
```

## Usage

```
dbtp <resource> <action> [flags]
```

### Output formats

Every command supports `--format` (`-f`):

```bash
dbtp project list                        # table (default)
dbtp project list -f json                # JSON
dbtp project get --id 1 -f yaml          # YAML
```

### Resources

#### project

```bash
dbtp project list
dbtp project get --id 123
dbtp project create --name "My Project"
dbtp project update --id 123 --name "Renamed"
```

#### environment

```bash
dbtp environment list --project-id 123
dbtp environment get --project-id 123 --id 456
```

#### job

```bash
dbtp job list --project-id 123
dbtp job get --id 789
dbtp job create --name "Daily run" \
  --project-id 123 \
  --environment-id 456 \
  --execute-steps "dbt run" \
  --execute-steps "dbt test"
dbtp job update --id 789 \
  --execute-steps "dbt build" \
  --triggers-schedule=true
dbtp job delete --id 789
```

#### credential (nested subcommands per adapter)

Credentials use nested subcommands because each warehouse adapter has a
different schema:

```bash
dbtp credential snowflake get --project-id 1 --id 5
dbtp credential snowflake create --project-id 1 \
  --auth-type password --database mydb --warehouse mywh \
  --schema public --user myuser --password secret --num-threads 4 --type ""
dbtp credential snowflake update --project-id 1 --id 5 --warehouse newwh
dbtp credential snowflake delete --project-id 1 --id 5

dbtp credential postgres create --project-id 1 \
  --type "" --default-schema public --username user --password pass --num-threads 4

dbtp credential --help          # list all adapters
```

Available adapters: `snowflake`, `postgres`, `redshift`, `bigquery`.

## How commands are generated

All command files in `pkg/cli/commands/*_gen.go` are **generated** -- do not
edit them by hand. They are produced by `cmd/codegen/` from three sources:

1. **Terraform resource schemas** -- flag names, types, required/optional,
   descriptions, and sensitive markers are extracted from the same schema
   definitions the Terraform provider uses.

2. **Go client methods** -- the `*Client` methods in `pkg/dbt_cloud/` are
   scanned to auto-derive get/create/update/delete operations, including
   parameter types, flag names, and API call expressions.

3. **YAML overrides** (`cli-resources/*.yaml`) -- for operations that can't be
   fully auto-derived (e.g., complex `create` methods, custom list validation),
   a YAML file supplies the operation definition.

See [`cmd/codegen/README.md`](../codegen/README.md) for the full details.
