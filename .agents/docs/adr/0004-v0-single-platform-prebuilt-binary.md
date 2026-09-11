# ADR 0004: v0はlinux/amd64向け事前ビルドバイナリのみで配布し、将来のマルチプラットフォーム移行に備えてインターフェースを固定する

## Status

Accepted

## Date

2026-09-11

## Context

配布方式には(1) GoReleaser等でOS/arch別バイナリをビルドしGitHub Releaseに置く方式、
(2) 実利用環境(GitHub-hosted ubuntu runner = linux/amd64)のみを対象にビルド済みバイナリを配布する方式がある。
v0時点での実利用は全てGitHub-hosted ubuntu runnerであり、複数OS/arch対応は現時点では過剰。

## Decision

v0はlinux/amd64向けにビルドしたバイナリのみをGitHub Releaseに添付し、composite actionが実行時に
タグ指定でダウンロードする方式にする。`go install`によるランタイムビルドは採用しない。

## Rules

- composite actionの利用側インターフェース(`uses: sgash708/mise-bump-action@vX`)は、将来GoReleaser等で
  複数OS/arch対応に移行しても変更しない。
- リリースアセット名は将来の複数OS/arch対応を見据えた命名規則(`mise-bump-action_<os>_<arch>`等)にしておく。

## Consequences

- v0のリリースパイプラインが単純になる。
- mac/arm等の他環境での利用要望が出た場合、GoReleaser移行が必要になる(利用側の呼び出し方は変わらない想定)。

## References

- `.agents/docs/specs/2026-09-11-mise-bump-action-design.md`
