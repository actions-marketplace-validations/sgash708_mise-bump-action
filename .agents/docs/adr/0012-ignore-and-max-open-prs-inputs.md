# ADR 0012: `ignore`(ツール除外)と`max-open-prs`(PR数上限)を入力として追加する

## Status

Accepted

## Date

2026-09-12

## Context

Fableの敵対的レビューで、Renovate/Dependabotが持つ「除外設定」「PR数上限」に相当する
制御手段がmise-bump-actionには一切無いと指摘された。closed PRの再生成防止
(ADR 0010)だけでは「特定のツールを恒久的に対象外にしたい」「大量のPRが一度に
開かれるのを防ぎたい」というニーズをカバーできない。

major-versionの更新だけ抑止する機能(Dependabotの`ignore`の`update-types`相当)は
見送った。mise管理下のツールはsemverに従わないものも多く(CLIツール、言語のバージョン
表記等)、汎用的に「メジャーバージョンとは何か」を判定するロジックが脆くなるため。

## Decision

- `ignore` input: カンマ区切りのツール名パターン。完全一致、または末尾`*`による
  前方一致(例: `terraform`, `aqua:foo/*`)。マッチしたエントリは`grouping.Group`に
  渡す前に除外し、除外した旨をjob summaryに出力する。
- `max-open-prs` input: 整数。0(既定)は無制限。実行開始時に一度だけ、`labels`を
  持つopen PR数を数え、以後そのカウントに到達したら残りのグループの処理を打ち切る
  (job summaryにスキップ理由を出力)。カウント取得自体が失敗した場合は上限を
  強制せず処理を継続する(enrichment等と同じくbest-effort)。

## Rules

- `ignore`のパターンマッチングは単純な文字列前方一致のみとし、正規表現やglobの
  完全実装は行わない。
- `max-open-prs`のカウントは実行開始時の1回のみ行う。処理中に他プロセスがPRを
  作成/closeしても再カウントしない。

## Consequences

- ツール名を恒久的に除外したい場合や、一度に大量のPRが開かれるのを防ぎたい場合に
  対応できるようになった。
- メジャーバージョンだけを抑止する機能は提供しない。必要な場合は`ignore`で
  当該ツールを完全に除外するか、手動でPRをcloseする(ADR 0010によりそのバージョンは
  再提案されない)運用でカバーする。

## References

- `internal/config/config.go`(`Ignore`, `MaxOpenPRs`)
- `internal/runner/runner.go`(`filterIgnored`, `matchesAny`, `Run`のmax-open-prs処理)
- `internal/githubapi/client.go`(`CountOpenBumpPRs`)
