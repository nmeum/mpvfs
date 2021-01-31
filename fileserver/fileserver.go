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
	"sync"
)

type File interface {
	Read(int64, []byte) (int, error)
	Write(int64, []byte) (int, error)
	Close() error
}

type Cons func() (File, error)

type FileMap map[string]Cons

type FileServer struct {
	neinp.NopP2000
	root File
	fids *fid.Map

	stat map[string]*stat.Stat
	open *sync.Map // fid.Fid → File
	cons FileMap
}

const version = "9P2000"

// TODO: Currently only supports a flat hierarchy with one root
// directory and several regular files located inside this directory.

func NewFileServer(files FileMap) *FileServer {
	fs := &FileServer{
		stat: make(map[string]*stat.Stat),
		open: new(sync.Map),
		cons: files,
		fids: fid.New(),
	}

	rootStat := createStat("/", 0555|stat.Dir)
	rootFile := directory{stat: rootStat}

	for name, _ := range files {
		s := createStat(name, 0644)
		fs.stat[name] = &s
		rootFile.children = append(rootFile.children, s)
	}

	fs.stat[rootStat.Name] = &rootStat
	fs.cons[rootStat.Name] = func() (File, error) { return rootFile, nil }

	return fs
}

func (s *FileServer) Version(ctx context.Context, msg message.TVersion) (message.RVersion, error) {
	if msg.Version != version {
		return message.RVersion{}, errors.New(message.BotchErrorString)
	}

	// TODO: Sanity check on msize
	return message.RVersion{Version: version, Msize: msg.Msize}, nil
}

func (s *FileServer) Attach(ctx context.Context, msg message.TAttach) (message.RAttach, error) {
	stat := s.stat["/"]

	s.fids.Set(msg.Fid, stat)
	return message.RAttach{Qid: stat.Qid}, nil
}

func (s *FileServer) Stat(ctx context.Context, msg message.TStat) (message.RStat, error) {
	stat, ok := s.fids.Get(msg.Fid).(*stat.Stat)
	if !ok {
		return message.RStat{}, errors.New(message.NoStatErrorString)
	}

	return message.RStat{Stat: *stat}, nil
}

func (s *FileServer) Walk(ctx context.Context, msg message.TWalk) (message.RWalk, error) {
	stat, ok := s.fids.Get(msg.Fid).(*stat.Stat)
	if !ok {
		return message.RWalk{}, errors.New(message.UnknownFidErrorString)
	} else if len(msg.Wname) > 1 {
		return message.RWalk{}, errors.New(message.WalkNoDirErrorString)
	}

	newStat := stat
	if len(msg.Wname) == 1 {
		name := msg.Wname[0]
		newStat, ok = s.stat[name]
		if !ok {
			return message.RWalk{}, errors.New(message.NotFoundErrorString)
		}
	}

	if msg.Newfid != msg.Fid {
		if s.fids.Get(msg.Newfid) != nil {
			return message.RWalk{}, errors.New(message.DupFidErrorString)
		}
		s.fids.Set(msg.Newfid, newStat)
	}

	if len(msg.Wname) == 1 {
		return message.RWalk{Wqid: []qid.Qid{newStat.Qid}}, nil
	} else {
		return message.RWalk{}, nil
	}
}

func (s *FileServer) Open(ctx context.Context, msg message.TOpen) (message.ROpen, error) {
	stat, ok := s.fids.Get(msg.Fid).(*stat.Stat)
	if !ok {
		return message.ROpen{}, errors.New(message.UnknownFidErrorString)
	}

	file, err := s.cons[stat.Name]()
	if err != nil {
		return message.ROpen{}, err
	}

	s.open.Store(msg.Fid, file)
	return message.ROpen{Qid: stat.Qid}, nil
}

func (s *FileServer) Read(ctx context.Context, msg message.TRead) (message.RRead, error) {
	f, ok := s.open.Load(msg.Fid)
	if !ok {
		return message.RRead{}, errors.New(message.UnknownFidErrorString)
	}
	file := f.(File)

	// TODO: Sanity check count
	buf := make([]byte, msg.Count)
	n, err := file.Read(int64(msg.Offset), buf)
	if err == io.EOF {
		return message.RRead{Count: 0}, nil
	} else if err != nil {
		return message.RRead{}, err
	}

	return message.RRead{Count: uint32(n), Data: buf[:n]}, nil
}

func (s *FileServer) Write(ctx context.Context, msg message.TWrite) (message.RWrite, error) {
	f, ok := s.open.Load(msg.Fid)
	if !ok {
		return message.RWrite{}, errors.New(message.UnknownFidErrorString)
	}
	file := f.(File)

	n, err := file.Write(int64(msg.Offset), msg.Data)
	if err != nil {
		return message.RWrite{}, err
	}

	return message.RWrite{Count: uint32(n)}, nil
}

func (s *FileServer) Clunk(ctx context.Context, msg message.TClunk) (message.RClunk, error) {
	defer s.fids.Delete(msg.Fid)
	defer s.open.Delete(msg.Fid)

	f, ok := s.open.Load(msg.Fid)
	if ok {
		file := f.(File)
		err := file.Close()
		if err != nil {
			return message.RClunk{}, err
		}
	}

	return message.RClunk{}, nil
}
