package playlistfs

import (
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"bytes"
	"time"
)

const playVolumeName = "playvol"

type playVolume struct {
	*fs.File
}

func newPlayVolume() *playVolume {
	q := qid.Qid{Type: qid.TypeDir, Version: 0, Path: hashPath(playVolumeName)}
	s := stat.Stat{
		Qid:    q,
		Mode:   0644,
		Atime:  time.Now(),
		Mtime:  time.Now(),
		Length: 0,
		Name:   playVolumeName,
	}

	return &playVolume{File: fs.NewFile(s, nil)}
}

func (v *playVolume) Open() error {
	v.ReadSeeker = bytes.NewReader([]byte("volume 100"))
	return nil
}
