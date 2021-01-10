package fileserver

import (
	"go.rbn.im/neinp/qid"
	"go.rbn.im/neinp/stat"

	"hash/fnv"
	"time"
)

func createStat(name string, mode stat.Mode) stat.Stat {
	now := time.Now()

	var qtype qid.Type
	if mode&stat.Dir != 0 {
		qtype = qid.TypeDir
	} else {
		qtype = qid.TypeFile
	}

	q := qid.Qid{Type: qtype, Version: 0, Path: hashPath(name)}
	s := stat.Stat{
		Qid:    q,
		Mode:   mode,
		Atime:  now,
		Mtime:  now,
		Length: 0,
		Name:   name,
	}

	return s
}

func hashPath(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
