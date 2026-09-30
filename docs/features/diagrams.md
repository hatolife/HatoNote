# Markdown図表拡張

## 共通

- DIAGRAM-001: Markdownコードフェンスを図表としてプレビューできる。
- DIAGRAM-002: 図表描画結果をMarkdown本文へ書き戻さない。
- DIAGRAM-003: 図表描画用の外部依存が利用できない場合は元のコードブロックを表示する。
- DIAGRAM-004: 図表描画失敗時は元のコードブロックを表示し、失敗理由を確認できる。
- DIAGRAM-005: 同じ図表種類、ソース、描画設定の組み合わせは再利用できる。
- DIAGRAM-006: 図表描画処理には実行時間の上限を設ける。
- DIAGRAM-007: 生成SVGはスクリプトとして実行せず画像として表示する。
- DIAGRAM-008: 外部CDNを図表描画の必須要件としない。

## Mermaid

- DIAGRAM-010: `mermaid` コードフェンスをMermaid図として描画できる。
- DIAGRAM-011: Mermaid描画には `mmdc` を利用できる。
- DIAGRAM-012: Mermaid実行ファイルのパスが空欄の場合はPATHから `mmdc` を検出する。
- DIAGRAM-013: MermaidはSVGとして生成する。
- DIAGRAM-014: Mermaid描画時はstrictセキュリティ設定を使用する。

## PlantUML

- DIAGRAM-020: `plantuml` または `puml` コードフェンスをPlantUML図として描画できる。
- DIAGRAM-021: PlantUML描画にはJavaとPlantUML JARを利用する。
- DIAGRAM-022: Java実行ファイルのパスが空欄の場合はPATHから `java` を検出する。
- DIAGRAM-023: PlantUML JARは設定から指定できる。
- DIAGRAM-024: PlantUMLはSVGとして生成する。
- DIAGRAM-025: PlantUMLは `SANDBOX` セキュリティプロファイルで実行する。
- DIAGRAM-026: `@start...` を含まない `plantuml` コードフェンスは `@startuml` と `@enduml` で補完できる。

## 表示対象

- DIAGRAM-030: 通常Markdownプレビューで図表を表示できる。
- DIAGRAM-031: 通常MDZプレビューで図表を表示できる。
- DIAGRAM-032: mdBookの通常Markdownフォールバックプレビューで図表を表示できる。
- DIAGRAM-033: スライドプレビューで図表を表示できる。
- DIAGRAM-034: 単一HTMLエクスポートへ図表描画結果を埋め込める。
- DIAGRAM-035: PDFエクスポートは単一HTMLと同じ図表描画結果を利用する。

## 数式

- DIAGRAM-040: `math`、`tex`、`latex` コードフェンスを数式として描画できる。
- DIAGRAM-041: 単独段落の `$$ ... $$` を表示数式として描画できる。
- DIAGRAM-042: インラインの `$...$` は自動解釈しない。
- DIAGRAM-043: 数式描画にはKaTeX CLIを利用できる。
- DIAGRAM-044: KaTeX実行ファイルのパスが空欄の場合はPATHから `katex` を検出する。
- DIAGRAM-045: KaTeXはMathMLを出力し、外部CSSやフォントを必須としない。
- DIAGRAM-046: KaTeXはstrict設定で実行する。
- DIAGRAM-047: 数式描画依存が利用できない場合は元のコードまたは `$$ ... $$` テキストを表示する。
- DIAGRAM-048: 数式描画結果を通常プレビュー、スライド、単一HTML、PDFへ反映する。
