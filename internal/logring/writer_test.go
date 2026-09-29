package logring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriterKeepsLatestBytesWithinLimit(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "HatoNote.log")
	writer, err := Open(filename, 12)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write([]byte("second\n")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > 12 || !strings.HasSuffix(string(data), "second\n") {
		t.Fatalf("log=%q size=%d", data, len(data))
	}
}

func TestWriterTrimsExistingAndOversizedWrites(t *testing.T) {
	filename := filepath.Join(t.TempDir(), "HatoNote.log")
	if err := os.WriteFile(filename, []byte("0123456789abcdef"), 0600); err != nil {
		t.Fatal(err)
	}
	writer, err := Open(filename, 8)
	if err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "89abcdef" {
		t.Fatalf("trimmed=%q err=%v", data, err)
	}
	if _, err := writer.Write([]byte("ABCDEFGHIJK")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(filename); err != nil || string(data) != "DEFGHIJK" {
		t.Fatalf("oversized=%q err=%v", data, err)
	}
}
