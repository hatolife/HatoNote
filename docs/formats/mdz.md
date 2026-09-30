# 通常MDZ形式

## コンテナ

- FORMAT-MDZ-001: `.mdz` はZIP互換コンテナとして扱う。
- FORMAT-MDZ-002: Markdownページと画像等のアセットを同一コンテナへ格納できる。
- FORMAT-MDZ-003: 文書内パスは相対パスとして扱う。
- FORMAT-MDZ-004: コンテナ外へ抜けるパスを有効な文書内パスとして扱わない。
- FORMAT-MDZ-005: `manifest.mode` は独立したアプリ状態として保持せず、保存時に文書構造から決定する。
- FORMAT-MDZ-006: `slides.json` を持つMDZの `manifest.mode` は `slides` とする。
- FORMAT-MDZ-007: `book.toml` を持ちスライドではないMDZの `manifest.mode` は `project` とする。
- FORMAT-MDZ-008: それ以外のMDZの `manifest.mode` は `document` とする。

## ページ

- FORMAT-MDZ-010: 複数のMarkdownページを格納できる。
- FORMAT-MDZ-011: ページの表示順序を保持できる。
- FORMAT-MDZ-012: ページの表示順序をファイル名だけから推測することを必須としない。
- FORMAT-MDZ-013: ページ名に連番を強制しない。

## 判別

- FORMAT-MDZ-020: mdBookまたはスライドの識別情報がないMDZを通常MDZとして扱う。


## 新規文書

- FORMAT-MDZ-030: 新規通常MDZの最初のページ名は `本文.md` とする。
- FORMAT-MDZ-031: 新規通常MDZでmdBook由来の `index.md` を既定ページ名として使用しない。
- FORMAT-MDZ-032: 新規通常MDZの最初のページにHatoNoteで編集できることを示す短い案内を含める。
- FORMAT-MDZ-033: 新規通常MDZは作成直後に編集を開始できる。

## ページ順メタデータ

- FORMAT-MDZ-040: ページ順メタデータキーは `x-hatonote-pageOrder` とする。
- FORMAT-MDZ-041: ページ順を保存する場合は `x-hatonote-pageOrder` を使用する。
- FORMAT-MDZ-042: HatoNoteは旧製品固有のページ順キーを読み書きしない。
- FORMAT-MDZ-043: ページ順キーの別名を定義しない。
- FORMAT-MDZ-044: 引用元メタデータを保存する場合は `x-hatonote-references` を使用する。
- FORMAT-MDZ-045: `x-hatonote-references` は通常MDZの引用元正規データとして扱う。
