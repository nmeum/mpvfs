package playlistfs

import (
	"github.com/nmeum/mpvfs/fileserver"
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"time"
)

func NewPlaylistFS() *fileserver.FileServer {
	q := qid.Qid{Type: qid.TypeDir, Version: 0, Path: hashPath("/")}
	s := stat.Stat{
		Qid:    q,
		Mode:   0555 | stat.Dir,
		Atime:  time.Now(),
		Mtime:  time.Now(),
		Length: 0,
		Name:   "/",
	}

	children := []fs.Entry{
		newPlayVolume(),
		newPlayControl(),
	}

	dir := fs.NewDir(s, children)
	return fileserver.NewFileServer(dir)
}
