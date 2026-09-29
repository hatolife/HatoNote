// Package logring は容量上限付きの診断ログ出力を提供します。
package logring

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

// Writer は古い内容を捨てながら最新のログを保持します。
type Writer struct {
	mu   sync.Mutex
	file *os.File
	max  int64
}

// Open は既存ログを引き継いで容量上限付きWriterを開きます。
func Open(filename string, maxBytes int64) (*Writer, error) {
	if maxBytes < 1 {
		return nil, fmt.Errorf("ログ上限は1バイト以上にしてください")
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	writer := &Writer{file: file, max: maxBytes}
	if err := writer.trimExisting(); err != nil {
		file.Close()
		return nil, err
	}
	return writer, nil
}

// Write は新しい内容を保持し、上限を超える古い内容を破棄します。
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return 0, os.ErrClosed
	}
	written := len(p)
	if int64(len(p)) >= w.max {
		return written, w.replace(p[len(p)-int(w.max):])
	}
	info, err := w.file.Stat()
	if err != nil {
		return 0, err
	}
	if info.Size()+int64(len(p)) <= w.max {
		if _, err := w.file.Seek(0, io.SeekEnd); err != nil {
			return 0, err
		}
		_, err = w.file.Write(p)
		if err != nil {
			return 0, err
		}
		return written, nil
	}
	keep := w.max - int64(len(p))
	old := make([]byte, int(keep))
	if keep > 0 {
		if _, err := w.file.ReadAt(old, info.Size()-keep); err != nil && err != io.EOF {
			return 0, err
		}
	}
	data := append(old, p...)
	if err := w.replace(data); err != nil {
		return 0, err
	}
	return written, nil
}

// Close はログファイルを閉じます。
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

func (w *Writer) trimExisting() error {
	info, err := w.file.Stat()
	if err != nil {
		return err
	}
	if info.Size() <= w.max {
		_, err = w.file.Seek(0, io.SeekEnd)
		return err
	}
	data := make([]byte, int(w.max))
	if _, err := w.file.ReadAt(data, info.Size()-w.max); err != nil && err != io.EOF {
		return err
	}
	return w.replace(data)
}

func (w *Writer) replace(data []byte) error {
	if err := w.file.Truncate(0); err != nil {
		return err
	}
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := w.file.Write(data); err != nil {
		return err
	}
	return w.file.Sync()
}
