# 文書診断

## 実行

- DOCCHK-001: 開いている文書全体を明示操作で診断できる。
- DOCCHK-002: 診断前に現在の未保存編集内容を取り込む。
- DOCCHK-003: 診断は文書内容を変更しない。
- DOCCHK-004: 診断結果に種類、対象ページ、メッセージを表示する。
- DOCCHK-005: 対象ページを持つ診断結果からそのページへ移動できる。
- DOCCHK-006: 文書診断はCommand Paletteから実行できる。

## Markdown

- DOCCHK-010: 文書内の全Markdownページへ現在のLint設定を適用する。
- DOCCHK-011: Lint診断には規則IDと行番号を表示する。
- DOCCHK-012: 単一Markdownでは現在の1ファイルを診断する。

## リンク

- DOCCHK-020: Markdown内の相対リンク先が存在しない場合は切れたリンクとして報告する。
- DOCCHK-021: Markdown内の相対画像が存在しない場合は切れた画像として報告する。
- DOCCHK-022: HTTP、HTTPS、mailto等の外部URLは存在確認を行わない。
- DOCCHK-023: 同一ページ内のアンカーリンクはファイル存在確認の対象外とする。
- DOCCHK-024: MDZ内リンクは文書内ファイルの存在を確認する。
- DOCCHK-025: 単一Markdownの相対リンクと相対画像は元Markdownファイルの隣を基準に存在確認する。

## 未使用画像

- DOCCHK-030: MDZ内画像のうちどのMarkdownページからも参照されていない画像を報告する。
- DOCCHK-031: 未使用画像診断は単一Markdownには適用しない。
- DOCCHK-032: 未使用判定はMarkdownの画像参照だけを対象とする。
