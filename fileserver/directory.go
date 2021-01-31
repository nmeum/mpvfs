package fileserver

import (
	"go.rbn.im/neinp/stat"

	"io"
	"os"
)

type directory struct {
	stat     stat.Stat
	children []stat.Stat
}

func (d directory) Read(off int64, p []byte) (int, error) {
	reader := stat.NewReader(d.children...)

	n, err := reader.Seek(off, io.SeekStart)
	if err != nil {
		return int(n), io.EOF
	}

	return reader.Read(p)
}

func (d directory) Write(off int64, p []byte) (int, error) {
	return 0, os.ErrInvalid
}

func (d directory) Close() error {
	return nil
}
