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

	pos  chan int
	play chan bool

	state *playerState
	mpv   *mpv.Client
}

func newCtl() (fileserver.File, error) {
	c := &playctl{
		state: state,
		mpv:   mpvClient,
		pos:   make(chan int, 1),
		play:  make(chan bool, 1),
	}

	c.BlockRecv = playlistfs.NewBlockRecv(c)
	return c, nil
}

func (c *playctl) StateReader(pos int, playing bool) *strings.Reader {
	var str string
	if pos == -1 {
		str = "stop"
	} else if playing {
		str = "play"
	} else {
		str = "pause"
	}

	// XXX: This will set position to -1 on stop
	cmd := playlistfs.Control{Name: str, Arg: &pos}
	return strings.NewReader(cmd.String() + "\n")
}

func (c *playctl) CurrentReader() *strings.Reader {
	reader := c.StateReader(c.state.Index(), c.state.Playing())

	// TODO: Stop these goroutines on clunk
	go func(ch chan<- int) {
		for {
			ch <- c.state.WaitIndex()
		}
	}(c.pos)
	go func(ch chan<- bool) {
		for {
			ch <- c.state.WaitPlaying()
		}
	}(c.play)

	return reader
}

func (c *playctl) NextReader() *strings.Reader {
	var pos int
	var playing bool

	select {
	case i := <-c.pos:
		pos = i
		playing = c.state.Playing()
	case p := <-c.play:
		playing = p
		pos = c.state.Index()
	}

	return c.StateReader(pos, playing)
}

func (c *playctl) Write(off int64, p []byte) (int, error) {
	cmd, err := playlistfs.CtlCmd(p)
	if err != nil {
		return 0, err
	}

	if cmd.Arg != nil {
		err := c.mpv.SetProperty("playlist-pos", *cmd.Arg)
		if err != nil {
			return 0, err
		}
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

		idx := c.state.Index()
		if idx == -1 {
			inc = 1 // Start from beginning
		}

		newArg := idx + inc
		cmd.Arg = &newArg

		fallthrough
	case "play":
		fallthrough
	case "resume":
		err := c.mpv.SetProperty("pause", false)
		if err != nil {
			return 0, err
		}
	}

	return len(p), nil
}
