# バージョンとCLI

## 製品名

- CLI-001: GUI上の製品名は `HatoNote` とする。
- CLI-002: 実行ファイル名は `HatoNote.exe` を基本とする。
- CLI-003: 新規の内部パッケージ名、設定ディレクトリ名、ビルド成果物名に `mdz-gui` を使用しない。

## バージョン

- VERSION-001: 正式なバージョンが決まるまでは `v0.0.1-<コミット時刻>-<7桁コミットハッシュ>` 形式を使用する。
- VERSION-002: コミット時刻は `YYYY-MM-DD-HH:MM` 形式を使用する。
- VERSION-003: 開発ビルドでも取得可能な場合は実際のコミット情報を表示する。
- VERSION-004: 取得可能なコミット情報がある状態で `dev` または `unknown` だけをバージョンとして表示しない。

## CLI

- CLI-010: `HatoNote.exe --version` はバージョンを標準出力へ表示して終了する。
- CLI-011: `HatoNote.exe --help` は利用方法を標準出力へ表示して終了する。
- CLI-012: `--version` または `--help` の処理だけでGUIを起動しない。
- CLI-013: `HatoNote.exe <file.md>` で単一Markdownを開ける。
- CLI-014: `HatoNote.exe <file.mdz>` でMDZを開ける。
