package main

import (
	"context"
	"fmt"
	"os"

	dbtCli "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/cli"
	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/cli/commands"
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
			commands.ProjectCommands(),
			commands.EnvironmentCommands(),
			commands.JobCommands(),
			{
				Name:  "credential",
				Usage: "Manage warehouse credentials",
				Commands: []*cli.Command{
					commands.CredentialSnowflakeCommands(),
					commands.CredentialPostgresCommands(),
					commands.CredentialRedshiftCommands(),
					commands.CredentialBigqueryCommands(),
				},
			},
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
