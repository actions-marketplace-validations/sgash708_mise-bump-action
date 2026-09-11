// Package config parses this action's inputs from environment variables.
package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sgash708/mise-bump-action/internal/grouping"
)

// Config holds all inputs needed to run one invocation of mise-bump-action.
type Config struct {
	MiseConfigPaths []string
	PRStrategy      grouping.Strategy
	Labels          []string
	BaseBranch      string
	GitHubToken     string
	Repository      string
	APIURL          string
	DryRun          bool
	// Ignore lists tool-name patterns to exclude from bumping. Each pattern
	// is either an exact tool name or a prefix ending in "*".
	Ignore []string
	// MaxOpenPRs caps how many bump pull requests may be open at once
	// (existing ones carrying cfg.Labels count toward the cap). 0 means
	// unlimited.
	MaxOpenPRs int
}

const defaultAPIURL = "https://api.github.com"

// FromEnv builds a Config from environment variables. getenv is injected so
// tests do not depend on process-global environment state.
func FromEnv(getenv func(string) string) (Config, error) {
	token := getenv("GITHUB_TOKEN")
	if token == "" {
		return Config{}, fmt.Errorf("GITHUB_TOKEN is required")
	}

	repository := getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return Config{}, fmt.Errorf("GITHUB_REPOSITORY is required")
	}

	paths := splitNonEmpty(getenv("INPUT_MISE_CONFIG_PATH"), "\n")
	if len(paths) == 0 {
		paths = []string{"mise.toml"}
	}

	strategy := grouping.Strategy(getenv("INPUT_PR_STRATEGY"))
	if strategy == "" {
		strategy = grouping.PerTool
	}
	if strategy != grouping.PerTool && strategy != grouping.Single {
		return Config{}, fmt.Errorf("invalid pr-strategy %q: must be %q or %q", strategy, grouping.PerTool, grouping.Single)
	}

	labels := splitNonEmpty(getenv("INPUT_LABELS"), ",")
	if len(labels) == 0 {
		labels = []string{"dependencies"}
	}

	baseBranch := getenv("INPUT_BASE_BRANCH")
	if baseBranch == "" {
		baseBranch = getenv("GITHUB_REF_NAME")
	}
	if baseBranch == "" {
		return Config{}, fmt.Errorf("base-branch is required: set the base-branch input, or run this action on a workflow trigger where GITHUB_REF_NAME is a real branch (schedule or workflow_dispatch, not pull_request)")
	}
	if strings.Contains(baseBranch, "/merge") || strings.Contains(baseBranch, "/head") {
		return Config{}, fmt.Errorf("base-branch %q looks like a pull_request ref (GITHUB_REF_NAME is not a real branch on pull_request events); set the base-branch input explicitly", baseBranch)
	}

	apiURL := getenv("GITHUB_API_URL")
	if apiURL == "" {
		apiURL = defaultAPIURL
	}

	dryRun := strings.EqualFold(getenv("INPUT_DRY_RUN"), "true")

	ignore := splitNonEmpty(getenv("INPUT_IGNORE"), ",")

	maxOpenPRs := 0
	if raw := getenv("INPUT_MAX_OPEN_PRS"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			return Config{}, fmt.Errorf("invalid max-open-prs %q: must be a non-negative integer", raw)
		}
		maxOpenPRs = n
	}

	return Config{
		MiseConfigPaths: paths,
		PRStrategy:      strategy,
		Labels:          labels,
		BaseBranch:      baseBranch,
		GitHubToken:     token,
		Repository:      repository,
		APIURL:          apiURL,
		DryRun:          dryRun,
		Ignore:          ignore,
		MaxOpenPRs:      maxOpenPRs,
	}, nil
}

func splitNonEmpty(s, sep string) []string {
	if s == "" {
		return nil
	}
	var out []string
	for _, part := range strings.Split(s, sep) {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
