# ADR 0002: PR作成はGitHub API直呼び出しで自前実装し、サードパーティactionに依存しない

## Status

Accepted

## Date

2026-09-11

## Context

branch作成・commit・PR openを担う定番のサードパーティaction(`peter-evans/create-pull-request`等)は
広く使われているが、外部actionへの依存はサプライチェーンリスクになる。mise-bump-actionはmise管理下の
ツールバージョンという、比較的機微な性質を持つ設定ファイルを書き換えるため、依存を最小化したい。

## Decision

branch作成・commit・PR作成はGoバイナリ内からGitHub REST APIを直接呼び出して自前実装する。
サードパーティのPR作成actionは使わない。

## Rules

- git操作(branch作成/commit/push)はGitHub API経由で行い、ローカルの`git`コマンド実行にも依存しない。
- 認証は利用側リポジトリの既定`GITHUB_TOKEN`(`contents: write`, `pull-requests: write`権限)のみを前提とする。PAT等の追加シークレットは要求しない。

## Consequences

- 依存actionが増えないため、サプライチェーン攻撃面が小さい。
- git操作をGoで自前実装するコストがかかる(`peter-evans/create-pull-request`を使う場合に比べて実装量が増える)。

## References

- `.agents/docs/specs/2026-09-11-mise-bump-action-design.md`
