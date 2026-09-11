# mise-bump-action

`mise.toml` (mise-en-place) で管理しているツールのバージョンを、Dependabotと同じPR体裁で
自動追従させるGitHub Action。

- 言語: Go
- 配布形態: composite action(v0はlinux/amd64向け事前ビルドバイナリをGitHub Releaseで配布)
- 対象: `[tools]`直下の短縮名 / `aqua:owner/repo` / `go:module/path` など、mise CLIが解決できる全バックエンド

## 背景

Dependabotの`dependabot.yml`は公式`package-ecosystem`しか受け付けず、`mise.toml`を検知できない。
Renovateはmiseをネイティブサポートするがログインユーザーが限定される制約があり不採用。
そこで、Dependabotと同じPR体裁を出すactionを自作する。詳細な経緯は
[.agents/docs/specs/2026-09-11-mise-bump-action-design.md](.agents/docs/specs/2026-09-11-mise-bump-action-design.md)。

## 設計判断

確定済みの判断は [`.agents/docs/adr/`](.agents/docs/adr/README.md) に1ファイル1決定で記録する。

## 詳細ガイド

- [README.md](README.md) — プロジェクト概要 + 利用側での使い方
- [.agents/docs/README.md](.agents/docs/README.md) — ドキュメント索引
- [.agents/docs/adr/README.md](.agents/docs/adr/README.md) — アーキテクチャ決定記録

## コードコメント

WHY が非自明な場合のみ1〜2行で書く。WHAT は識別子・構造で伝える。経緯・代替案の比較・
背景情報はコードコメントに書かず `.agents/docs/adr/` にADRとして記録し、コードからは
`(ADR NNNN)` の形で参照する。

## Non-Interactive Shell Commands

エージェントが対話プロンプトでハングしないよう、ファイル操作は必ず非対話フラグを使う:

```bash
cp -f source dest         # NOT: cp source dest
mv -f source dest         # NOT: mv source dest
rm -f file                # NOT: rm file
rm -rf directory          # NOT: rm -r directory
```

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
