# ADR 0003: PRフォーマットはDependabot公式PRの規約(bump系タイトル + updated-dependenciesトレーラー)を踏襲する

## Status

Accepted

## Date

2026-09-11

## Context

実際にDependabotが出したPRを参照すると、Terraformのようなrange指定
依存では`chore(deps): update X requirement from A to B`、commitメッセージには
`updated-dependencies:`から始まるYAML trailer(`dependency-name`/`dependency-version`/`dependency-type`)が
付与されている。mise管理下のツールは`go = "1.26.1"`のような厳密ピンであり、range指定ではない。

## Decision

- PRタイトルはDependabotの`bump`系規約を採用する: `chore(deps): bump <tool> from <old> to <new>`
- commitメッセージには`updated-dependencies:`YAML trailerをそのまま踏襲して付与する。
- ラベルは`dependencies`を固定で付与し、必要に応じて`mise`ラベルを追加する。

## Rules

- rangeではなく厳密ピンの更新であるため、`update ... requirement`ではなく`bump`系タイトルのみを使う。
- 複数ファイルを対象にする場合はタイトル末尾に対象パスを付与する(例: `... in /mise.toml`)。

## Consequences

- Dependabotに慣れたレビュアーがそのままの感覚でPRを読める。
- commit trailerの規約を厳密に模倣する分、Dependabot本体の書式が将来変わった場合に追従コストが発生する。

## References

- `.agents/docs/specs/2026-09-11-mise-bump-action-design.md`
