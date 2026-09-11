# ADR 0013: 未対応の値形式(inline table/array)は個別エントリだけスキップし、グループ全体を失敗させない

## Status

Accepted

## Date

2026-09-12

## Context

`internal/misetoml.Bump`は`python = { version = "3.11" }`のようなinline table形式
(有効なmise.tomlだが単純な文字列置換では書き換えられない)を検出するとエラーを返す。
これまでの実装ではこのエラーがそのままグループ全体・実行全体の失敗として扱われ、
同じ`mise.toml`内の他の正常にbump可能なツールも巻き込んで処理が止まり、GitHub Actions
のジョブ全体が失敗表示になっていた。Fableの敵対的レビューで、これが「除外する手段が
無いまま実行全体を壊す」致命的な問題として指摘された。

## Decision

`internal/misetoml`に`ErrUnsupportedValueForm`というsentinelエラーを導入する。
`internal/runner.bumpGroup`は、グループ内の各エントリを個別に`Bump`し、
`errors.Is(err, misetoml.ErrUnsupportedValueForm)`の場合はそのエントリだけを
スキップ(job summaryに理由を記録)して残りのエントリの処理を続行する。グループ内の
全エントリがスキップされた場合のみ、そのグループ全体をスキップとして扱う
(PRを開かないが失敗にもしない)。

`ignore` input(ADR 0012)と組み合わせることで、恒久的にスキップしたい場合は
明示的に除外もできる。

## Rules

- `misetoml.Bump`が返すエラーのうち、既知の「対応できない値形式」は
  `ErrUnsupportedValueForm`でラップし、呼び出し元が判別できるようにする。
- `bumpGroup`は個別エントリのスキップとグループ全体の失敗を区別し、前者を理由に
  実行全体を失敗として終了させない。

## Consequences

- inline table/array形式でピンされたツールが1つあっても、同じファイル内の他の
  ツールは正常にbumpされる。
- スキップされたツールは`ignore`で明示的に除外しない限り、毎回同じ警告が
  job summaryに出続ける(実害はないが、通知が煩わしい場合は`ignore`の利用を促す)。

## References

- `internal/misetoml/misetoml.go`(`ErrUnsupportedValueForm`)
- `internal/runner/runner.go`(`bumpGroup`の per-entry ループ)
