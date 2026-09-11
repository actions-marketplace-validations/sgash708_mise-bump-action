package outdated

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		fixtureFile string
		inlineJSON  string
		configPath  string
		want        []Entry
	}{
		{
			name:        "skips tools already up to date",
			fixtureFile: "up_to_date.json",
			configPath:  "/repo/mise.toml",
			want:        nil,
		},
		{
			name:        "detects outdated and normalizes rel path",
			fixtureFile: "with_outdated.json",
			configPath:  "/repo/mise.toml",
			want:        []Entry{{Name: "go", Requested: "1.26.1", Latest: "1.27.0", RelPath: "mise.toml"}},
		},
		{
			// go:github.com/matryer/moq has requested="v0.6.0" and latest="0.6.0" —
			// the same version once the "v" prefix is normalized, and must NOT be
			// reported as outdated.
			name: "normalizes v prefix before comparing",
			inlineJSON: `{
				"go:github.com/matryer/moq": {
					"name": "go:github.com/matryer/moq",
					"requested": "v0.6.0",
					"current": null,
					"bump": null,
					"latest": "0.6.0",
					"source": { "type": "mise.toml", "path": "/repo/mise.toml" }
				}
			}`,
			configPath: "/repo/mise.toml",
			want:       nil,
		},
		{
			// ADR 0006
			name: "uses bump instead of latest for a fuzzy pin",
			inlineJSON: `{
				"node": {
					"name": "node",
					"requested": "2.12",
					"current": "2.12.5",
					"bump": "2.13",
					"latest": "2.13.2",
					"source": { "type": "mise.toml", "path": "/repo/mise.toml" }
				}
			}`,
			configPath: "/repo/mise.toml",
			want:       []Entry{{Name: "node", Requested: "2.12", Latest: "2.13", RelPath: "mise.toml"}},
		},
		{
			// troubleshooting.md
			name: "excludes entries from a merged parent or global config",
			inlineJSON: `{
				"go": {
					"name": "go",
					"requested": "1.26.1",
					"current": "1.26.1",
					"bump": "1.27.0",
					"latest": "1.27.0",
					"source": { "type": "mise.toml", "path": "/home/user/.config/mise/config.toml" }
				},
				"node": {
					"name": "node",
					"requested": "24.12.0",
					"current": "24.12.0",
					"bump": "24.13.0",
					"latest": "24.13.0",
					"source": { "type": "mise.toml", "path": "/repo/mise.toml" }
				}
			}`,
			configPath: "/repo/mise.toml",
			want:       []Entry{{Name: "node", Requested: "24.12.0", Latest: "24.13.0", RelPath: "mise.toml"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data []byte
			if tt.fixtureFile != "" {
				var err error
				data, err = os.ReadFile(filepath.Join("testdata", tt.fixtureFile))
				if err != nil {
					t.Fatalf("failed to read fixture: %v", err)
				}
			} else {
				data = []byte(tt.inlineJSON)
			}

			got, err := Parse(data, "/repo", tt.configPath)
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d outdated entries, got %d: %+v", len(tt.want), len(got), got)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("entry %d mismatch:\n got  %+v\n want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// installFakeMise writes script as "mise" in a fresh temp dir and prepends
// it to PATH, so Run's real exec.CommandContext invocation runs against a
// stand-in instead of requiring the real mise CLI.
func installFakeMise(t *testing.T, script string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake mise script requires a POSIX shell")
	}

	binDir := t.TempDir()
	scriptPath := filepath.Join(binDir, "mise")
	if err := os.WriteFile(scriptPath, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write fake mise script: %v", err)
	}

	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestRun(t *testing.T) {
	tests := []struct {
		name          string
		script        func(repoRoot string) string
		wantEntries   []Entry
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "invokes mise and parses its stdout",
			script: func(repoRoot string) string {
				return fmt.Sprintf(`#!/bin/sh
if [ "$1" != "outdated" ] || [ "$2" != "--json" ] || [ "$3" != "--bump" ] || [ "$4" != "-C" ]; then
  echo "unexpected args: $@" >&2
  exit 1
fi
cat <<'EOF'
{
  "go": {
    "name": "go",
    "requested": "1.26.1",
    "current": "1.26.1",
    "bump": "1.27.0",
    "latest": "1.27.0",
    "source": { "type": "mise.toml", "path": "%s/mise.toml" }
  }
}
EOF
`, repoRoot)
			},
			wantEntries: []Entry{{Name: "go", Requested: "1.26.1", Latest: "1.27.0", RelPath: "mise.toml"}},
		},
		{
			// A tool declared in a merged parent/global config must not leak
			// into the result for this specific mise.toml.
			name: "excludes merged config entries not matching the target file",
			script: func(repoRoot string) string {
				return fmt.Sprintf(`#!/bin/sh
cat <<'EOF'
{
  "go": {
    "name": "go",
    "requested": "1.26.1",
    "current": "1.26.1",
    "bump": "1.27.0",
    "latest": "1.27.0",
    "source": { "type": "mise.toml", "path": "/somewhere/else/mise.toml" }
  },
  "node": {
    "name": "node",
    "requested": "24.12.0",
    "current": "24.12.0",
    "bump": "24.13.0",
    "latest": "24.13.0",
    "source": { "type": "mise.toml", "path": "%s/mise.toml" }
  }
}
EOF
`, repoRoot)
			},
			wantEntries: []Entry{{Name: "node", Requested: "24.12.0", Latest: "24.13.0", RelPath: "mise.toml"}},
		},
		{
			name: "wraps command failure with stderr",
			script: func(repoRoot string) string {
				return `#!/bin/sh
echo "mise: network unreachable" >&2
exit 1
`
			},
			wantErr:       true,
			wantErrSubstr: "network unreachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := t.TempDir()
			installFakeMise(t, tt.script(repoRoot))

			entries, err := Run(context.Background(), repoRoot, "mise.toml")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Errorf("expected error to contain %q, got: %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if len(entries) != len(tt.wantEntries) {
				t.Fatalf("expected %d entries, got %d: %+v", len(tt.wantEntries), len(entries), entries)
			}
			for i := range tt.wantEntries {
				if entries[i] != tt.wantEntries[i] {
					t.Errorf("entry %d mismatch:\n got  %+v\n want %+v", i, entries[i], tt.wantEntries[i])
				}
			}
		})
	}
}
