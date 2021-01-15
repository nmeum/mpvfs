package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
)

type playlist struct {
	state *playerState
	mpv   *mpv.Client
}

func (l playlist) Read(off int64, p []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (l playlist) Write(off int64, p []byte) (int, error) {
	entry, err := playlistfs.PlaylistCmd(p)
	if err != nil {
		return 0, err
	}

	_, err = l.mpv.ExecCmd("loadfile", entry.FileName, "append")
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
