package playlistfs

import (
	"fmt"
	"strings"
)

type Playlist struct {
	FileName    string
	Description string
}

func PlaylistCmd(buf []byte) (*Playlist, error) {
	var entry Playlist

	fields, err := parseFields(buf, 1, -1)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case 0:
				entry.FileName = data
			case 1:
				entry.Description = strings.Join(field[i:], " ")
				return &entry, nil
			}
		}
	}

	return &entry, nil
}

func (p *Playlist) String() string {
	return fmt.Sprintf("%s '%s'", p.FileName, p.Description)
}
