package zlog

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/codeslala/igo/errs"
)

type FileWriter struct {
	File string

	MaxSize    int64
	Truncation bool

	fd   *os.File
	size int64
	mu   sync.Mutex
}

func (w *FileWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.fd == nil {
		if w.File == "" {
			return 0, errs.InvalidConfig.New("missing FileWriter.File")
		}

		CreateFileDir(w.File)

		fd, err := os.OpenFile(w.File,
			os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return 0, err
		}
		info, err := os.Stat(w.File)
		if err != nil {
			return 0, err
		}
		w.fd = fd
		w.size = info.Size()
	}

	writeLen := int64(len(b))
	if w.MaxSize != 0 && writeLen > w.MaxSize {
		return 0, errs.Internal.Newf("write length %d exceeds maximum file size %d", writeLen, w.MaxSize)
	}

	if w.MaxSize != 0 && w.size+writeLen > w.MaxSize {
		if w.Truncation {
			err := os.Truncate(w.File, 0)
			if err != nil {
				return 0, err
			}
			w.size = 0
		} else {
			err := w.rotate()
			if err != nil {
				return 0, err
			}
		}
	}

	n, err := w.fd.Write(b)
	w.size += int64(n)

	return n, err
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.close()
}

func (w *FileWriter) close() error {
	if w.fd == nil {
		return nil
	}
	err := w.fd.Close()
	w.fd = nil
	return err
}

func (w *FileWriter) rotate() error {
	err := w.close()
	if err != nil {
		return err
	}
	err = os.Rename(w.File, w.backupName())
	if err != nil {
		return err
	}
	fd, err := os.OpenFile(w.File,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	w.fd = fd
	w.size = 0

	return nil
}

func (w *FileWriter) backupName() string {
	newFile := ""
	for i := 1; i <= 100; i++ {
		if _, err := os.Stat(w.File + "." + strconv.Itoa(i)); err == nil {
			continue
		} else {
			newFile = w.File + "." + strconv.Itoa(i)
		}
	}
	return newFile
}

type logFiles []struct {
	timestamp time.Time
	os.FileInfo
}

func (b logFiles) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b logFiles) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b logFiles) Len() int {
	return len(b)
}

func CreateFileDir(files ...string) {
	for _, file := range files {
		if file == "" {
			continue
		}
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			must(err)
		}
		if err := os.Chmod(dir, os.ModePerm); err != nil {
			panic(err)
		}
	}
}
