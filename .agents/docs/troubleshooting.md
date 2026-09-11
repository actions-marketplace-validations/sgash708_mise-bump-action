# トラブルシューティング

開発・運用で実際に遭遇した詰まりどころと対処法。時系列の経緯は
[`plans/2026-09-11-mise-bump-action-implementation.md`](plans/2026-09-11-mise-bump-action-implementation.md)
の各Taskの「実施後の追補」を参照。

## moqがgo.mod `go 1.27`以降と組み合わさると失敗する(解決済み: go.modをgo 1.26系に固定)

**症状:**

```
internal error: package "context" without types was imported
```

`go generate`で`mocks.go`を再生成しようとすると、moq(v0.6.0/v0.7.1いずれも)がこのエラーで落ちる。

**原因:** moqが内部で使う`golang.org/x/tools`のpackage loaderが、`go.mod`の`go`ディレクティブが
`1.27`以降の場合に対応できていない([matryer/moq#246](https://github.com/matryer/moq/issues/246)、
2026-09-09時点でopen)。修正コミット(`51bed092a2bd1eaecca4ffe2b91ab1c4eec59741`,
[#245](https://github.com/matryer/moq/pull/245))はmainにマージ済みだが、これを含む
`v0.7.2`はまだリリースされていない。

一方、Go 1.26対応自体は[#242](https://github.com/matryer/moq/pull/242)で`v0.7.1`(現在
pinしているタグ)に既に含まれており、`go.mod`の`go`ディレクティブが`1.26.x`であれば
moq v0.7.1はエラー無く動く。

**対応:** `go.mod`の`go`ディレクティブを`1.26.8`に固定し、v0.7.2がリリースされ次第
`1.27.x`へ戻すか、v0.7.2にpinし直す。一時的なgo.mod書き換えは不要になった。

以前の回避策(生成のたびに`go.mod`を`1.25.0`へ一時的に下げて戻す)は、この対応により
不要になったため廃止した。

## `GITHUB_TOKEN`が作成したPRのworkflow runが`action_required`になる

**症状:** mise-bump-actionが開いたPRに対応するCI(`ci.yml`/`lint.yml`)のworkflow runが
`action_required`ステータスのまま止まり、自動実行されない。

**原因:** GitHubは`GITHUB_TOKEN`で作成されたコンテンツ(PR等)から発火したworkflow runに対して、
たとえfork PRでなく同一リポジトリ内であっても、手動承認を要求する。ループ防止のためのプラット
フォーム側の仕様であり、mise-bump-actionのコード不具合ではない。

**対処法:** 該当runを手動で承認する。

```bash
gh api -X POST repos/{owner}/{repo}/actions/runs/{run_id}/approve
```

## 「GitHub Actions is not permitted to create or approve pull requests」で403になる

**症状:** mise-bump-actionのPR作成ステップが403エラーで失敗する。

**原因:** リポジトリの既定設定では「Allow GitHub Actions to create and approve pull requests」が
無効になっている。workflow側で`permissions: pull-requests: write`を指定するだけでは不十分。

**対処法:** リポジトリ設定でこのオプションを有効化する。

```bash
gh api -X PUT repos/{owner}/{repo}/actions/permissions/workflow \
  -f default_workflow_permissions=write \
  -F can_approve_pull_request_reviews=true
```

## `go:github.com/matryer/moq`が間欠的に検出されない([jdx/mise#2766](https://github.com/jdx/mise/issues/2766))

**症状:** `mise outdated --json --bump`を実行しても、`go install`バックエンド経由のツール
(例: `go:github.com/matryer/moq`)が`mise WARN Error getting latest version for
go:github.com/matryer/moq: no latest version found`という警告とともに結果から抜け落ちる。

**原因:** mise本体の`go install`バックエンド([jdx/mise#2766](https://github.com/jdx/mise/issues/2766))
の既知のバグで、`go list -m -versions`まわりの解決に失敗することがある。

**対処法:** [ADR 0001](adr/0001-delegate-version-resolution-to-mise-cli.md)の「バージョン解決は
mise CLIに委譲する」方針上、mise-bump-action側では回避できない既知の制限として受け入れる。
aquaバックエンド(`aqua:owner/repo`形式)は影響を受けない。

## `mise outdated`が親ディレクトリ・グローバル設定のツールまで結果に含めてしまう

**症状:** `mise-config-path`で特定の`mise.toml`を指定しているのに、そのファイルに存在しない
ツール(親ディレクトリの`mise.toml`や`~/.config/mise/config.toml`で管理されているもの)まで
outdatedの結果に混ざる。

**原因:** mise本体はカレントディレクトリから親方向へ`mise.toml`を探索し、グローバル設定も含めて
マージした結果を返す仕様になっている。`mise outdated --json`はこのマージ後の全ツールを対象にする。

**対処法:** `internal/outdated.Parse`は各エントリの`source.path`が対象の`mise-config-path`の
絶対パスと厳密に一致するものだけを採用し、それ以外(親ディレクトリ・グローバル設定由来)を
除外する。この対応は実装済みで、追加の運用対応は不要。

## release notesのコミット件名がGitHubの@メンションとして解釈される

**症状:** リリースページの「Contributors」欄に、実際にはこのリポジトリと無関係な
GitHubアカウントが表示される。

**原因:** `release.yml`はコミット件名をそのままrelease notesに埋め込む。件名に
`@v0`のような文字列が含まれていると、GitHubはそれを実在ユーザーへの@メンションと
解釈し、たまたま存在するアカウントをContributorとして表示してしまう(実際に
floating major tagの"v0"という文字列が既存のGitHubユーザー名と衝突して発生した)。

**対処法:** `release.yml`はコミット件名から`@name`パターンを抽出しバッククォートで
囲むことで、GitHubにインラインコードとして描画させメンション化を防ぐ(実装済み)。
コミットメッセージを書く際も、`@`から始まる文字列は説明目的であっても
バッククォートで囲む習慣をつけること。

## release.ymlのsedスクリプトがshellcheck SC2016でreviewdog上fail扱いになる

**症状:** `sed -E 's/@(...)/`@\1`/g'`のようなバックリファレンス(`\1`)を含む
シングルクォートのsedスクリプトに対し、shellcheckがSC2016(「シングルクォート内は
展開されない、ダブルクォートを使うべき」)を報告する。ローカルの`actionlint`は
exit code 0で通るが、reviewdog経由のGitHub Check(`actionlint`という名前)は
`fail`扱いになりPRがブロックされる。

**原因:** SC2016はshellcheck内部では`:info:`重要度だが、actionlintはshellcheckの
重要度情報を保持せず一律の問題として報告するため、reviewdogの`fail_level: error`が
これを拾ってしまう。`\1`はsedのバックリファレンスであり、シェル変数展開の意図では
ないため実害はない誤検知。

**対処法:** 対象行の直前に`# shellcheck disable=SC2016`コメントを置いて明示的に
抑制する(実装済み)。

## `422 Reference already exists`でブランチ作成が失敗する

**症状:** mise-bump-actionのPR作成が`422 Reference already exists`で失敗する。

**原因:** 決定論的ブランチ名(`mise-bump/<tool>-<version>`)の設計上、以前の実行で残った
ブランチと衝突する。

**対処法:** `internal/githubapi.OpenBumpPR`は同じブランチに対応する既存のopen PRがあればその
PR番号を返し、無ければstaleなブランチを削除してから再作成する(冪等化済み)。この対応は実装済み。
