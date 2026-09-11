# ADR 0001: ツールのバージョン解決はmise CLIに委譲し、複数バックエンドを自前実装しない

## Status

Accepted

## Date

2026-09-11

## Context

`mise.toml`が管理するツールは`[tools]`直下の短縮名だけでなく、`"aqua:owner/repo"`(aquaレジストリ経由)、
`"go:module/path"`(go install経由)など複数バックエンドが混在するケースがある。
これらの最新バージョン解決ロジックをmise-bump-action側で再実装すると、mise本体の対応バックエンド追加・
変更に追従できず陳腐化する。

## Decision

バージョン差分検出は`mise outdated --json`相当をシェルアウトして呼び出し、mise CLI自身に解決を任せる。
mise-bump-actionは差分結果を受け取ってPR生成ロジックだけを担当する。

## Rules

- 利用側のGitHub Actions workflowでは`jdx/mise-action`等でmise本体のセットアップを事前に行うことを前提とする。
- mise-bump-action内にバックエンド固有(aqua/go install/asdf等)の版解決コードを書かない。

## Consequences

- mise本体が新しいバックエンドに対応すれば、mise-bump-actionは無改修で追従できる。
- 実行環境に`mise`のインストールが前提になる(ADR 0004のv0スコープにも影響)。

## References

- `.agents/docs/specs/2026-09-11-mise-bump-action-design.md`
