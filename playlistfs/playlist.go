package playlistfs

import (
	"go.rbn.im/neinp/fs"

	"bytes"
)

const playlistName = "playlist"

type playlist struct {
	*fs.File
}

func newPlaylist() *playlist {
	stat := createStat(playlistName, 0644)
	return &playlist{File: fs.NewFile(stat, nil)}
}

func (v *playlist) Open() error {
	v.ReadSeeker = bytes.NewReader([]byte("/media/openbsd/song32.ogg"))
	return nil
}
