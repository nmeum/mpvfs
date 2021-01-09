package main

import (
	"hash/fnv"
)

func hashPath(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
