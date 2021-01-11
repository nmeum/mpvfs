package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
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
	return 0, errors.New("not implemented")
}
