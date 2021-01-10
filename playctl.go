package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"
)

type playctl struct {
	mpv *mpv.Client
}

func (c playctl) Read(off int64, p []byte) (int, error) {
	return 0, nil
}

func (c playctl) Write(off int64, p []byte) (int, error) {
	cmd, err := playlistfs.ParseCtlCmd(p)
	if err != nil {
		return 0, err
	}

	switch cmd.Name {
	case "stop":
		panic("not implemented")
	case "pause":
		panic("not implemented")
	case "play":
		err := c.mpv.SetProperty("pause", false)
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
