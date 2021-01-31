package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
	"strings"
)

var ErrEmptyPlaylist = errors.New("playlist is empty")

type playctl struct {
	*playlistfs.BlockRecv

	state *playerState
	mpv   *mpv.Client
}

func newCtl() (fileserver.File, error) {
	c := &playctl{state: state, mpv: mpvClient}
	c.BlockRecv = playlistfs.NewBlockRecv(c)
	return c, nil
}

func (c *playctl) StateReader(pos int, pback Playback) *strings.Reader {
	// XXX: This will set position to -1 on stop
	cmd := playlistfs.Control{
		Name: pback.String(),
		Arg:  &pos,
	}

	return strings.NewReader(cmd.String() + "\n")
}

func (c *playctl) CurrentReader() *strings.Reader {
	return c.StateReader(c.state.State())
}

func (c *playctl) NextReader() *strings.Reader {
	return c.StateReader(c.state.WaitState())
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
