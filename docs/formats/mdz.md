# 通常MDZ形式

## コンテナ

- FORMAT-MDZ-001: `.mdz` はZIP互換コンテナとして扱う。
- FORMAT-MDZ-002: Markdownページと画像等のアセットを同一コンテナへ格納できる。
- FORMAT-MDZ-003: 文書内パスは相対パスとして扱う。
- FORMAT-MDZ-004: コンテナ外へ抜けるパスを有効な文書内パスとして扱わない。

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


## 互換性

- FORMAT-MDZ-040: 旧mdz-guiが保存した `x-mdz-gui-pageOrder` をページ順情報として読み取れる。
- FORMAT-MDZ-041: HatoNoteが新しく保存するページ順情報には `x-hatonote-pageOrder` を使用する。
- FORMAT-MDZ-042: 旧ページ順キーを読み込んだ文書を保存する場合はHatoNoteのページ順キーへ移行できる。

- FORMAT-MDZ-043: HatoNote形式へ更新する場合は旧キーを新規生成しない。
