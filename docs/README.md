# HatoNote 仕様書

## 目的

- SPEC-001: `docs/` はHatoNoteの機能仕様を定義する。
- SPEC-002: 実装は `docs/` に記載された条件を満たす。
- SPEC-003: 機能追加または仕様変更は、関連仕様を更新してから実装する。
- SPEC-004: 既存仕様に反する不具合の修正では、正しい既存仕様を変更しない。
- SPEC-005: 仕様に記載がない動作を追加する場合は、先に仕様を追加する。

## 構成

- SPEC-010: `principles.md` は製品全体の設計原則を定義する。
- SPEC-011: `terminology.md` は用語を定義する。
- SPEC-012: `document-types.md` は文書種別と能力を定義する。
- SPEC-013: `view-states.md` は表示状態を定義する。
- SPEC-014: `editors.md` は編集エンジンを定義する。
- SPEC-015: `screens/` はユーザーが認識する画面またはオーバーレイごとの条件を定義する。
- SPEC-016: `features/` は複数画面にまたがる機能の条件を定義する。
- SPEC-017: `formats/` はファイル形式ごとの条件を定義する。

## 仕様の書式

- SPEC-020: 仕様は箇条書きの条件として記述する。
- SPEC-021: 1項目には原則として1つの条件だけを記述する。
- SPEC-022: 条件には一意な仕様IDを付ける。
- SPEC-023: 既存の仕様IDは原則として変更または振り直しを行わない。
- SPEC-024: 合否を判定できない曖昧な表現を避ける。
- SPEC-025: 見た目を飾る目的の太字、斜体、絵文字、引用を使用しない。
- SPEC-026: 見出し、仕様ID付き箇条書き、必要なコードブロック、インラインコードを使用してよい。
- SPEC-028: 仕様条件を表または通常の説明段落として記述しない。
- SPEC-027: 実装方法より、ユーザーまたは外部から観測できる条件を優先する。

## 機能仕様索引

- SPEC-030: `features/search.md` は文書内検索を定義する。
- SPEC-031: `features/recent-documents.md` は最近開いた文書を定義する。
- SPEC-032: `features/commands.md` はCommand体系とCommand Paletteを定義する。
- SPEC-033: `features/macros.md` はCommandマクロを定義する。
- SPEC-034: `features/templates.md` はページテンプレートと文書テンプレートを定義する。
- SPEC-035: `features/history.md` はバックアップ履歴、名前付き世代、差分比較を定義する。
- SPEC-036: `features/export.md` は単一HTMLとPDFエクスポートを定義する。
- SPEC-037: `features/markdown-quality.md` はMarkdown FormatterとLintを定義する。
- SPEC-038: `features/diagrams.md` はMermaid、PlantUML、数式表示を定義する。
- SPEC-039: `features/references.md` は引用元と参考文献管理を定義する。
- SPEC-040: `features/diagnostics.md` は診断表示を定義する。
- SPEC-041: `features/document-diagnostics.md` は開いている文書全体の診断を定義する。
