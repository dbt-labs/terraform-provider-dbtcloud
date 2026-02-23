package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dbt-labs/terraform-provider-dbtcloud/internal/codegen"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/bigquery_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/environment"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/job"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/postgres_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/project"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/redshift_credential"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/framework/objects/snowflake_credential"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// cliResource defines a resource for CLI code generation.
type cliResource struct {
	Name          string                // CLI command name (e.g., "project")
	FuncPrefix    string                // prefix for generated function names; defaults to Name
	GoType        string                // Go struct for detail view (e.g., "Project")
	ListType      string                // Go struct for list view, defaults to GoType
	TFResource    resource.Resource     // TF resource (for schema extraction)
	TFPluralDS    datasource.DataSource // TF plural datasource (for filter attrs)
	DefaultDelete *codegen.Operation    // fallback delete op when none is auto-derived
}

// Resource registry: adding a new resource = adding one line here.
var cliResources = []cliResource{
	{
		Name:       "project",
		GoType:     "Project",
		TFResource: project.ProjectResource(),
		TFPluralDS: project.ProjectsDataSource(),
	},
	{
		Name:       "environment",
		GoType:     "Environment",
		TFResource: environment.EnvironmentResource(),
		TFPluralDS: environment.EnvironmentsDataSource(),
	},
	{
		Name:       "job",
		GoType:     "Job",
		ListType:   "JobWithEnvironment",
		TFResource: job.JobResource(),
		TFPluralDS: job.JobsDataSource(),
	},
	{
		Name:          "snowflake",
		FuncPrefix:    "credentialSnowflake",
		GoType:        "SnowflakeCredential",
		TFResource:    snowflake_credential.SnowflakeCredentialResource(),
		DefaultDelete: credentialDelete(),
	},
	{
		Name:          "postgres",
		FuncPrefix:    "credentialPostgres",
		GoType:        "PostgresCredential",
		TFResource:    postgres_credential.PostgresCredentialResource(),
		DefaultDelete: credentialDelete(),
	},
	{
		Name:          "redshift",
		FuncPrefix:    "credentialRedshift",
		GoType:        "RedshiftCredential",
		TFResource:    redshift_credential.RedshiftCredentialResource(),
		DefaultDelete: credentialDelete(),
	},
	{
		Name:          "bigquery",
		FuncPrefix:    "credentialBigquery",
		GoType:        "BigQueryCredential",
		TFResource:    bigquery_credential.BigqueryCredentialResource(),
		DefaultDelete: credentialDelete(),
	},
}

func main() {
	sourceDir := "pkg/dbt_cloud"
	resourceDir := "cli-resources"
	outputDir := "pkg/cli/commands"

	if len(os.Args) > 1 {
		resourceDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		outputDir = os.Args[2]
	}

	// Step 1: Parse struct field types from Go source (for template type lookups).
	if err := codegen.ParseFieldTypes(sourceDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing field types: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Parse json tags from Go structs.
	structTags, err := codegen.ParseStructTags(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing struct tags: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Scan *Client methods.
	methods, err := codegen.ScanMethods(sourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning methods: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Load YAML overrides from cli-resources/.
	yamlOverrides := loadYAMLOverrides(resourceDir)

	// Step 5: For each resource, extract TF schemas → derive → merge → generate.
	for _, res := range cliResources {
		listType := res.ListType
		if listType == "" {
			listType = res.GoType
		}

		// Extract TF schemas
		resourceAttrs, err := codegen.ExtractResourceAttrs(res.TFResource)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting resource attrs for %s: %v\n", res.Name, err)
			os.Exit(1)
		}

		var filterAttrs []codegen.TFAttrInfo
		if res.TFPluralDS != nil {
			filterAttrs, err = codegen.ExtractDatasourceFilterAttrs(res.TFPluralDS)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error extracting filter attrs for %s: %v\n", res.Name, err)
				os.Exit(1)
			}
		}

		description := codegen.GetResourceDescription(res.TFResource)

		// Build config for derivation
		cfg := codegen.TFResourceConfig{
			Name:          res.Name,
			GoType:        res.GoType,
			ListType:      listType,
			ResourceAttrs: resourceAttrs,
			FilterAttrs:   filterAttrs,
			GoTypeTags:    structTags[res.GoType],
			ListTypeTags:  structTags[listType],
			AllStructTags: structTags,
			Methods:       methods[res.GoType],
			Description:   description,
		}

		derived, err := codegen.DeriveTFResourceDef(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deriving resource def for %s: %v\n", res.Name, err)
			os.Exit(1)
		}

		// Merge with YAML override if one exists
		override := yamlOverrides[res.Name]
		final := codegen.MergeResourceDef(derived, override)

		// Pass through FuncPrefix from registry
		if res.FuncPrefix != "" {
			final.FuncPrefix = res.FuncPrefix
		}

		// Inject default delete if none was auto-derived or overridden
		if final.Delete == nil && res.DefaultDelete != nil {
			final.Delete = res.DefaultDelete
		}

		// Resolve extra columns per struct type (ListType for table rows, GoType for detail view)
		if len(final.ExtraColumns) > 0 {
			final.Columns = append(final.Columns, codegen.ResolveExtraColumns(final.ExtraColumns, structTags[final.ListType])...)
			final.Details = append(final.Details, codegen.ResolveExtraColumns(final.ExtraColumns, structTags[final.GoType])...)
			final.ExtraColumns = nil
		}

		fmt.Printf("Generating %s (from TF schema%s)...\n", final.Name, mergeLabel(override))
		if err := codegen.GenerateFromDef(final, outputDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", final.Name, err)
			os.Exit(1)
		}
	}

	fmt.Println("Done!")
}

func loadYAMLOverrides(resourceDir string) map[string]*codegen.ResourceDef {
	overrides := map[string]*codegen.ResourceDef{}
	entries, err := os.ReadDir(resourceDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading resource directory %s: %v\n", resourceDir, err)
		os.Exit(1)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		yamlPath := filepath.Join(resourceDir, entry.Name())
		res, err := codegen.LoadYAMLResourceDef(yamlPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading %s: %v\n", yamlPath, err)
			os.Exit(1)
		}

		name := strings.TrimSuffix(entry.Name(), ext)
		if res.Name == "" {
			res.Name = name
		}
		overrides[res.Name] = res
	}
	return overrides
}

// credentialDelete returns a shared delete operation for all credential types.
// All credentials use the same DeleteCredential(credentialId, projectId) endpoint.
func credentialDelete() *codegen.Operation {
	return &codegen.Operation{
		Flags: []codegen.Flag{
			{Name: "id", GoType: "int", Required: true, Usage: "Credential ID"},
			{Name: "project-id", GoType: "int", Required: true, Usage: "Project ID"},
		},
		Call:      `client.DeleteCredential(strconv.Itoa(int(cmd.Int("id"))), strconv.Itoa(int(cmd.Int("project-id"))))`,
		ResultVar: "credential",
	}
}

func mergeLabel(override *codegen.ResourceDef) string {
	if override == nil {
		return ""
	}
	if override.GoType != "" {
		return " + full YAML override"
	}
	return " + partial YAML override"
}
