package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
	"io"
	"strings"
)

var ErrEmptyPlaylist = errors.New("playlist is empty")

type playctl struct {
	state *playerState
	mpv   *mpv.Client
}

func newCtl() (fileserver.File, error) {
	return &playctl{state, mpvClient}, nil
}

func (c *playctl) Read(off int64, p []byte) (int, error) {
	pos, pback := c.state.State()

	var name string
	switch (pback) {
	case Playing:
		name = "play"
	case Paused:
		name = "pause"
	case Stopped:
		name = "stop"
	}

	// XXX: This will set position to -1 on stop
	cmd := playlistfs.Control{Name: name, Arg: &pos}
	reader := strings.NewReader(cmd.String() + "\n")

	_, err := reader.Seek(off, io.SeekStart)
	if err != nil {
		return 0, io.EOF
	}

	return reader.Read(p)
}

func (c *playctl) Write(off int64, p []byte) (int, error) {
	cmd, err := playlistfs.CtlCmd(p)
	if err != nil {
		return 0, err
	}

	switch cmd.Name {
	case "stop":
		err := c.mpv.SetProperty("playlist-pos", -1)
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
		var inc int
		if cmd.Arg != nil {
			inc = *cmd.Arg
		} else {
			inc = 1
		}

		idx, _ := c.state.State()
		if idx == -1 {
			inc = 1 // Start from beginning
		}

		newArg := idx + inc
		cmd.Arg = &newArg

		fallthrough
	case "play":
		if cmd.Arg != nil {
			err := c.mpv.SetProperty("playlist-pos", *cmd.Arg)
			if err != nil {
				return 0, err
			}
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
