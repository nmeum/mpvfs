package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
	"io"
	"strings"
)

type playctl struct {
	state *playerState
	mpv   *mpv.Client
}

func (c playctl) Read(off int64, p []byte) (int, error) {
	var name string
	if c.state.IsPlaying() {
		name = "play"
	} else {
		name = "pause"
	}

	pos := c.state.Index()
	if pos < 0 {
		return 0, errors.New("no current playlist")
	}
	upos := uint(pos)

	cmd := playlistfs.Control{Name: name, Arg: &upos}
	reader := strings.NewReader(cmd.String() + "\n")

	_, err := reader.Seek(off, io.SeekStart)
	if err != nil {
		return 0, io.EOF
	}

	return reader.Read(p)
}

func (c playctl) Write(off int64, p []byte) (int, error) {
	cmd, err := playlistfs.CtlCmd(p)
	if err != nil {
		return 0, err
	}

	switch cmd.Name {
	case "stop":
		err := c.mpv.SetProperty("playlist-pos", 0)
		if err != nil {
			return 0, err
		}

		fallthrough
	case "pause":
		err := c.mpv.SetProperty("pause", true)
		if err != nil {
			return 0, err
		}
	case "skip":
		var inc uint
		if cmd.Arg == nil {
			inc = 1
		}

		idx := c.state.Index()
		if idx > 0 {
			inc += uint(idx)
		}

		cmd.Arg = &inc
		fallthrough
	case "play":
		if cmd.Arg == nil {
			break // Ignore
		}

		err := c.mpv.SetProperty("playlist-pos", *cmd.Arg)
		if err != nil {
			return 0, err
		}

		fallthrough
	case "resume":
		err := c.mpv.SetProperty("pause", false)
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
