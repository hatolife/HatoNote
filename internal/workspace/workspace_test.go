package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hatolife/HatoNote/internal/bundle"
	"github.com/hatolife/HatoNote/internal/settings"
)


func testDocument() *bundle.Document {
	d := bundle.New()
	d.Files["index.md"] = []byte("# Test\n")
	d.Entry = "index.md"
	return d
}

func TestWorkspaceSaveBackupsAndConflict(t *testing.T) {
	base := t.TempDir()
	cfg := settings.Default()
	cfg.BackupGenerations = 2
	s, err := New(base, testDocument(), "", false)
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "document.mdz")
	for _, text := range []string{"one", "two", "three", "four"} {
		if err := s.Put("index.md", []byte(text)); err != nil {
			t.Fatal(err)
		}
		if err := s.Save(base, file, "test", cfg); err != nil {
			t.Fatal(err)
		}
	}
	d, err := bundle.Read(file)
	if err != nil || string(d.Files["index.md"]) != "four" {
		t.Fatal("save failed", err)
	}
	var archives []string
	filepath.WalkDir(filepath.Join(base, "backups"), func(p string, e os.DirEntry, err error) error {
		if err == nil && strings.HasSuffix(p, ".mdz") {
			archives = append(archives, p)
		}
		return err
	})
	if len(archives) != 2 {
		t.Fatalf("generations=%d", len(archives))
	}
	for i, want := range []string{"two", "three"} {
		d, err := bundle.Read(archives[i])
		if err != nil || string(d.Files["index.md"]) != want {
			t.Fatalf("backup %d: %v", i, err)
		}
	}
	if err := os.WriteFile(file, []byte("externally changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(base, file, "test", cfg); err == nil {
		t.Fatal("overwrote external change")
	}
	if b, _ := os.ReadFile(file); string(b) != "externally changed" {
		t.Fatal("lost external data")
	}
}

func TestRecoveryAndImageRetention(t *testing.T) {
	base := t.TempDir()
	s, err := New(base, testDocument(), "", false)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Put("images/paste.png", []byte{1, 2, 3}); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("index.md", []byte("![x](images/paste.png)")); err != nil {
		t.Fatal(err)
	}
	// 本文のUndoを再現し、画像を保持することを確認します。
	if err := s.Put("index.md", []byte("undone")); err != nil {
		t.Fatal(err)
	}
	s.PID = 0
	if err := s.Checkpoint(); err != nil {
		t.Fatal(err)
	}
	list, err := Recoveries(base)
	if err != nil || len(list) != 1 {
		t.Fatal("recovery missing", err)
	}
	r, err := Recover(base, s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(r.Doc.Files["index.md"]) != "undone" || len(r.Doc.Files["images/paste.png"]) != 3 {
		t.Fatal("recovery lost content")
	}
	if !r.Dirty {
		t.Fatal("recovered workspace marked clean")
	}
	if _, err := Recover(base, "../escape"); err == nil {
		t.Fatal("unsafe recovery accepted")
	}
}

func TestPortablePathsAndSymlinks(t *testing.T) {
	for _, name := range []string{"../x.md", "CON.md", "images/AUX.png", "a./x.md", "x /x.md"} {
		if PortablePath(name) {
			t.Fatal("accepted", name)
		}
	}
	base := t.TempDir()
	d := testDocument()
	d.Files["INDEX.md"] = []byte("duplicate")
	if _, err := New(base, d, "", false); err == nil {
		t.Fatal("case collision accepted")
	}
	s, err := New(base, testDocument(), "", false)
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "secret.md")
	os.WriteFile(outside, []byte("secret"), 0600)
	if err := os.Symlink(outside, filepath.Join(s.Content(), "link.md")); err != nil {
		t.Skip("symlink unavailable")
	}
	if err := s.Capture(); err == nil {
		t.Fatal("symlink accepted")
	}
}


func TestHistoryAndNamedVersions(t *testing.T) {
	base := t.TempDir()
	cfg := settings.Default()
	cfg.BackupGenerations = 5
	session, err := New(base, testDocument(), "", false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	filename := filepath.Join(t.TempDir(), "history.mdz")
	if err := session.Save(base, filename, "test", cfg); err != nil {
		t.Fatal(err)
	}
	if err := session.Put("index.md", []byte("saved second")); err != nil {
		t.Fatal(err)
	}
	if err := session.Save(base, filename, "test", cfg); err != nil {
		t.Fatal(err)
	}
	entries, err := History(base, filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Kind != "backup" {
		t.Fatalf("automatic history = %+v", entries)
	}
	if err := session.Put("index.md", []byte("named unsaved")); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	named, err := session.SaveNamedVersion(base, "test", "確認用")
	if err != nil {
		t.Fatal(err)
	}
	if named.Kind != "named" || named.Name != "確認用" || !strings.HasPrefix(named.ID, "named/") {
		t.Fatalf("named history = %+v", named)
	}
	after, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("named version changed the saved document")
	}
	if !session.Dirty {
		t.Fatal("named version unexpectedly cleared dirty state")
	}
	entries, err = History(base, filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Kind != "named" {
		t.Fatalf("history = %+v", entries)
	}
	restored, err := LoadHistory(base, filename, named.ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(restored.Files["index.md"]) != "named unsaved" {
		t.Fatalf("restored named version = %q", restored.Files["index.md"])
	}
	backupDoc, err := LoadHistory(base, filename, entries[1].ID)
	if err != nil {
		t.Fatal(err)
	}
	if string(backupDoc.Files["index.md"]) != "# Test\n" {
		t.Fatalf("restored automatic backup = %q", backupDoc.Files["index.md"])
	}
}

func TestNamedVersionRequiresSavedDocumentAndName(t *testing.T) {
	session, err := New(t.TempDir(), testDocument(), "", true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = session.Close() })
	if _, err := session.SaveNamedVersion(t.TempDir(), "test", "name"); err == nil {
		t.Fatal("unsaved document accepted a named version")
	}
	session.Filename = filepath.Join(t.TempDir(), "saved.mdz")
	if _, err := session.SaveNamedVersion(t.TempDir(), "test", "   "); err == nil {
		t.Fatal("empty name accepted")
	}
}
