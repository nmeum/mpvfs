package playlistfs

import (
	"errors"
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

	fields, err := parseFields(buf)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		var i int
		for i = 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case posFn:
				entry.FileName = data
			case posDesc:
				entry.Description = strings.Join(field[i:], " ")
			}
		}

		if i < 1 || i > 2 {
			return nil, errors.New("insufficient amount of fields")
		}
	}

	return &entry, nil
}

func (e *Entry) String() string {
	return fmt.Sprintf("%s '%s'", e.FileName, e.Description)
}
