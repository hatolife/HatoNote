# mdBook形式

## 判別

- FORMAT-MDBOOK-001: `book.toml` を持つMDZをmdBook形式として識別できる。
- FORMAT-MDBOOK-002: `SUMMARY.md` を目次データとして扱う。

## 構成

- FORMAT-MDBOOK-010: `book.toml` を文書設定として保持する。
- FORMAT-MDBOOK-011: `SUMMARY.md` を目次構造として保持する。
- FORMAT-MDBOOK-012: 本文Markdownページを複数保持できる。
- FORMAT-MDBOOK-013: 画像等のアセットを文書内へ保持できる。

## プレビュー

- FORMAT-MDBOOK-020: 通常ページは利用可能な場合に実mdBookを使ってプレビューできる。
- FORMAT-MDBOOK-021: `SUMMARY.md` は通常本文ページとしてmdBookプレビューしない。
- FORMAT-MDBOOK-022: `SUMMARY.md` は目次相当の専用Markdownプレビューを利用する。
- FORMAT-MDBOOK-023: mdBookが利用できない場合でもMarkdown本文を閲覧できる。

## 新規文書

- FORMAT-MDBOOK-030: 新規mdBookの初期タイトルは `新規mdBook` とする。
- FORMAT-MDBOOK-031: 新規mdBookのdescriptionは `新規mdBookテンプレート` とする。
- FORMAT-MDBOOK-032: 新規mdBookには編集方法を確認できるテンプレートページを含める。
- FORMAT-MDBOOK-033: `SUMMARY.md` の説明と目次編集方法は専用ページとして分離できる。
