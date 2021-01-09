package playlistfs

import (
	"go.rbn.im/neinp/fs"

	"bytes"
)

const playVolumeName = "playvol"

type playVolume struct {
	*fs.File
}

func newPlayVolume() *playVolume {
	stat := createStat(playVolumeName, 0644)
	return &playVolume{File: fs.NewFile(stat, nil)}
}

func (v *playVolume) Open() error {
	v.ReadSeeker = bytes.NewReader([]byte("volume 100"))
	return nil
}
