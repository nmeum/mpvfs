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

	// TODO: There currently doesn't seem to be any way to retrieve
	// this information for a playlist-entry, only for the currently
	// loaded file.
	opts := map[string]string{
		"media-title": entry.Description,
	}

	_, err = l.mpv.ExecCmd("loadfile", entry.FileName, "append", opts)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}

func (l *playlist) Close() error {
	return nil
}
