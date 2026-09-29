# 文書種別

## 共通

- DOCTYPE-001: 文書種別は `markdown`、`mdz`、`mdbook`、`slides` の4種類を定義する。
- DOCTYPE-002: 文書種別は表示状態とは独立して管理する。
- DOCTYPE-003: 文書種別は編集エンジンとは独立して管理する。
- DOCTYPE-004: 文書種別の追加時に、既存4種類を前提とした条件分岐をアプリ全体へ追加しない。
- DOCTYPE-005: 文書種別ごとに利用可能な能力を定義する。

## markdown

- DOC-MARKDOWN-001: `markdown` は単一Markdownファイルを直接開く。
- DOC-MARKDOWN-002: `markdown` は通常保存で元のMarkdownファイルへ保存する。
- DOC-MARKDOWN-003: `markdown` は複数ページを持たない。
- DOC-MARKDOWN-004: `markdown` は外部の相対画像を元Markdownファイルの位置を基準に解決する。
- DOC-MARKDOWN-005: `markdown` は見出しからTOCを生成できる。
- DOC-MARKDOWN-006: `markdown` は内蔵エディターを利用できる。
- DOC-MARKDOWN-007: `markdown` はNeovimを利用できる。
- DOC-MARKDOWN-008: `markdown` はMDZとして保存する変換操作を提供できる。

## mdz

- DOC-MDZ-001: `mdz` は複数のMarkdownページを持てる。
- DOC-MDZ-002: `mdz` は画像等のアセットを文書内へ格納できる。
- DOC-MDZ-003: `mdz` はページ順序を保持する。
- DOC-MDZ-004: `mdz` は見出しからページ内TOCを生成できる。
- DOC-MDZ-005: `mdz` は内蔵エディターを利用できる。
- DOC-MDZ-006: `mdz` はNeovimを利用できる。
- DOC-MDZ-007: `mdz` はスライドへ変換できる。

## mdbook

- DOC-MDBOOK-001: `mdbook` は複数のMarkdownページを持てる。
- DOC-MDBOOK-002: `mdbook` は `SUMMARY.md` を目次構造の正規データとして扱う。
- DOC-MDBOOK-003: `mdbook` は `book.toml` を文書設定として扱う。
- DOC-MDBOOK-004: `mdbook` は実際のmdBookを利用したプレビューを提供できる。
- DOC-MDBOOK-005: `mdbook` は内蔵エディターを利用できる。
- DOC-MDBOOK-006: `mdbook` はNeovimを利用できる。
- DOC-MDBOOK-007: `SUMMARY.md` の編集時は専用プレビューを使用できる。
- DOC-MDBOOK-008: mdBook実行ファイルが利用できない場合でも文書内容へアクセスできる。

## slides

- DOC-SLIDES-001: `slides` は複数のスライドページを持てる。
- DOC-SLIDES-002: `slides` は各ページをMarkdownで編集する。
- DOC-SLIDES-003: `slides` は画像等のアセットを文書内へ格納できる。
- DOC-SLIDES-004: `slides` はスライド順序を保持する。
- DOC-SLIDES-005: `slides` は内蔵エディターを利用できる。
- DOC-SLIDES-006: `slides` はNeovimを利用できる。
- DOC-SLIDES-007: `slides` は全画面プレゼンテーションを利用できる。
- DOC-SLIDES-008: `slides` は2画面プレゼンテーションを利用できる。
- DOC-SLIDES-009: `slides` は通常MDZへ変換できる。


## 能力

- DOCTYPE-010: 文書種別から複数ページを持てるか判定できる。
- DOCTYPE-011: 文書種別から画像等のアセットを文書内へ格納できるか判定できる。
- DOCTYPE-012: 文書種別から見出しTOCを利用できるか判定できる。
- DOCTYPE-013: 文書種別からmdBookプレビューを利用できるか判定できる。
- DOCTYPE-014: 文書種別からプレゼンテーションを利用できるか判定できる。
- DOCTYPE-015: 文書種別から外部編集エンジンを利用できるか判定できる。
- DOCTYPE-016: UIは文書種別ごとの個別条件を増やす前に、既存の能力で判定できないか確認する。
