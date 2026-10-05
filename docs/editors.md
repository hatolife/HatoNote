# 編集エンジン

## 共通

- EDITOR-001: 編集エンジンは `builtin`、`wysiwyg`、`neovim` を定義する。
- EDITOR-002: 編集エンジンは表示状態とは独立して管理する。
- EDITOR-003: 編集中に、現在の文書で利用可能な編集エンジンを切り替えられる。
- EDITOR-004: 編集エンジンを切り替えても現在の編集内容を失わない。
- EDITOR-005: 設定画面の編集エンジン選択は編集開始時の初期値とする。
- EDITOR-006: 現在利用中の編集エンジンをUIから判別できる。

## builtin

- EDITOR-BUILTIN-001: `builtin` はHatoNote内でMarkdown本文を編集する。
- EDITOR-BUILTIN-002: `builtin` はUndoとRedoを利用できる。
- EDITOR-BUILTIN-003: Undoできない場合はUndo操作を無効表示する。
- EDITOR-BUILTIN-004: Redoできない場合はRedo操作を無効表示する。
- EDITOR-BUILTIN-005: `builtin` のスクロール位置をプレビューと同期できる。

## wysiwyg（リッチ編集）

- EDITOR-WYSIWYG-001: `wysiwyg` はMarkdownの表示結果に近い見た目のまま本文を編集できる。
- EDITOR-WYSIWYG-002: `wysiwyg` の保存データはMarkdown本文とし、WYSIWYG専用の文書形式へ変換して保存しない。
- EDITOR-WYSIWYG-003: `wysiwyg` は見出し、太字、斜体、取り消し線、リンク、箇条書き、番号付きリスト、引用、コード、水平線、画像を編集できる。
- EDITOR-WYSIWYG-004: `wysiwyg` はUndoとRedoを利用できる。
- EDITOR-WYSIWYG-005: `wysiwyg` で安全に往復変換できないMarkdownを検出した場合は、元のMarkdown本文を変更せず `builtin` で編集できる状態を維持する。
- EDITOR-WYSIWYG-006: `wysiwyg` のスクロール位置をプレビューと同期できる。
- EDITOR-WYSIWYG-007: `wysiwyg` の画像追加は文書種別の画像格納能力に従う。
- EDITOR-WYSIWYG-008: スライド文書の `wysiwyg` は現在スライドのレイアウト、余白、フォント、文字サイズ、背景色を編集表示へ反映する。
- EDITOR-WYSIWYG-009: スライド文書の `wysiwyg` は自由配置編集を提供せず、Markdown本文を正規データとして維持する。
- EDITOR-WYSIWYG-010: `columns` レイアウトのWYSIWYG編集では列区切り `<!-- column -->` を失わない。
- EDITOR-WYSIWYG-011: リッチ編集中は編集面の上部に書式バーを表示する。
- EDITOR-WYSIWYG-012: 書式バーから段落・見出し、太字、斜体、取り消し線、コード、箇条書き、番号付きリスト、リンク、水平線を変更できる。
- EDITOR-WYSIWYG-013: スライド文書では1スライドにつき1つのMarkdownファイルだけをWYSIWYG編集対象とする。
- EDITOR-WYSIWYG-014: スライドWYSIWYGは1つのMarkdownファイル内に複数スライドを保持する機能を提供しない。
- EDITOR-WYSIWYG-015: スライドWYSIWYGでは `---` を現在スライド内の水平線として扱い、ページ区切りとして解釈しない。
- EDITOR-WYSIWYG-016: スライドのページ分割操作は分割後の各ページを別のMarkdownファイルとして作成する。
- EDITOR-WYSIWYG-017: UI上の編集エンジン名は `WYSIWYG` ではなく「リッチ編集」と表示する。
- EDITOR-WYSIWYG-018: 書式バーから3列3行の表を挿入できる。
- EDITOR-WYSIWYG-019: 表内では現在セルを基準に行を上または下へ追加でき、現在行を削除できる。
- EDITOR-WYSIWYG-020: 表内では現在セルを基準に列を左または右へ追加でき、現在列を削除できる。
- EDITOR-WYSIWYG-021: 表の配置指定はMarkdown表で保持できる列単位の左寄せ、中央寄せ、右寄せとして保存する。
- EDITOR-WYSIWYG-022: 通常MDZのリッチ編集では折りたたみ要素を挿入できる。
- EDITOR-WYSIWYG-023: スライドのリッチ編集では折りたたみ要素を提供しない。
- EDITOR-WYSIWYG-024: リッチ編集では選択範囲へ文字色を指定でき、Markdownには安全な `<span style="color:...">` として保存する。
- EDITOR-WYSIWYG-025: リッチ編集では選択範囲へ文字サイズを指定でき、Markdownには安全な `<span style="font-size:...">` として保存する。
- EDITOR-WYSIWYG-026: HTML書式は許可済みタグと属性だけを描画し、スクリプト、イベント属性、危険なURL、未許可CSSを実行または保持しない。
- EDITOR-WYSIWYG-027: 安全に往復変換できないMarkdownを検出してリッチ編集への切り替えを拒否した場合は、判定原因となった構文またはDOM差分をログへ出力する。
- EDITOR-WYSIWYG-028: リッチ編集への切り替えを拒否した場合は、判定原因の要約をステータス表示へ含める。

## neovim

- EDITOR-NVIM-001: `neovim` はユーザー指定または自動検出したNeovimを使用する。
- EDITOR-NVIM-002: Windows側でNeovimを検出できない場合はWSLの既定ディストリビューションを探索できる。
- EDITOR-NVIM-003: WSLでは既定ユーザーのPATHを探索対象とする。
- EDITOR-NVIM-004: Neovim実行ファイルを設定画面から明示指定できる。
- EDITOR-NVIM-005: `init.lua` を設定画面から明示指定できる。
- EDITOR-NVIM-006: `init.lua` を指定しないフォールバックとして `--clean` を利用できる。
- EDITOR-NVIM-007: 指定したNeovim実行ファイルの存在状態を設定画面に表示する。
- EDITOR-NVIM-008: 指定した `init.lua` の存在状態を設定画面に表示する。
- EDITOR-NVIM-009: Neovimの表示位置をプレビューと同期できる。
- EDITOR-NVIM-010: Neovimの設定読み込みエラーによる入力待ち状態が発生してもHatoNote全体を操作不能にしない。
- EDITOR-NVIM-011: Neovimが設定エラーまたはプロセス異常で継続利用できない場合は、作業ファイルへ保存できる内容を保持して内蔵エディターへ切り替える。
- EDITOR-NVIM-012: Neovimから内蔵エディターへ自動切替した場合は、原因をユーザーへ表示する。
