# ADR 0007: アップストリームrelease本文の@メンション/#issue参照をサニタイズしてから埋め込む

## Status

Accepted

## Date

2026-09-11

## Context

`internal/githubapi.ReleaseNotesHTML`/`CommitsHTML`は、bump対象ツールのアップストリーム
リポジトリのrelease本文・コミット件名をそのままmise-bump-actionが開くPR本文に埋め込む。
これらの文字列に含まれる`@user`や`#123`をそのまま埋め込むと、GitHubは**このリポジトリの
コンテキストで**解釈する。

- `@user`: アップストリームで意図された言及であっても、このリポジトリ上で実在するアカウント
  への意図しないメンション通知になる。
- `#123`: アップストリームのissue/PR番号のつもりでも、このリポジトリの`#123`へクロス
  リンクしてしまう。

実際に`@v0`(floating major tagを指す自リポジトリのコミット件名)がGitHubの実在ユーザー名
"v0"と衝突し、無関係なアカウントがreleaseのContributorとして表示される実害が発生した
(`.agents/docs/troubleshooting.md`参照)。この経験から、自リポジトリ発の文字列だけでなく
サードパーティ由来の文字列も同じリスクを持つと判断した。

## Decision

`sanitizeReleaseBody`で、埋め込み前に以下を機械的に変換する(Dependabot本家と同じ方式)。

- `@user` → `@​ user`(ゼロ幅スペースを挿入しメンション化を防ぐ)
- `#123` → `[#123](https://redirect.github.com/{repo}/issues/123)`(アップストリーム
  リポジトリへの明示リンクにし、自リポジトリへのクロスリンク化を防ぐ。
  `redirect.github.com`経由なので対象issueへの通知も発生しない)

## Rules

- サードパーティ由来のテキスト(release本文・コミット件名)をPR本文やrelease notesに
  埋め込む処理は、必ず`sanitizeReleaseBody`を経由させる。
- 新しい埋め込み元を追加する場合も同じ関数を再利用し、個別にメンション対策を書かない。

## Consequences

- アップストリームの`@user`/`#123`は正しくレンダリングされるが、リンク先が変わる
  (`redirect.github.com`経由になる)。
- 自リポジトリのコミット件名にも同じリスクがある(実際に`@v0`で発生した)ため、
  `.github/workflows/release.yml`のリリースノート生成でも同じ「バッククォートで
  囲んでインラインコード化する」方式で`@name`パターンを機械的にエスケープしている。

## References

- `internal/githubapi/enrichment.go`(`sanitizeReleaseBody`)
- `.github/workflows/release.yml`(コミット件名の`@name`エスケープ)
- `.agents/docs/troubleshooting.md`(release notesの@メンション化の実例)
