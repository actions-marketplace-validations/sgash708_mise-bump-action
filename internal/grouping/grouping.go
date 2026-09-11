// Package grouping splits outdated tool entries into pull request groups
// according to the configured pr-strategy.
package grouping

import (
	"fmt"

	"github.com/sgash708/mise-bump-action/internal/outdated"
)

// Strategy controls how outdated entries are grouped into pull requests.
type Strategy string

const (
	// PerTool opens one pull request per outdated tool, matching Dependabot's
	// default behavior.
	PerTool Strategy = "per-tool"
	// Single bundles all outdated tools into a single pull request.
	Single Strategy = "single"
)

// PRGroup is a set of outdated entries that will be bumped together in one
// pull request.
type PRGroup struct {
	Entries []outdated.Entry
}

// Group splits entries into pull request groups according to strategy. An
// empty entries slice always yields zero groups.
func Group(entries []outdated.Entry, strategy Strategy) ([]PRGroup, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	switch strategy {
	case PerTool, "":
		groups := make([]PRGroup, len(entries))
		for i, e := range entries {
			groups[i] = PRGroup{Entries: []outdated.Entry{e}}
		}
		return groups, nil
	case Single:
		return groupByFile(entries), nil
	default:
		return nil, fmt.Errorf("unknown pr-strategy %q", strategy)
	}
}

// groupByFile bundles entries sharing the same RelPath into one group each,
// since a single bumped pull request rewrites exactly one file (see
// runner.bumpGroup). Groups are emitted in first-seen file order.
func groupByFile(entries []outdated.Entry) []PRGroup {
	var order []string
	byPath := make(map[string][]outdated.Entry)
	for _, e := range entries {
		if _, seen := byPath[e.RelPath]; !seen {
			order = append(order, e.RelPath)
		}
		byPath[e.RelPath] = append(byPath[e.RelPath], e)
	}

	groups := make([]PRGroup, len(order))
	for i, path := range order {
		groups[i] = PRGroup{Entries: byPath[path]}
	}
	return groups
}
