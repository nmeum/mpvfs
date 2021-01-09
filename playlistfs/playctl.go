package playlistfs

import (
	"go.rbn.im/neinp/fs"

	"bytes"
)

const playControlName = "playctl"

type playControl struct {
	*fs.File
}

func newPlayControl() *playControl {
	stat := createStat(playControlName, 0644)
	return &playControl{File: fs.NewFile(stat, nil)}
}

func (v *playControl) Open() error {
	v.ReadSeeker = bytes.NewReader([]byte("cmd stop"))
	return nil
}
