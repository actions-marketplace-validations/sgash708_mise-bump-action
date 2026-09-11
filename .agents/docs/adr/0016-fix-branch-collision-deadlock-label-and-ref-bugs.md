# ADR 0016: ブランチプレフィックス衝突・max-open-prsデッドロック・ラベル衝突・`/head`誤検知の4件を修正する

## Status

Accepted

## Date

2026-09-12

## Context

Fableによる2回目の敵対的レビューで、直前の急ぎの修正サイクル(ADR 0012〜0015)が
新たに持ち込んだ4件の重大な不具合が指摘された。

1. **ブランチ名プレフィックス衝突**: `branchPrefix`が`"mise-bump/go-"`のような
   `-`区切りを使っていたため、`go`(ツール名`"go"`)と
   `go:github.com/matryer/moq`(ツール名をsanitizeすると`go-github.com-matryer-moq`)が
   どちらも`"go-"`で始まってしまい、`strings.HasPrefix`による前方一致が
   ツール名の境界を越えて誤爆する。`go`をbumpした際に無関係な`moq`のPRが
   supersededとして誤ってcloseされ得る。
2. **max-open-prsの永久デッドロック**: supersededなPRのcloseは`OpenBumpPR`内で
   新PR作成に成功した後にしか実行されない。一方`max-open-prs`の上限判定は
   `OpenBumpPR`(=`bumpGroup`)の呼び出し自体をスキップする形で効いていた。
   古いPRがcap分の枠を占有している状況では、そのPRを置き換えるはずの新しい
   bumpが「capに達している」という理由でブロックされ続け、古いPRを一生closeできない。
3. **`CountOpenBumpPRs`とDependabotのラベル衝突**: `max-open-prs`のカウントは
   `cfg.Labels`(既定`["dependencies"]`)に一致するPRの数で数えていた。
   Dependabotも既定で同じ`"dependencies"`ラベルを使うため、両方を運用する
   リポジトリではDependabot自身のPRまで数に含まれ、エラーも出さずに
   mise-bump-actionが新しいPRを開けなくなる。
4. **`/head`部分一致による誤検知**: `internal/config/config.go`の
   pull_requestイベント検知が`strings.Contains(baseBranch, "/head")`だったため、
   `"feature/header-fix"`のような正当なブランチ名まで
   「pull_requestイベントのrefらしい」と誤判定してエラーになる。

## Decision

### 1. ブランチ名の区切り文字を`_`にする

`sanitize()`は`[A-Za-z0-9.-]`以外の文字をすべて`-`に変換するため`_`を
決して生成しない。ツール名とバージョンの区切りを`-`から`_`に変えることで、
`branchPrefix`の`HasPrefix`判定がツール名の境界を安全に越えなくなる
(`branchName`/`branchPrefix`、`internal/runner/runner.go`)。

### 2. supersededのcloseをmax-open-prsのcap判定より前に評価する

`GitHub`インターフェースに`HasOpenPRWithPrefix(ctx, base, prefix) (bool, error)`を
追加する。`Run`はグループごとに、そのツール用の同じprefixを持つopen PRが
既に存在するかを`max-open-prs`が設定されている場合のみ確認し、存在する場合は
そのグループを「置き換えであり正味の増加ではない」として上限判定の対象から
除外する(`replacesExisting`)。実際のclose(コメント付き、新PR番号を含む)は
従来どおり`OpenBumpPR`成功後の`closeSupersededPRs`で行う — 変更するのは
「upper capが置き換えをブロックしない」という判定順序だけであり、
close自体のタイミングやコメント内容は変えない。

### 3. `CountOpenBumpPRs`/`HasOpenPRWithPrefix`をブランチ名prefixで判定する

`CountOpenBumpPRs`からラベル引数を削除し、`head.ref`が
`runner.BranchNamespace`(`"mise-bump/"`)で始まるPRだけを数える。
`branchName`が生成するブランチは`pr-strategy`に関わらず必ずこのprefixを
持つため、ラベルという間接的な手掛かりに頼らず「これがmise-bump-action自身が
開いたPRかどうか」を直接判定できる。

### 4. pull_requestイベントのref判定を正規表現の完全一致にする

`strings.Contains`を、`^[0-9]+/(merge|head)$`にマッチする専用の正規表現
(`pullRequestRefName`)へ置き換える。GitHubの`GITHUB_REF_NAME`は
pull_requestイベントで必ず`"<pr番号>/merge"`の形になるため、この形式に
厳密一致する場合のみエラーにする。

## Rules

- ブランチ名・prefixの生成/判定ロジックを変更する際は、`sanitize()`が
  出力に含めない文字(現在は`_`)を区切りに使う。ツール名やバージョン文字列
  自体が生成し得る文字を区切りに使わない。
- `max-open-prs`のカウント・置き換え判定は、ラベルではなく
  `runner.BranchNamespace`によるブランチ名prefix判定に統一する。
  ラベルは引き続きPR作成時に付与するが、「これが自分のPRか」の判定には使わない。
- 環境変数やユーザー入力の形式チェックは、部分一致(`Contains`/`HasSuffix`)ではなく
  実際に取り得る値の形式全体に一致する正規表現/完全一致を優先する。

## Consequences

- `go`と`go:github.com/matryer/moq`のように名前が前方一致するツール同士でも
  supersededのclose対象を誤らなくなった。
- `max-open-prs`が既に上限に達している状態でも、既存PRの置き換えとなる
  bumpは正常に処理され、古いPRが永久に残り続けることがなくなった。
- Dependabotなど他ツールと同じデフォルトラベルを共有していても
  `max-open-prs`のカウントに影響しなくなった。
- `feature/header-fix`のような、たまたま`/head`を含むだけの正当なブランチ名を
  `base-branch`として使えるようになった。
- `internal/runner/mocks.go`は`GitHub`インターフェースの変更
  (`CountOpenBumpPRs`のシグネチャ変更、`HasOpenPRWithPrefix`の追加)に伴い
  再生成した。既知の制約(moq v0.7.1とgo.mod `go 1.27`以降の非互換、
  `.agents/docs/troubleshooting.md`)により、生成時のみ`go.mod`の`go`
  ディレクティブを`1.26.0`に一時的に下げて実行し、生成後に`1.27.1`へ戻した。

## References

- `internal/runner/runner.go`(`BranchNamespace`, `branchName`, `branchPrefix`, `Run`)
- `internal/githubapi/client.go`(`CountOpenBumpPRs`, `HasOpenPRWithPrefix`, `listOpenPRRefs`)
- `internal/config/config.go`(`pullRequestRefName`)
- `.agents/docs/troubleshooting.md`(「moqがgo.mod `go 1.27`以降と組み合わさると失敗する」)
