# mise-bump-action

**English** | [日本語](README.ja.md)

A GitHub Action that keeps tools managed by [mise](https://mise.jdx.dev/) (`mise.toml`) up to date, opening pull requests in the same style as [Dependabot](https://docs.github.com/en/code-security/dependabot).

Dependabot's `dependabot.yml` only understands its own built-in `package-ecosystem` list and can't see `mise.toml`. This action delegates version resolution to `mise` itself (`mise outdated`) and opens a Dependabot-style pull request for every tool that's behind.

Design rationale lives in [.agents/docs/adr/](.agents/docs/adr/README.md); the full design doc is at [.agents/docs/specs/2026-09-11-mise-bump-action-design.md](.agents/docs/specs/2026-09-11-mise-bump-action-design.md).

## Usage

Add a workflow like this to `.github/workflows/` in the repository that uses it:

```yaml
on:
  schedule:
    - cron: "0 3 * * 1"
  workflow_dispatch: {}

jobs:
  mise-bump:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write
    steps:
      - uses: actions/checkout@v7
      - uses: jdx/mise-action@v4
      - uses: sgash708/mise-bump-action@v1
        with:
          mise-config-path: mise.toml
          pr-strategy: per-tool
```

## Examples

Ready-to-copy workflows for `.github/workflows/mise-bump.yml`:

| File | Use case |
|---|---|
| [`examples/single-tool/mise-bump.yml`](examples/single-tool/mise-bump.yml) | Basic single-file setup (`pr-strategy: per-tool`) |
| [`examples/grouped-tools/mise-bump.yml`](examples/grouped-tools/mise-bump.yml) | Bundle multiple tools into one PR (`pr-strategy: single`) |
| [`examples/monorepo/mise-bump.yml`](examples/monorepo/mise-bump.yml) | Monorepo with multiple `mise-config-path` values |

See [examples/](examples/README.md) for details on each pattern.

> **Note:** GitHub requires a maintainer to manually approve the first workflow
> run triggered on a pull request opened via `GITHUB_TOKEN` (even for
> same-repo, non-fork PRs) — this applies to the CI checks on PRs this action
> opens. Approve once from the PR's checks tab (or `gh api -X POST
> repos/{owner}/{repo}/actions/runs/{run_id}/approve`) and it won't ask again
> for that branch.

## Inputs

| input | description | default |
|---|---|---|
| `mise-config-path` | Path to the target `mise.toml`. For multiple files, pass a newline-separated multi-line string | `mise.toml` |
| `pr-strategy` | `per-tool` (one PR per tool) or `single` (bundle everything into one PR) | `per-tool` |
| `labels` | Labels to apply, comma-separated | `dependencies` |
| `base-branch` | Base branch for pull requests | the ref that triggered the run (`GITHUB_REF_NAME`) |
| `dry-run` | If `true`, print the intended pull request title/body/diff to the job summary without creating any branch or pull request | `false` |

## Pull request lifecycle

- Closing a pull request without merging it means "don't reopen this exact version" — the next run won't recreate it. A newer version is still proposed normally.
- With `pr-strategy: per-tool`, opening a new pull request for a tool automatically closes any older still-open pull request for that same tool, with a comment pointing at the new one.

## Tech stack

- Go
- Distribution: composite action (pre-built linux/amd64 and linux/arm64 binaries via GitHub Releases; macOS/Windows runners are not supported)
- Auth: the calling repository's default `GITHUB_TOKEN` only (no extra PAT required)

## Directory structure

```
.
├── AGENTS.md              # Project overview (for agents)
├── CLAUDE.md              # Symlink to AGENTS.md
├── README.md              # This file
├── examples/              # Configuration patterns (single-tool/ is a live demo, others are illustrative)
└── .agents/
    └── docs/
        ├── README.md      # Documentation index
        ├── adr/           # Architecture decision records
        ├── specs/         # Design docs
        └── plans/         # Implementation plans
```

## License

MIT
