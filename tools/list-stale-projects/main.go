// list-stale-projects is a read-only diagnostic for the shared live dbt Cloud
// acceptance-test account. It lists every project in the account and reports
// how old each one is, so we can see how much cruft has accumulated from
// acceptance-test runs whose destroy step failed (leaving the project
// dangling) without needing to delete anything.
//
// It makes GET requests only. It does not delete or modify any resource.
//
// Usage (same env vars the acceptance tests already use):
//
//	DBT_CLOUD_ACCOUNT_ID=... DBT_CLOUD_TOKEN=... DBT_CLOUD_HOST_URL=... \
//	  go run ./tools/list-stale-projects [-min-age=1h]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/dbt-labs/terraform-provider-dbtcloud/pkg/dbt_cloud"
)

func mustEnv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		fmt.Fprintf(os.Stderr, "missing required env var %s\n", name)
		os.Exit(1)
	}
	return v
}

func main() {
	minAge := flag.Duration(
		"min-age",
		time.Hour,
		"only list projects older than this (a real CI run creates and destroys a project within ~20 minutes, so anything older than this is almost certainly orphaned)",
	)
	flag.Parse()

	accountID, err := strconv.ParseInt(mustEnv("DBT_CLOUD_ACCOUNT_ID"), 10, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid DBT_CLOUD_ACCOUNT_ID: %v\n", err)
		os.Exit(1)
	}
	token := mustEnv("DBT_CLOUD_TOKEN")
	// Matches acctest_helper.SharedClient()'s fallback: DBT_CLOUD_HOST_URL
	// isn't actually set as a secret anywhere in this repo, so the existing
	// acceptance-test workflows already rely on this default too.
	hostURL := os.Getenv("DBT_CLOUD_HOST_URL")
	if hostURL == "" {
		hostURL = "https://cloud.getdbt.com/api"
	}

	maxRetries := 3
	retryInterval := 10
	timeoutSeconds := 60

	client, err := dbt_cloud.NewClient(
		&accountID,
		&token,
		&hostURL,
		&maxRetries,
		&retryInterval,
		[]string{"429", "500", "502", "503", "504"},
		true, // skipCredentialsValidation - we're about to hit the API anyway
		&timeoutSeconds,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to build client: %v\n", err)
		os.Exit(1)
	}

	if body, err := client.GetEndpoint(client.BuildV2URL(dbt_cloud.ResourceAccounts)); err == nil {
		var ar dbt_cloud.AuthResponse
		if err := json.Unmarshal(body, &ar); err == nil {
			for _, acct := range ar.Data {
				if acct.Id == accountID {
					fmt.Printf("Account: %q (plan=%s, state=%d)\n\n", acct.Name, acct.Plan, acct.State)
				}
			}
		}
	}

	projects, err := client.GetAllProjects("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to list projects: %v\n", err)
		os.Exit(1)
	}

	now := time.Now().UTC()

	type aged struct {
		dbt_cloud.ProjectConnectionRepository
		age time.Duration
	}

	dateLayouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05Z07:00",
	}

	var all []aged
	var unparsed int
	var sampleRaw string
	for _, p := range projects {
		var createdAt time.Time
		var err error = fmt.Errorf("no layouts tried")
		for _, layout := range dateLayouts {
			createdAt, err = time.Parse(layout, p.CreatedAt)
			if err == nil {
				break
			}
		}
		if err != nil {
			unparsed++
			if sampleRaw == "" && p.CreatedAt != "" {
				sampleRaw = p.CreatedAt
			}
			continue
		}
		all = append(all, aged{p, now.Sub(createdAt.UTC())})
	}
	if unparsed > 0 && sampleRaw != "" {
		fmt.Fprintf(os.Stderr, "sample unparsed created_at value: %q\n", sampleRaw)
	}

	sort.Slice(all, func(i, j int) bool { return all[i].age > all[j].age })

	buckets := map[string]int{"<1h": 0, "1h-24h": 0, "1d-7d": 0, "7d-30d": 0, ">30d": 0}
	for _, a := range all {
		switch {
		case a.age < time.Hour:
			buckets["<1h"]++
		case a.age < 24*time.Hour:
			buckets["1h-24h"]++
		case a.age < 7*24*time.Hour:
			buckets["1d-7d"]++
		case a.age < 30*24*time.Hour:
			buckets["7d-30d"]++
		default:
			buckets[">30d"]++
		}
	}

	fmt.Printf("Total projects in account %d: %d (unparsed created_at: %d)\n", accountID, len(projects), unparsed)
	fmt.Println("Age distribution:")
	for _, k := range []string{"<1h", "1h-24h", "1d-7d", "7d-30d", ">30d"} {
		fmt.Printf("  %-8s %d\n", k, buckets[k])
	}
	fmt.Println()

	fmt.Printf("Candidates older than %s (likely orphaned from failed test destroys):\n", *minAge)
	fmt.Println("id\tage\tcreated_at\tname")
	var candidateCount int
	for _, a := range all {
		if a.age < *minAge {
			continue
		}
		candidateCount++
		fmt.Printf("%d\t%s\t%s\t%s\n", a.ID, a.age.Round(time.Minute), a.CreatedAt, a.Name)
	}
	fmt.Printf("\n%d candidate(s) out of %d total projects\n", candidateCount, len(projects))
}
