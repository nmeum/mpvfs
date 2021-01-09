package main

import (
	"go.rbn.im/neinp"
	"go.rbn.im/neinp/fid"
	"go.rbn.im/neinp/fs"
	"go.rbn.im/neinp/message"
	"go.rbn.im/neinp/qid"

	"context"
	"errors"
	"io"
)

const version = "9P2000"

type FileServer struct {
	neinp.NopP2000
	root fs.Entry
	fids *fid.Map
}

func NewFileServer(root fs.Entry) *FileServer {
	return &FileServer{root: root, fids: fid.New()}
}

func (f *FileServer) Version(ctx context.Context, msg message.TVersion) (message.RVersion, error) {
	if msg.Version != version {
		return message.RVersion{}, errors.New(message.BotchErrorString)
	}

	// TODO: Sanity check on msize
	return message.RVersion{Version: version, Msize: msg.Msize}, nil
}

func (f *FileServer) Attach(ctx context.Context, msg message.TAttach) (message.RAttach, error) {
	f.fids.Set(msg.Fid, f.root)
	return message.RAttach{Qid: f.root.Qid()}, nil
}

func (f *FileServer) Stat(ctx context.Context, msg message.TStat) (message.RStat, error) {
	entry, ok := f.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.RStat{}, errors.New(message.NoStatErrorString)
	}

	return message.RStat{Stat: entry.Stat()}, nil
}

func (f *FileServer) Walk(ctx context.Context, msg message.TWalk) (message.RWalk, error) {
	entry, ok := f.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.RWalk{}, errors.New(message.UnknownFidErrorString)
	}
	stat := entry.Stat()
	if !stat.IsDir() && len(msg.Wname) != 0 {
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
		if f.fids.Get(msg.Newfid) != nil {
			return message.RWalk{}, errors.New(message.DupFidErrorString)
		}
		f.fids.Set(msg.Newfid, entry)
	}

	return message.RWalk{Wqid: wqid}, nil
}

func (f *FileServer) Open(ctx context.Context, msg message.TOpen) (message.ROpen, error) {
	entry, ok := f.fids.Get(msg.Fid).(fs.Entry)
	if !ok {
		return message.ROpen{}, errors.New(message.UnknownFidErrorString)
	}

	err := entry.Open()
	if err != nil {
		return message.ROpen{}, err
	}

	return message.ROpen{Qid: entry.Qid()}, nil
}

func (f *FileServer) Read(ctx context.Context, msg message.TRead) (message.RRead, error) {
	entry, ok := f.fids.Get(msg.Fid).(fs.Entry)
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

func (f *FileServer) Clunk(ctx context.Context, msg message.TClunk) (message.RClunk, error) {
	f.fids.Delete(msg.Fid)
	return message.RClunk{}, nil
}
