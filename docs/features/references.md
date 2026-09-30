# 引用管理

## 対象

- REFERENCES-001: 引用管理は通常MDZで利用できる。
- REFERENCES-002: 単一Markdown、mdBook、スライドでは引用管理UIを表示しない。
- REFERENCES-003: 引用元の正規データはMarkdown本文とは別にMDZのmanifest拡張へ保持する。

## データ

- REFERENCES-010: 引用元は一意なID、タイトル、URL、補足を保持できる。
- REFERENCES-011: 引用元のmanifestキーは `x-hatonote-references` とする。
- REFERENCES-012: URLと補足は空欄を許可する。
- REFERENCES-013: 本文から参照中の引用元は削除できない。

## 本文

- REFERENCES-020: 本文中の引用記法は `[@<引用ID>]` とする。
- REFERENCES-021: リッチ編集から現在位置へ引用記法を挿入できる。
- REFERENCES-022: 引用記法はプレビューで参考文献への番号付き参照として表示する。
- REFERENCES-023: ページ末尾の参考文献は引用元の正規データから生成できる。
- REFERENCES-024: 参考文献更新は既存のHatoNote管理ブロックだけを置換し、それ以外の本文を変更しない。
- REFERENCES-025: 参考文献は本文で最初に現れる引用順に並べる。
- REFERENCES-026: 本文から参照されていない引用元はそのページの参考文献へ出力しない。
- REFERENCES-027: 参考文献更新後も `[@<引用ID>]` 記法を維持する。

## UI

- REFERENCES-030: リッチ編集の書式バーから引用管理画面を開ける。
- REFERENCES-031: 引用管理画面から引用元を新規作成、更新、削除できる。
- REFERENCES-032: 引用管理画面から選択中の引用元を本文へ挿入できる。
- REFERENCES-033: 引用管理画面から現在ページ末尾の参考文献を更新できる。
