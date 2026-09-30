# 単一Markdown形式

## 対象

- FORMAT-MARKDOWN-001: 拡張子 `.md` を単一Markdownとして開く。
- FORMAT-MARKDOWN-002: 拡張子 `.markdown` を単一Markdownとして開く。
- FORMAT-MARKDOWN-003: 単一Markdownを開くためにMDZへ暗黙変換しない。

## 保存

- FORMAT-MARKDOWN-010: 通常保存では元ファイル形式を維持する。
- FORMAT-MARKDOWN-011: 文字コードは既存実装が安全に扱える範囲でUTF-8を基準とする。
- FORMAT-MARKDOWN-012: MDZとして保存する場合だけMDZコンテナを新規作成する。

## リソース

- FORMAT-MARKDOWN-020: 相対画像はMarkdownファイルの親ディレクトリを基準にする。
- FORMAT-MARKDOWN-021: 単一Markdown外部の画像をプレビューするために元画像をMDZへコピーしない。

## 安全なHTML書式

- FORMAT-MARKDOWN-030: Markdown内の生HTMLは任意に実行せず、HatoNoteが許可した安全なタグと属性だけを描画する。
- FORMAT-MARKDOWN-031: 折りたたみ用途として `details` と `summary` を許可する。
- FORMAT-MARKDOWN-032: 部分書式用途として `span` を許可する。
- FORMAT-MARKDOWN-033: `span` のstyleは文字色と文字サイズだけを許可する。
- FORMAT-MARKDOWN-034: 文字色は安全な色値だけを許可する。
- FORMAT-MARKDOWN-035: 文字サイズはHatoNoteが定めた範囲のpx指定だけを許可する。
- FORMAT-MARKDOWN-036: `script`、イベント属性、危険なURL、未許可CSSは描画しない。
