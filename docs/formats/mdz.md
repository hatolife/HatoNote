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
