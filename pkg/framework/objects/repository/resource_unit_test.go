package repository

import "testing"

func intPtr(v int) *int {
	return &v
}

// TestRepositoryGitCloneStrategyStatePreservation covers the three git_clone_strategy /
// gitlab_project_id drift behaviours fixed in resource.go:
//  1. gitlab_project_id must only be sent to the API when git_clone_strategy is
//     "deploy_token".
//  2. on create, the API's normalised git_clone_strategy is only adopted when a
//     GitHub App installation is actually attached; otherwise the planned value wins.
//  3. on read/update, an API response of "github_app" with no installation attached
//     must not overwrite the prior (state/planned) git_clone_strategy.
func TestRepositoryGitCloneStrategyStatePreservation(t *testing.T) {
	t.Run("decideGitlabProjectIDToSend", func(t *testing.T) {
		cases := []struct {
			name             string
			gitCloneStrategy string
			planned          int
			want             int
		}{
			{"deploy_token sends the planned id", "deploy_token", 42, 42},
			{"github_app strategy sends zero", "github_app", 42, 0},
			{"azure_active_directory_app strategy sends zero", "azure_active_directory_app", 42, 0},
			{"empty strategy sends zero", "", 42, 0},
			{"deploy_token with zero planned id stays zero", "deploy_token", 0, 0},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := decideGitlabProjectIDToSend(tc.gitCloneStrategy, tc.planned)
				if got != tc.want {
					t.Errorf("decideGitlabProjectIDToSend(%q, %d) = %d, want %d", tc.gitCloneStrategy, tc.planned, got, tc.want)
				}
			})
		}
	})

	t.Run("decideCreateGitCloneStrategy", func(t *testing.T) {
		cases := []struct {
			name                 string
			apiGitCloneStrategy  string
			apiGithubInstallID   *int
			plannedGitCloneStrat string
			want                 string
		}{
			{"nil installation id keeps planned value", "github_app", nil, "deploy_token", "deploy_token"},
			{"zero installation id keeps planned value", "github_app", intPtr(0), "deploy_token", "deploy_token"},
			{"non-zero installation id adopts api value", "github_app", intPtr(99), "deploy_token", "github_app"},
			{"non-zero installation id adopts api value even if matching", "github_app", intPtr(99), "github_app", "github_app"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := decideCreateGitCloneStrategy(tc.apiGitCloneStrategy, tc.apiGithubInstallID, tc.plannedGitCloneStrat)
				if got != tc.want {
					t.Errorf("decideCreateGitCloneStrategy(%q, %v, %q) = %q, want %q",
						tc.apiGitCloneStrategy, tc.apiGithubInstallID, tc.plannedGitCloneStrat, got, tc.want)
				}
			})
		}
	})

	t.Run("decideGitCloneStrategyDriftGuard", func(t *testing.T) {
		cases := []struct {
			name                string
			apiGitCloneStrategy string
			apiGithubInstallID  *int
			priorGitCloneStrat  string
			want                string
		}{
			{"github_app with no installation id preserves prior value", "github_app", nil, "deploy_token", "deploy_token"},
			{"github_app with zero installation id preserves prior value", "github_app", intPtr(0), "deploy_token", "deploy_token"},
			{"github_app with real installation id adopts api value", "github_app", intPtr(7), "deploy_token", "github_app"},
			{"non github_app value always adopted", "deploy_token", nil, "github_app", "deploy_token"},
			{"azure strategy always adopted regardless of installation id", "azure_active_directory_app", nil, "github_app", "azure_active_directory_app"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got := decideGitCloneStrategyDriftGuard(tc.apiGitCloneStrategy, tc.apiGithubInstallID, tc.priorGitCloneStrat)
				if got != tc.want {
					t.Errorf("decideGitCloneStrategyDriftGuard(%q, %v, %q) = %q, want %q",
						tc.apiGitCloneStrategy, tc.apiGithubInstallID, tc.priorGitCloneStrat, got, tc.want)
				}
			})
		}
	})
}
