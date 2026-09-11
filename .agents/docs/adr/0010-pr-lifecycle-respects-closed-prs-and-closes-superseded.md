# ADR 0010: closeされたPRは再生成せず、古いバージョンのPRはsupersededとして自動closeする

## Status

Accepted

## Date

2026-09-12

## Context

Fableによる敵対的レビューで、以下2点が実運用上の致命的な欠陥として指摘された。

- **closed PRの復活**: `OpenBumpPR`は「対象ブランチにopenなPRがあるか」しか見ておらず、
  ユーザーが不要と判断してcloseした(mergeしていない)PRが、次回の定期実行で同じブランチ名
  ・同じバージョンのまま再生成されてしまう。Dependabotであれば、closeしたPRのバージョンは
  再提案しない。
- **superseded PRの放置**: ブランチ名にバージョンを含むため、あるツールのPRがopenのまま
  さらに新しいバージョンが出ると、古いPRをcloseしないまま新しいPRが追加され、同じツールの
  PRが並存し続ける。

## Decision

`internal/githubapi.Client.OpenBumpPR`に以下を追加する。

- ブランチ名で`state=open`のPRが無い場合、続けて`state=closed`のPRを検索し、
  `merged_at`が`null`(mergeされずにcloseされた)ものが見つかれば、書き込みを一切行わずに
  `runner.ErrClosedPreviously`を返す。`runner.Run`/`bumpGroup`はこのエラーを失敗として
  扱わず、スキップとして処理し(PR番号にもエラーにもカウントしない)、`out`
  ($GITHUB_STEP_SUMMARY)にスキップした旨を書き出す。
- `runner.BumpPRInput.BranchPrefix`(例: `mise-bump/go-`)が設定されている場合、新しいPRを
  開いた後、同じprefixを持つ他のopen PR(=同じツールの別バージョン)を検索し、closeして
  「Superseded by #N」とコメントする。`BranchPrefix`は単一エントリのグループでのみ設定され、
  グループ化(`single`戦略で複数ツールをまとめたPR)では設定しない
  (ブランチ名がハッシュベースになりツール単位のprefixが存在しないため)。

## Rules

- ブランチ名にバージョンを含める設計(既存)は維持し、「closeされた特定バージョン」だけを
  再提案しないようにする。新しいバージョンが出れば通常通り新しいPRを提案する。
- superseded-close は best-effort とする(検索・close・コメントいずれかのAPI呼び出しが
  失敗しても、新しいPRの作成自体は既に成功しているため、全体の失敗として扱わない)。
- superseded-close の対象は同一ツール(`BranchPrefix`一致)のみ。異なるツールのPRには
  一切干渉しない。

## Consequences

- Dependabotに近い「closeしたら黙る、mergeされたら次のバージョンだけ提案する」という
  振る舞いになる。
- `single`戦略(複数ツールを1PRにまとめる)では、superseded-close機能の恩恵を受けない。
  ツールごとに`pr-strategy: per-tool`を使う場合にのみ有効。
- `OpenBumpPR`のAPI呼び出し回数が増える(closed PR検索・superseded検索が追加)。

## References

- `internal/githubapi/client.go`(`findClosedUnmergedPR`, `closeSupersededPRs`)
- `internal/runner/runner.go`(`ErrClosedPreviously`, `branchPrefix`)
