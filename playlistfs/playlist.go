package playlistfs

import (
	"fmt"
	"strings"
)

const (
	posFn = iota
	posDesc
)

type Entry struct {
	FileName    string
	Description string
}

func PlaylistEntry(buf []byte) (*Entry, error) {
	var entry Entry

	fields, err := parseFields(buf, 1, -1)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case posFn:
				entry.FileName = data
			case posDesc:
				entry.Description = strings.Join(field[i:], " ")
				return &entry, nil
			}
		}
	}

	return &entry, nil
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s '%s'", e.FileName, e.Description)
}
