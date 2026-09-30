package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hatolife/HatoNote/internal/workspace"
)

const userTemplateVersion = 1
const maxUserTemplates = 100
const maxUserTemplateContent = 1 << 20

type UserTemplate struct {
	ID      string `json:"id"`
	Scope   string `json:"scope"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type userTemplateStore struct {
	Version int            `json:"version"`
	Items   []UserTemplate `json:"items"`
}

func userTemplatePath(base string) string {
	return filepath.Join(base, "templates.json")
}

func validateUserTemplate(template UserTemplate) error {
	title := strings.TrimSpace(template.Title)
	if title == "" {
		return fmt.Errorf("テンプレート名を入力してください")
	}
	if len([]rune(title)) > 100 {
		return fmt.Errorf("テンプレート名は100文字以下にしてください")
	}
	switch template.Scope {
	case "page":
		if template.Kind != "mdz" {
			return fmt.Errorf("ページテンプレートは通常MDZ用だけ登録できます")
		}
	case "document":
		if template.Kind != "mdz" {
			return fmt.Errorf("文書テンプレートは通常MDZ用だけ登録できます")
		}
	default:
		return fmt.Errorf("テンプレートの対象が不正です")
	}
	if len(template.Content) > maxUserTemplateContent {
		return fmt.Errorf("テンプレート本文は1MiB以下にしてください")
	}
	if template.ID != "" && (!strings.HasPrefix(template.ID, "user.") || len(template.ID) > 80) {
		return fmt.Errorf("テンプレートIDが不正です")
	}
	return nil
}

func loadUserTemplates(base string) ([]UserTemplate, error) {
	data, err := os.ReadFile(userTemplatePath(base))
	if err != nil {
		if os.IsNotExist(err) {
			return []UserTemplate{}, nil
		}
		return nil, err
	}
	var store userTemplateStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("テンプレート設定を読み込めません: %w", err)
	}
	if store.Version != userTemplateVersion {
		return nil, fmt.Errorf("テンプレート設定の形式が不正です")
	}
	if len(store.Items) > maxUserTemplates {
		return nil, fmt.Errorf("テンプレート数が上限を超えています")
	}
	ids := map[string]bool{}
	for _, template := range store.Items {
		if err := validateUserTemplate(template); err != nil {
			return nil, err
		}
		if template.ID == "" || ids[template.ID] {
			return nil, fmt.Errorf("テンプレートIDが重複しています")
		}
		ids[template.ID] = true
	}
	return store.Items, nil
}

func saveUserTemplates(base string, items []UserTemplate) error {
	if len(items) > maxUserTemplates {
		return fmt.Errorf("ユーザー定義テンプレートは100件までです")
	}
	store := userTemplateStore{Version:userTemplateVersion, Items:items}
	data, err := json.MarshalIndent(store, "", "\t")
	if err != nil {
		return err
	}
	return workspace.AtomicWrite(userTemplatePath(base), data)
}

func newUserTemplateID() (string, error) {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return "user." + hex.EncodeToString(value[:]), nil
}

// UserTemplates はユーザー定義テンプレートの編集用情報を返します。
func (a *App) UserTemplates() ([]UserTemplate, error) {
	items, err := loadUserTemplates(a.base)
	if err != nil {
		return nil, err
	}
	result := make([]UserTemplate, len(items))
	copy(result, items)
	return result, nil
}

// SaveUserTemplate はユーザー定義テンプレートを追加または更新します。
func (a *App) SaveUserTemplate(template UserTemplate) (UserTemplate, error) {
	template.Title = strings.TrimSpace(template.Title)
	if err := validateUserTemplate(template); err != nil {
		return UserTemplate{}, err
	}
	items, err := loadUserTemplates(a.base)
	if err != nil {
		return UserTemplate{}, err
	}
	if template.ID == "" {
		if len(items) >= maxUserTemplates {
			return UserTemplate{}, fmt.Errorf("ユーザー定義テンプレートは100件までです")
		}
		template.ID, err = newUserTemplateID()
		if err != nil {
			return UserTemplate{}, err
		}
		items = append(items, template)
	} else {
		found := false
		for index := range items {
			if items[index].ID == template.ID {
				items[index] = template
				found = true
				break
			}
		}
		if !found {
			return UserTemplate{}, fmt.Errorf("編集するユーザー定義テンプレートがありません")
		}
	}
	if err := saveUserTemplates(a.base, items); err != nil {
		return UserTemplate{}, err
	}
	return template, nil
}

// DeleteUserTemplate はユーザー定義テンプレートだけを削除します。
func (a *App) DeleteUserTemplate(id string) error {
	id = strings.TrimSpace(id)
	if !strings.HasPrefix(id, "user.") {
		return fmt.Errorf("ユーザー定義テンプレートだけ削除できます")
	}
	items, err := loadUserTemplates(a.base)
	if err != nil {
		return err
	}
	next := make([]UserTemplate, 0, len(items))
	found := false
	for _, template := range items {
		if template.ID == id {
			found = true
			continue
		}
		next = append(next, template)
	}
	if !found {
		return fmt.Errorf("削除するテンプレートがありません")
	}
	return saveUserTemplates(a.base, next)
}

// userTemplate は指定IDのユーザー定義テンプレートを返します。
func (a *App) userTemplate(id string) (UserTemplate, bool, error) {
	items, err := loadUserTemplates(a.base)
	if err != nil {
		return UserTemplate{}, false, err
	}
	for _, template := range items {
		if template.ID == id {
			return template, true, nil
		}
	}
	return UserTemplate{}, false, nil
}
