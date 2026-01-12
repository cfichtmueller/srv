package spa

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
)

type Reader interface {
	Read(filename string) (b []byte, ok bool, err error)
}

type FsReader struct {
	fs fs.FS
}

func (r *FsReader) Read(filename string) (b []byte, ok bool, err error) {
	f, err := r.fs.Open(filename)
	if err != nil {
		slog.Error("spa handler unable to open file", "error", err)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, false, nil
		}
	}
	stat, err := f.Stat()
	if err != nil {
		return nil, false, err
	}
	if stat.IsDir() {
		return nil, false, nil
	}
	b, err = io.ReadAll(f)
	if err != nil {
		return nil, false, err
	}
	return b, true, nil
}
