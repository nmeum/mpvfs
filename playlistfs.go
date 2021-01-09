package main

import (
	"go.rbn.im/neinp"
	"go.rbn.im/neinp/fid"
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/message"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"context"
	"errors"
	"io"
	"time"
)

const version = "9P2000"

var ctlFiles = []string{
	"playctl",
	"playvol",
	"playlist",
}

// TODO: FileServer Abstraction which only requires to model files, not the server.
type PlaylistFS struct {
	neinp.NopP2000
	root fs.Entry

	fids *fid.Map
}

func NewPlaylistFS() *PlaylistFS {
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
	return &PlaylistFS{root: dir, fids: fid.New()}
}

func (r *PlaylistFS) Version(ctx context.Context, msg message.TVersion) (message.RVersion, error) {
	if msg.Version != version {
		return message.RVersion{}, errors.New(message.BotchErrorString)
	}

	// TODO: Sanity check on msize
	return message.RVersion{Version: version, Msize: msg.Msize}, nil
}

func (p *PlaylistFS) Attach(ctx context.Context, msg message.TAttach) (message.RAttach, error) {
	p.fids.Set(msg.Fid, p.root)
	return message.RAttach{Qid: p.root.Qid()}, nil
}

func (p *PlaylistFS) Stat(ctx context.Context, msg message.TStat) (message.RStat, error) {
	entry, ok := p.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.RStat{}, errors.New(message.NoStatErrorString)
	}

	return message.RStat{Stat: entry.Stat()}, nil
}

func (p *PlaylistFS) Walk(ctx context.Context, msg message.TWalk) (message.RWalk, error) {
	entry, ok := p.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.RWalk{}, errors.New(message.UnknownFidErrorString)
	}
	stat := entry.Stat()
	if !stat.IsDir() {
		return message.RWalk{}, errors.New(message.WalkNoDirErrorString)
	}

	wqid := make([]qid.Qid, len(msg.Wname))
	for i := 0; i < len(msg.Wname); i++ {
		var err error
		entry, err = entry.Walk(msg.Wname[i])
		if err != nil {
			return message.RWalk{}, err
		}
		wqid[i] = entry.Qid()
	}

	if msg.Newfid != msg.Fid {
		if p.fids.Get(msg.Newfid) != nil {
			return message.RWalk{}, errors.New(message.DupFidErrorString)
		}
		p.fids.Set(msg.Newfid, entry)
	}

	return message.RWalk{Wqid: wqid}, nil
}

func (p *PlaylistFS) Open(ctx context.Context, msg message.TOpen) (message.ROpen, error) {
	entry, ok := p.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.ROpen{}, errors.New(message.UnknownFidErrorString)
	}

	err := entry.Open()
	if err != nil {
		return message.ROpen{}, err
	}

	return message.ROpen{Qid: entry.Qid()}, nil
}

func (p *PlaylistFS) Read(ctx context.Context, msg message.TRead) (message.RRead, error) {
	entry, ok := p.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.RRead{}, errors.New(message.UnknownFidErrorString)
	}

	_, err := entry.Seek(int64(msg.Offset), io.SeekStart)
	if err != nil {
		return message.RRead{}, err
	}

	// TODO: Sanity check count
	buf := make([]byte, msg.Count)
	n, err := entry.Read(buf)
	if err != nil && err != io.EOF {
		return message.RRead{}, err
	}

	return message.RRead{Count: uint32(n), Data: buf[:n]}, nil
}

func (p *PlaylistFS) Clunk(ctx context.Context, msg message.TClunk) (message.RClunk, error) {
	p.fids.Delete(msg.Fid)
	return message.RClunk{}, nil
}
