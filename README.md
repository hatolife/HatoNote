# HatoNote

Markdown文書を閲覧、編集、管理するWindowsアプリです。

## 対応文書

- 単一Markdown: `.md` と `.markdown` をそのまま閲覧、編集する。
- 通常MDZ: Markdown、画像、複数ページを1つの `.mdz` にまとめる。
- mdBook: `book.toml` と `SUMMARY.md` を持つMDZを本として編集、プレビューする。
- スライド: Markdownでスライドを作成、編集し、1画面または2画面で発表する。

## 主な機能

- Markdown内蔵エディター、WYSIWYG内蔵エディター、Neovimを切り替えて編集できる。WYSIWYGは単一Markdown、通常MDZ、スライドで利用できる。
- Markdownプレビューとエディターの表示位置を同期できる。
- 画像を文書へ貼り付けてMDZ内へ保存できる。
- 通常MDZとスライドを相互変換できる。
- mdBookの目次をGUIと `SUMMARY.md` の両方から編集できる。
- スライドの文字サイズ、余白、レイアウトを編集できる。
- 未保存変更の確認、自動保存、バックアップ、作業データ復旧を利用できる。

## MDZ

- MDZはMarkdownと画像等をZIPコンテナへまとめる文書形式として扱う。
- 通常MDZ、mdBook形式、スライド形式はいずれも拡張子 `.mdz` を使用する。
- MDZip仕様は https://github.com/mdzip-project/mdzip-spec を参照する。

## 実行

- GitHub Releasesから `HatoNote-windows-amd64-日時.zip` を取得する。
- ZIPを展開する。
- `HatoNote.exe` を起動する。
- インストーラーは不要とする。
- Microsoft Edge WebView2 Runtimeを必要とする。
- Neovimは任意とする。
- mdBookは任意とし、未導入時は設定または案内からwingetで導入できる。

## コマンドライン

```sh
HatoNote.exe document.md
HatoNote.exe document.mdz
HatoNote.exe --help
HatoNote.exe --version
```

## データ

- 設定、作業データ、バックアップはWindowsのユーザーキャッシュ領域下の `HatoNote` ディレクトリへ保存する。
- スライドの発表者ノートはMDZ内へ保存するため、MDZを共有するとノートも共有される。

## ヘルプ

- Releaseに `HatoNote-help.mdz` を同梱する。
- `HatoNote-help.mdz` はHatoNoteで開いて閲覧する。

## 開発

- Go 1.26を使用する。
- Node.js 22以降を使用する。
- Wails v2を使用する。
- フロントエンドはTypeScriptを使用する。

```sh
wails build
```

## 仕様

- `docs/` を機能仕様の正とする。
- 機能追加または仕様変更では仕様書を先に更新する。
- 仕様書の記述規則は `AGENTS.md` と `docs/README.md` に従う。
