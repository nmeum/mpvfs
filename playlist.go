package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"
	"go.rbn.im/neinp/message"

	"io"
	"strings"
)

type playlist struct {
	state *playerState
	mpv   *mpv.Client
}

func (l playlist) Open(message.OpenMode) error {
	return nil
}

func (l playlist) Read(off int64, p []byte) (int, error) {
	playlist := strings.Join(l.state.Playlist(), "\n")
	reader := strings.NewReader(playlist + "\n")

	_, err := reader.Seek(off, io.SeekStart)
	if err != nil {
		return 0, io.EOF
	}

	return reader.Read(p)
}

func (l playlist) Write(off int64, p []byte) (int, error) {
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
