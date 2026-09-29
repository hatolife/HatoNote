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
