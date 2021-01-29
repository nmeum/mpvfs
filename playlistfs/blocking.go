package playlistfs

import (
	"errors"
	"io"
	"strings"
)

type RecvInput interface {
	CurrentReader() *strings.Reader
	NextReader() *strings.Reader
}

type BlockRecv struct {
	// Absolute offset at which (and beyond which) EOF will be returned.
	// This member must be initalized with -1.
	eofAt int64

	// Absolute base offset used to calculate a relative offset for
	// the current string reader.
	baseOff int64

	// Current string reader on which the read function operates.
	reader *strings.Reader

	in RecvInput
}

func NewBlockRecv(in RecvInput) *BlockRecv {
	return &BlockRecv{eofAt: -1, in: in}
}

func (e *BlockRecv) Read(off int64, p []byte) (int, error) {
	if e.reader == nil {
		e.reader = e.in.CurrentReader()
	} else if e.eofAt > 0 && off >= e.eofAt {
		// We are reading beyond EOF for the second time.
		// Block until new data is available and return it.
		e.baseOff += e.reader.Size()
		e.reader = e.in.NextReader()
	} else if off < e.baseOff {
		return 0, errors.New("invalid seek")
	}

	// Calculate offset relative to current reader
	relOff := off - e.baseOff

	_, err := e.reader.Seek(relOff, io.SeekStart)
	if err != nil {
		e.eofAt = off
		return 0, io.EOF
	}

	n, err := e.reader.Read(p)
	if err == io.EOF || (err != nil && n == 0) {
		e.eofAt = off + int64(n)
		return 0, io.EOF
	}

	return n, nil
}
