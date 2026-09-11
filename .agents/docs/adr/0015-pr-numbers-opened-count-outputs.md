# ADR 0015: `pr-numbers`/`opened-count`をaction outputとして公開する

## Status

Accepted

## Date

2026-09-12

## Context

これまでmise-bump-actionは開いたPR番号を`stderr`にしか出力しておらず、呼び出し側の
workflowが後続ステップで「今回何個・どのPRが開かれたか」を使うことができなかった
(例: Slack通知、レビュー依頼の自動割当)。GitHub Actionsの標準的な仕組みである
action outputを使えばこれが可能になる。

## Decision

- `cmd/mise-bump-action/main.go`が`$GITHUB_OUTPUT`に`opened-count`(件数)と
  `pr-numbers`(カンマ区切りのPR番号、無ければ空文字列)を書き込む。
- 出力はdry-run時・"no outdated tools"の早期return時・部分失敗
  (`errors.Join`で一部グループが失敗)時も含め、常に書き込む。呼び出し側が
  常に参照できるようにするため。
- `action.yml`の`Run mise-bump-action`ステップに`id: run`を付与し、
  `outputs.pr-numbers`/`outputs.opened-count`として公開する。

## Rules

- 新しい出力を追加する場合も、成功/失敗/dry-runの全パスで一貫して値を設定する
  (未設定のまま呼び出し側から参照されて空扱いになる、という状態を避ける)。

## Consequences

- 呼び出し側のworkflowが`steps.<id>.outputs.pr-numbers`を使って後続処理を組める
  ようになった。
- `$GITHUB_OUTPUT`はGitHub Actions実行環境でのみ設定される。それ以外の環境で
  `mise-bump-action`を実行した場合、出力は単に破棄される(エラーにはしない)。

## References

- `cmd/mise-bump-action/main.go`(`writeOutputs`)
- `action.yml`(`outputs:`)
