# ADR 0014: PR本文はGitHubの65,536文字上限に合わせて切り詰める

## Status

Accepted

## Date

2026-09-12

## Context

`internal/prtext.Build`は、リリースの多いツールの場合、アップストリームのrelease
本文・コミット一覧をそのままPR本文に埋め込む。GitHubのissue/PR本文には65,536文字の
上限があり、これを超えるとPR作成APIが422で失敗する。Fableの敵対的レビューで、この
上限が一切考慮されていないと指摘された。

## Decision

`prtext.Build`の返り値の`Body`を、常に65,536バイト以下になるよう切り詰める
(`maxBodyBytes`定数)。切り詰める際はUTF-8のルーン境界で切り、末尾に
「truncated」であることを示す注記を付与する。

## Rules

- 切り詰め処理は`Build`の出口(全戦略共通)で一度だけ行い、`buildSingle`/
  `buildGrouped`それぞれに個別実装しない。
- 切り詰めた場合は必ずその旨が本文から読み取れるようにする(サイレントに
  情報を落とさない)。

## Consequences

- 変更履歴が非常に長いツールでは、release notesの後半が失われる。PR作成自体は
  失敗しなくなる。
- 上限は定数化しているため、GitHub側の仕様変更があれば1箇所の変更で追従できる。

## References

- `internal/prtext/prtext.go`(`truncateBody`, `maxBodyBytes`)
