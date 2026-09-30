package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/document"
)

const referenceManifestKey = "x-hatonote-references"

var referenceIDValid = regexp.MustCompile("^[A-Za-z0-9_-]{1,80}$")

type Reference struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
	Note  string `json:"note"`
}

type ReferenceSet struct {
	Version int         `json:"version"`
	Items   []Reference `json:"items"`
}

func emptyReferences() ReferenceSet {
	return ReferenceSet{Version: 1, Items: []Reference{}}
}

func (a *App) referencesLocked() (ReferenceSet, error) {
	if a.session == nil || a.documentTypeLocked() != document.MDZ {
		return emptyReferences(), fmt.Errorf("引用管理は通常MDZで利用できます")
	}
	raw := a.session.Doc.Manifest[referenceManifestKey]
	if len(raw) == 0 {
		return emptyReferences(), nil
	}
	var result ReferenceSet
	if err := json.Unmarshal(raw, &result); err != nil {
		return ReferenceSet{}, fmt.Errorf("引用元データを読み込めません: %w", err)
	}
	if result.Version != 1 {
		return ReferenceSet{}, fmt.Errorf("未対応の引用元データです")
	}
	if result.Items == nil {
		result.Items = []Reference{}
	}
	ids := map[string]bool{}
	for _, item := range result.Items {
		if err := validateReference(item); err != nil {
			return ReferenceSet{}, err
		}
		if ids[item.ID] {
			return ReferenceSet{}, fmt.Errorf("引用IDが重複しています: %s", item.ID)
		}
		ids[item.ID] = true
	}
	return result, nil
}

func validateReference(value Reference) error {
	if !referenceIDValid.MatchString(value.ID) {
		return fmt.Errorf("引用IDが不正です")
	}
	if strings.TrimSpace(value.Title) == "" || len(value.Title) > 500 {
		return fmt.Errorf("引用タイトルは1〜500文字で指定してください")
	}
	if len(value.URL) > 2000 || len(value.Note) > 4000 {
		return fmt.Errorf("引用元のURLまたは補足が長すぎます")
	}
	if value.URL != "" {
		parsed, err := url.Parse(value.URL)
		if err != nil || (strings.ToLower(parsed.Scheme) != "http" && strings.ToLower(parsed.Scheme) != "https") || parsed.Host == "" {
			return fmt.Errorf("引用元URLはhttpまたはhttpsで指定してください")
		}
	}
	return nil
}

// References は通常MDZに保存された引用元を返します。
func (a *App) References() (ReferenceSet, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.referencesLocked()
}

// SaveReference は引用元を新規作成または更新します。
func (a *App) SaveReference(value Reference) (Reference, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.documentTypeLocked() != document.MDZ {
		return Reference{}, fmt.Errorf("引用管理は通常MDZで利用できます")
	}
	set, err := a.referencesLocked()
	if err != nil {
		return Reference{}, err
	}
	value.Title = strings.TrimSpace(value.Title)
	value.URL = strings.TrimSpace(value.URL)
	value.Note = strings.TrimSpace(value.Note)
	if value.ID == "" {
		used := map[string]bool{}
		for _, item := range set.Items {
			used[item.ID] = true
		}
		for i := 1; i <= 100000; i++ {
			id := fmt.Sprintf("ref-%03d", i)
			if !used[id] {
				value.ID = id
				break
			}
		}
	}
	if err := validateReference(value); err != nil {
		return Reference{}, err
	}
	found := false
	for i := range set.Items {
		if set.Items[i].ID == value.ID {
			set.Items[i] = value
			found = true
			break
		}
	}
	if !found {
		set.Items = append(set.Items, value)
	}
	sort.SliceStable(set.Items, func(i, j int) bool { return set.Items[i].ID < set.Items[j].ID })
	if err := a.session.SetManifestValue(referenceManifestKey, set); err != nil {
		return Reference{}, err
	}
	return value, nil
}

// DeleteReference は本文から参照されていない引用元だけを削除します。
func (a *App) DeleteReference(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if !referenceIDValid.MatchString(id) {
		return fmt.Errorf("引用IDが不正です")
	}
	set, err := a.referencesLocked()
	if err != nil {
		return err
	}
	marker := "[@" + id + "]"
	for name, data := range a.session.Doc.Files {
		if bundle.IsMarkdown(name) && strings.Contains(string(data), marker) {
			return fmt.Errorf("本文から参照中の引用元は削除できません")
		}
	}
	next := set.Items[:0]
	found := false
	for _, item := range set.Items {
		if item.ID == id {
			found = true
			continue
		}
		next = append(next, item)
	}
	if !found {
		return fmt.Errorf("引用元が見つかりません")
	}
	set.Items = append([]Reference(nil), next...)
	return a.session.SetManifestValue(referenceManifestKey, set)
}

// ApplyReferences は現在ページの引用記法から末尾の参考文献を再生成します。
func (a *App) ApplyReferences(page string) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.session == nil || a.documentTypeLocked() != document.MDZ {
		return "", fmt.Errorf("引用管理は通常MDZで利用できます")
	}
	data, ok := a.session.Doc.Files[page]
	if !ok || !bundle.IsMarkdown(page) {
		return "", fmt.Errorf("Markdownページを指定してください")
	}
	set, err := a.referencesLocked()
	if err != nil {
		return "", err
	}
	byID := map[string]Reference{}
	for _, item := range set.Items {
		byID[item.ID] = item
	}
	text := string(data)
	body := referenceBlockPattern.ReplaceAllString(text, "")
	ids := citationIDs(body)
	block := ""
	if len(ids) > 0 {
		var builder strings.Builder
		builder.WriteString("<!-- HATONOTE_REFERENCES_BEGIN -->\n")
		builder.WriteString("## 参考文献\n\n")
		for index, id := range ids {
			item, exists := byID[id]
			if !exists {
				return "", fmt.Errorf("未登録の引用元です: %s", id)
			}
			builder.WriteString(fmt.Sprintf("%d. <!-- HATONOTE_REF:%s --> %s\n", index+1, id, referenceMarkdown(item)))
		}
		builder.WriteString("<!-- HATONOTE_REFERENCES_END -->\n")
		block = builder.String()
	}
	result := strings.TrimSpace(body)
	if block != "" {
		if result != "" {
			result += "\n\n"
		}
		result += block
	}
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	if err := a.session.Put(page, []byte(result)); err != nil {
		return "", err
	}
	a.pending.Store(false)
	return result, nil
}

func citationIDs(text string) []string {
	result := []string{}
	seen := map[string]bool{}
	fence := byte(0)
	fenceLength := 0
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if len(line)-len(trimmed) <= 3 && len(trimmed) > 0 && (trimmed[0] == '`' || trimmed[0] == '~') {
			n := 0
			for n < len(trimmed) && trimmed[n] == trimmed[0] {
				n++
			}
			if fenceLength == 0 && n >= 3 {
				fence, fenceLength = trimmed[0], n
				continue
			}
			if fenceLength > 0 && trimmed[0] == fence && n >= fenceLength {
				fenceLength = 0
				continue
			}
		}
		if fenceLength != 0 {
			continue
		}
		for _, match := range citationPattern.FindAllStringSubmatch(line, -1) {
			if !seen[match[1]] {
				seen[match[1]] = true
				result = append(result, match[1])
			}
		}
	}
	return result
}

func referenceMarkdown(value Reference) string {
	title := escapeReferenceMarkdown(value.Title)
	content := title
	if value.URL != "" {
		content = "[" + title + "](" + strings.ReplaceAll(value.URL, ")", "%29") + ")"
	}
	if value.Note != "" {
		content += " — " + escapeReferenceMarkdown(value.Note)
	}
	return content
}

func escapeReferenceMarkdown(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_", "`", "\\`")
	return replacer.Replace(strings.ReplaceAll(value, "\n", " "))
}
