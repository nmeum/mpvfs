package fileserver

import (
	"go.rbn.im/neinp"
	"go.rbn.im/neinp/fid"
	"go.rbn.im/neinp/message"
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"context"
	"errors"
	"io"
)

type File interface {
	Open(message.OpenMode) error
	Read(int64, []byte) (int, error)
	Write(int64, []byte) (int, error)
}

type FileMap map[string]File

type pair struct {
	stat stat.Stat
	file File
}

type FileServer struct {
	neinp.NopP2000
	files map[string]*pair
	fids  *fid.Map
}

const version = "9P2000"

// TODO: Currently only supports a flat hierarchy with one root
// directory and several regular files located inside this directory.

func NewFileServer(files FileMap) *FileServer {
	fs := &FileServer{files: make(map[string]*pair), fids: fid.New()}

	root := createStat("/", 0555|stat.Dir)
	rootDir := directory{stat: root}

	for name, file := range files {
		s := createStat(name, 0644)

		fs.files[name] = &pair{s, file}
		rootDir.children = append(rootDir.children, s)
	}
	fs.files[root.Name] = &pair{root, rootDir}

	return fs
}

func (f *FileServer) Version(ctx context.Context, msg message.TVersion) (message.RVersion, error) {
	if msg.Version != version {
		return message.RVersion{}, errors.New(message.BotchErrorString)
	}

	// TODO: Sanity check on msize
	return message.RVersion{Version: version, Msize: msg.Msize}, nil
}

func (f *FileServer) Attach(ctx context.Context, msg message.TAttach) (message.RAttach, error) {
	pair := f.files["/"]

	f.fids.Set(msg.Fid, pair)
	return message.RAttach{Qid: pair.stat.Qid}, nil
}

func (f *FileServer) Stat(ctx context.Context, msg message.TStat) (message.RStat, error) {
	pair, ok := f.fids.Get(msg.Fid).(*pair)
	if !ok {
		return message.RStat{}, errors.New(message.NoStatErrorString)
	}

	return message.RStat{Stat: pair.stat}, nil
}

func (f *FileServer) Walk(ctx context.Context, msg message.TWalk) (message.RWalk, error) {
	pair, ok := f.fids.Get(msg.Fid).(*pair)
	if !ok {
		return message.RWalk{}, errors.New(message.UnknownFidErrorString)
	} else if len(msg.Wname) > 1 {
		return message.RWalk{}, errors.New(message.WalkNoDirErrorString)
	}

	newPair := pair
	if len(msg.Wname) == 1 {
		name := msg.Wname[0]
		newPair, ok = f.files[name]
		if !ok {
			return message.RWalk{}, errors.New(message.NotFoundErrorString)
		}
	}

	if msg.Newfid != msg.Fid {
		if f.fids.Get(msg.Newfid) != nil {
			return message.RWalk{}, errors.New(message.DupFidErrorString)
		}
		f.fids.Set(msg.Newfid, newPair)
	}

	if len(msg.Wname) == 1 {
		return message.RWalk{Wqid: []qid.Qid{newPair.stat.Qid}}, nil
	} else {
		return message.RWalk{}, nil
	}
}

func (f *FileServer) Open(ctx context.Context, msg message.TOpen) (message.ROpen, error) {
	pair, ok := f.fids.Get(msg.Fid).(*pair)
	if !ok {
		return message.ROpen{}, errors.New(message.UnknownFidErrorString)
	}

	err := pair.file.Open(msg.Mode)
	if err != nil {
		return message.ROpen{}, err
	}

	return message.ROpen{Qid: pair.stat.Qid}, nil
}

func (f *FileServer) Read(ctx context.Context, msg message.TRead) (message.RRead, error) {
	pair, ok := f.fids.Get(msg.Fid).(*pair)
	if !ok {
		return message.RRead{}, errors.New(message.UnknownFidErrorString)
	}

	// TODO: Sanity check count
	buf := make([]byte, msg.Count)
	n, err := pair.file.Read(int64(msg.Offset), buf)
	if err != nil && err != io.EOF {
		return message.RRead{}, err
	}

	return message.RRead{Count: uint32(n), Data: buf[:n]}, nil
}

func (f *FileServer) Write(ctx context.Context, msg message.TWrite) (message.RWrite, error) {
	pair, ok := f.fids.Get(msg.Fid).(*pair)
	if !ok {
		return message.RWrite{}, errors.New(message.UnknownFidErrorString)
	}

	n, err := pair.file.Write(int64(msg.Offset), msg.Data)
	if err != nil {
		return message.RWrite{}, err
	}

	return message.RWrite{Count: uint32(n)}, nil
}

func (f *FileServer) Clunk(ctx context.Context, msg message.TClunk) (message.RClunk, error) {
	f.fids.Delete(msg.Fid)
	return message.RClunk{}, nil
}
