package playlistfs

import (
	"github.com/nmeum/mpvfs/fileserver"
)

type Config struct {
	Playctl  fileserver.Cons
	Playlist fileserver.Cons
	Playvol  fileserver.Cons
}

func NewPlaylistFS(c Config) *fileserver.FileServer {
	return fileserver.NewFileServer(fileserver.FileMap{
		"playctl":  c.Playctl,
		"playlist": c.Playlist,
		"playvol":  c.Playvol,
	})
}
