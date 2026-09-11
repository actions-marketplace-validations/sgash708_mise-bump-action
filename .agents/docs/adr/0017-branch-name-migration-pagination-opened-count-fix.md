# ADR 0017: ブランチ名の無告知破壊的変更・ページネーション欠如・opened-countの二重計上を修正する

## Status

Accepted

## Date

2026-09-12

## Context

Fableによる3回目の敵対的レビューで、ADR 0016(2回目レビューへの対応)自体が
新たに持ち込んだ問題と、以前から未対応だった問題が指摘された。

1. **ブランチ命名の無告知破壊的変更**: v1.0.0〜v1.4.0は
   `mise-bump/<name>-<version>`、v1.5.0(ADR 0016)は`mise-bump/<name>_<version>`
   と、区切り文字を変えるだけの破壊的変更を移行手順なしに行った。結果、
   v1.5.0へ上げた直後に同じツール・同じバージョンの組み合わせに再度遭遇すると、
   新しいブランチ名で`findClosedUnmergedPR`/`findOpenPR`を検索するため、
   旧ブランチ名で管理されていた「closeされた記憶」(ADR 0010)や「既に開いている
   PR」を認識できず、closeされたはずのPRが復活したり、同じ内容のPRが重複して
   開かれたりする。さらに`branchPrefix`も新形式のみを見るため、旧形式の
   stale PRはsuperseded closeの対象にならず、`max-open-prs`の枠を消費し
   続ける。
2. **`sanitize`の非可逆性による衝突**: `sanitize`は`:`・`/`・`-`をすべて`-`に
   変換するため、`go:github.com/foo/bar`と`go:github.com/foo-bar`のように
   異なるツール名が同一のsanitize結果(`go-github.com-foo-bar`)に潰れうる。
   ADR 0016の区切り文字変更は「同じツール名の前方一致衝突」は解決したが、
   この「異なるツール名がsanitize結果として完全に一致してしまう」根本原因は
   未解決のままだった。
3. **`opened-count`/`pr-numbers`の二重計上**: `OpenBumpPR`が既存のopen PRを
   見つけた場合(何も新規作成しない)も、新規作成した場合と同じ
   `(number, nil)`を返していたため、`bumpGroup`/`Run`はどちらも「成功して
   PRを開いた」と区別できず、`opened-count`/`pr-numbers`に含めていた。
   結果、既にopenなPRを毎回検出するだけの実行でも「今回開いた」という
   通知が繰り返し飛ぶ。
4. **open PR一覧のページネーション欠如**: `listOpenPRRefs`が`per_page=100`の
   1ページ目しか取得しておらず、`CountOpenBumpPRs`・`HasOpenPRWithPrefix`・
   `closeSupersededPRs`の3機能すべてがこれに依存しているため、baseへの
   open PRが100件を超えるリポジトリでは、101件目以降のカウント漏れ・
   superseded close漏れが起きる。

## Decision

### 1. ブランチ名にツール名のfingerprintを混ぜ、恒久的に衝突しない形式にする

`branchName`/`branchPrefix`に、ツール名のFNV-32aハッシュを8桁16進数
(`nameFingerprint`)として埋め込む: `mise-bump/<sanitize(name)>-<fingerprint>_<sanitize(version)>`。
sanitizeの結果が衝突しても、fingerprintが異なる限り最終的なブランチ名は
衝突しない(現実的なツール数に対しては十分な確率で成立する、下のbatchハッシュ
と同じ信頼モデル)。

### 2. 過去のブランチ命名スキームを`LegacyBranchNames`として引き継ぐ

`BumpPRInput`に`LegacyBranchNames []string`を追加し、`legacyBranchNames`
関数がv1.0.0〜v1.4.0形式・v1.5.0形式のブランチ名を計算して渡す。
`OpenBumpPR`は`BranchName`に加えて`LegacyBranchNames`のすべてについても
open/closed-unmergedを検索し、いずれかで見つかればそれを既存の状態として扱う。
これにより、ブランチ命名スキームが変わっても「closeされた記憶」
(ADR 0010)と「既にopenなPR」の両方が引き継がれる。

一方、superseded close(`closeSupersededPRs`)は`BranchPrefix`(現行スキーム
のみ)に基づく前方一致のままとし、旧スキームの前方一致は追加しない。
旧スキームの区切り文字(`-`)は本質的に曖昧(ADR 0016参照)であり、
prefix一致を過去のスキームにまで拡張すると無関係なツールを誤ってclose
するリスクを再導入してしまうため。**移行時は、v1.5.0以前に開かれた
bump PRを手動で一度closeすることを推奨する**(README「Upgrading」参照)。
新規に作るPRはすべて現行スキームを使うため、この手動対応は一度きりで済む。

### 3. `OpenBumpPR`が「新規作成したか」を返すようにする

`OpenBumpPR`の戻り値を`(prNumber int, created bool, err error)`に変更する。
既存のopen PRを見つけた場合は`created=false`。`bumpGroup`は`created=false`
の場合をskip扱いにし、`Run`が返す`prNumbers`(→`opened-count`/`pr-numbers`)
に含めない。

### 4. `listOpenPRRefs`をページネーション対応にする

`page`パラメータを1から増やしながら、返ってきた件数が`per_page`(100)
未満になるまで取得を続ける。無限ループ防止のため`maxOpenPRPages`(1000、
=最大10万件)で打ち切る。

### 5. `mise outdated`失敗時もoutputsを書き込む

`cmd/mise-bump-action/main.go`の`run`が、`lookup`(`mise outdated`)自体が
失敗した早期returnパスでも`writeOutputs(output, nil)`を呼ぶよう修正する。
ADR 0015の「成功/失敗/dry-runの全パスで一貫して値を設定する」という
ルールに、このパスだけ従っていなかった。

## Rules

- ブランチ命名スキームを変更する場合は、必ず`legacyBranchNames`に旧形式を
  追加し、idempotency(open/closed-unmerged検索)だけは旧スキームでも
  引き継ぐ。superseded closeまで旧スキームに広げない(prefix一致の
  曖昧性を再導入しないため)。
- ツール名を含む識別子(ブランチ名等)を生成する際、sanitizeのような
  文字置換ベースの正規化だけに頼らず、元の文字列のハッシュ
  (fingerprint)を混ぜて衝突耐性を持たせる。
- GitHub REST APIの一覧系エンドポイントを叩く共通ヘルパーは、
  最初から複数ページを前提にしたページネーションを実装する。
- action outputは、`writeOutputs`のような一貫した書き込みポイントを
  経由させ、早期returnを追加するたびに書き込み漏れがないか確認する。

## Consequences

- `go:github.com/foo/bar`と`go:github.com/foo-bar`のように、sanitizeの
  結果が一致してしまう異なるツール名でも、ブランチ名は衝突しなくなった。
- v1.5.0以前に開かれたPRも、closeされた記憶・既にopenである状態の両方が
  現行バージョンに引き継がれるようになった(superseded closeだけは
  引き継がれないため、移行時に手動closeが必要な場合がある)。
- `opened-count`/`pr-numbers`は「今回新規に開いたPR」だけを反映するように
  なり、既存PRの再検出だけの実行で同じ通知が繰り返されなくなった。
- open PRが100件を超えるリポジトリでも、`max-open-prs`のカウントと
  superseded closeが正しく機能するようになった。
- `mise outdated`自体が失敗した場合でも`opened-count=0`/`pr-numbers=`が
  設定されるようになり、後続ステップが`if: always()`で参照しても
  未設定値ではなく`"0"`を受け取れる。
- `internal/runner/mocks.go`は`GitHub`インターフェースの変更
  (`OpenBumpPR`のシグネチャ変更)に伴い再生成した(ADR 0016と同じ手順、
  go.modの`go`ディレクティブを一時的に`1.26.0`へ下げて実行)。

## References

- `internal/runner/runner.go`(`nameFingerprint`, `legacyBranchNames`, `BumpPRInput.LegacyBranchNames`, `bumpGroup`)
- `internal/githubapi/client.go`(`OpenBumpPR`, `listOpenPRRefs`, `maxOpenPRPages`)
- `cmd/mise-bump-action/main.go`(`run`)
- README「Upgrading」節
