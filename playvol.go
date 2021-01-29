package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"strings"
)

type playvol struct {
	*playlistfs.BlockRecv

	state *playerState
	mpv   *mpv.Client
}

func newVol() (fileserver.File, error) {
	p := &playvol{state: state, mpv: mpvClient}
	p.BlockRecv = playlistfs.NewBlockRecv(p)
	return p, nil
}

func (c *playvol) CurrentReader() *strings.Reader {
	v := playlistfs.Volume{[]uint{c.state.Volume()}}
	return strings.NewReader(v.String() + "\n")
}

func (c *playvol) NextReader() *strings.Reader {
	v := playlistfs.Volume{[]uint{c.state.WaitVolume()}}
	return strings.NewReader(v.String() + "\n")
}

func (c *playvol) Write(off int64, p []byte) (int, error) {
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
