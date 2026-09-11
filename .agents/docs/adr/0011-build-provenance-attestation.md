# ADR 0011: バイナリの真正性はGitHub Artifact Attestation(Sigstore)で検証する

## Status

Accepted

## Date

2026-09-12

## Context

これまでのバイナリ検証は、同じリリースに同梱された`checksums.txt`を`sha256sum -c`で
照合するだけだった。Fableの敵対的レビューで指摘された通り、これは**自己参照**である:
リリースアセット(バイナリと`checksums.txt`)を書き換えられる権限を持つ攻撃者は、
両方を同時に差し替えられるため、この検証は転送中の破損は検知できても、リリース自体が
改竄された場合には無力である。

action.yml(gitのタグで保護される側)にSHA256を直接埋め込む案も検討したが、タグ作成時点
ではまだビルドが完了しておらず正しいハッシュを知り得ない(ビルド→ハッシュ算出→タグ、の
順序を守れない)ため、事後にタグを付け替える必要が生じ、実質的に同じ自己参照問題を
先送りするだけだった。

## Decision

`actions/attest-build-provenance`を使い、GitHubのOIDC発行者とSigstoreの透明性ログ
(rekor)に基づく来歴証明(provenance attestation)をビルド時に生成する。この証明は
「どのworkflow run・どのcommitがこのバイナリをビルドしたか」を暗号学的に記録し、
リリースアセット自体とは独立した検証経路(GitHubのAttestations API + Sigstoreの
公開ログ)を持つため、リリースアセットを差し替えるだけでは偽装できない。

- `release.yml`: `id-token: write` / `attestations: write` 権限を付与し、両アーキ
  テクチャのバイナリに対して`actions/attest-build-provenance`を実行する。
- `action.yml`: バイナリダウンロード後、`sha256sum -c`(転送中の破損チェック)に加えて
  `gh attestation verify <binary> --repo sgash708/mise-bump-action`を実行し、
  Sigstore側の証明を検証する。

## Rules

- `sha256sum -c`とattestation検証は両方とも実行する(前者は安価な破損検知、後者が
  実質的な真正性の担保)。どちらか一方を省略しない。
- 新しいビルド成果物(将来アーキテクチャを追加する場合等)は必ず`subject-path`に含める。

## Consequences

- `gh attestation verify`はネットワーク経由でSigstoreの透明性ログを参照するため、
  完全にオフラインでは動作しない(GitHub Actions環境では問題にならない)。
- 攻撃者がリポジトリへの書き込み権限とGitHub Actionsのworkflow実行権限の両方を
  掌握しない限り、正規の来歴証明を偽造できない。書き込み権限だけでリリースアセットを
  差し替えるケースには対応できる。

## References

- `.github/workflows/release.yml`
- `action.yml`
