package main

import (
	"github.com/nmeum/mpvfs/fileserver"
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"

	"strings"
)

type playlist struct {
	*playlistfs.BlockRecv

	state *playerState
	mpv   *mpv.Client
}

func newPlaylist() (fileserver.File, error) {
	p := &playlist{state: state, mpv: mpvClient}
	p.BlockRecv = playlistfs.NewBlockRecv(p)
	return p, nil
}

func (l *playlist) CurrentReader() *strings.Reader {
	out := strings.Join(l.state.Playlist(), "\n")
	return strings.NewReader(out + "\n")
}

func (l *playlist) NextReader() *strings.Reader {
	out := l.state.WaitPlayist()
	return strings.NewReader(out + "\n")
}

func (l *playlist) Write(off int64, p []byte) (int, error) {
	entry, err := playlistfs.PlaylistCmd(p)
	if err != nil {
		return 0, err
	}

	_, err = l.mpv.ExecCmd("loadfile", entry.FileName, "append")
	if err != nil {
		return 0, err
	}

	// TODO: Somehow make sure that mpv/mpvfs stores the description
	// and returns it on read again.

	return len(p), nil
}
