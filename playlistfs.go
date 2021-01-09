package main

import (
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"time"
)

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

	children := []fs.Entry{
		newPlayVolume(),
	}

	dir := fs.NewDir(s, children)
	return NewFileServer(dir)
}
