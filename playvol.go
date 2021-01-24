package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"errors"
	"io"
	"strings"
)

type playvol struct {
	// Absolute offset at which (and beyond which) EOF will be returned.
	// This member must be initalized with -1.
	eofAt int64

	// Absolute base offset used to calculate a relative offset for
	// the current string reader.
	baseOff int64

	// Current string reader on which the read function operates.
	reader *strings.Reader

	state *playerState
	mpv   *mpv.Client
}

func newVol() (fileserver.File, error) {
	return &playvol{eofAt: -1, state: state, mpv: mpvClient}, nil
}

func (c *playvol) newReader(volume uint) *strings.Reader {
	vol := playlistfs.Volume{[]uint{volume}}
	return strings.NewReader(vol.String() + "\n")
}

func (c *playvol) Read(off int64, p []byte) (int, error) {
	if c.reader == nil {
		c.reader = c.newReader(c.state.Volume())
	} else if c.eofAt > 0 && off >= c.eofAt {
		// We are reading beyond EOF for the second time.
		// Block until new data is available and return it.
		c.baseOff += c.reader.Size()
		c.reader = c.newReader(c.state.WaitVolume())
	} else if off < c.baseOff {
		return 0, errors.New("invalid seek")
	}

	// Calculate offset relative to current reader
	relOff := off - c.baseOff

	_, err := c.reader.Seek(relOff, io.SeekStart)
	if err != nil {
		c.eofAt = off
		return 0, io.EOF
	}

	n, err := c.reader.Read(p)
	if err == io.EOF || (err != nil && n == 0) {
		c.eofAt = off + int64(n)
		return 0, io.EOF
	}

	return n, nil
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
