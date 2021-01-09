package main

import (
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"time"
)

var ctlFiles = []string{
	"playctl",
	"playvol",
	"playlist",
}

func NewPlaylistFS() *FileServer {
	q := qid.Qid{Type: qid.TypeDir, Version: 0, Path: hashPath("/")}
	s := stat.Stat{
		Qid:    q,
		Mode:   0555 | stat.Dir,
		Atime:  time.Now(),
		Mtime:  time.Now(),
		Length: 0,
		Name:   "/",
	}

	children := make([]fs.Entry, len(ctlFiles))
	for i := 0; i < len(ctlFiles); i++ {
		name := ctlFiles[i]

		q := qid.Qid{Type: qid.TypeFile, Version: 0, Path: hashPath(name)}
		s := stat.Stat{
			Qid:    q,
			Mode:   0644,
			Atime:  time.Now(),
			Mtime:  time.Now(),
			Length: 0,
			Name:   name,
		}

		children[i] = fs.NewFile(s, nil)
	}

	dir := fs.NewDir(s, children)
	return NewFileServer(dir)
}
