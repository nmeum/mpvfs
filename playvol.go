package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"strings"
)

type playvol struct {
	*blockFile

	state *playerState
	mpv   *mpv.Client
}

func newVol() (fileserver.File, error) {
	p := &playvol{state: state, mpv: mpvClient}
	p.blockFile = newBlockFile(p.getReader)
	return p, nil
}

func (c *playvol) getReader(block bool) *strings.Reader {
	var vol uint
	if block {
		vol = c.state.WaitVolume()
	} else {
		vol = c.state.Volume()
	}

	v := playlistfs.Volume{[]uint{vol}}
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
