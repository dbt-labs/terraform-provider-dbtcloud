package commands

import (
	"fmt"

	dbtCli "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/cli"
	dbtcloud "github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
	"github.com/urfave/cli/v3"
)

// getClient authenticates and returns a dbt Cloud API client.
func getClient() (*dbtcloud.Client, error) {
	cfg, err := dbtCli.LoadAuthConfig()
	if err != nil {
		return nil, fmt.Errorf("not authenticated: %w", err)
	}
	client, err := dbtCli.NewClientFromAuth(cfg)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return client, nil
}

// getFormat reads the global --format flag from the root command.
func getFormat(cmd *cli.Command) string {
	return cmd.Root().String("format")
}

// stateLabel returns a human-readable label for a resource state integer.
func stateLabel(state int) string {
	if state == dbtcloud.STATE_DELETED {
		return "deleted"
	}
	return "active"
}
