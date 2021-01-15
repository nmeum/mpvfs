package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"io"
	"strings"
)

type playvol struct {
	state *playerState
	mpv   *mpv.Client
}

func (c playvol) Read(off int64, p []byte) (int, error) {
	cmd := playlistfs.Command{Name: "vol", Arg: c.state.Volume()}
	reader := strings.NewReader(cmd.String() + "\n")

	_, err := reader.Seek(off, io.SeekStart)
	if err != nil {
		return 0, io.EOF
	}

	return reader.Read(p)
}

func (c playvol) Write(off int64, p []byte) (int, error) {
	vol, err := playlistfs.VolCmd(p)
	if err != nil {
		return 0, err
	}

	// TODO: handle channels
	level := float64(vol.Levels[0])
	err = c.mpv.SetProperty("volume", level)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
