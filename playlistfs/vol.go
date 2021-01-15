package playlistfs

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	posVolCmd = iota
	posVol
)

type Volume struct {
	Levels []uint
}

func VolCmd(buf []byte) (*Volume, error) {
	var vol Volume

	fields, err := parseFields(buf)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		var i int
		for i = 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case posVolCmd:
				if data != "vol" {
					return nil, ErrNoCmd
				}
			case posVol:
				lvl, err := strconv.ParseUint(field[i], 10, 8)
				if err != nil {
					return nil, err
				} else if lvl > 100 {
					return nil, errors.New("invalid volume level")
				}

				vol.Levels = append(vol.Levels, uint(lvl))
			}
		}

		if i != 2 {
			return nil, errors.New("insufficient amount of fields")
		}
	}

	return &vol, nil
}

func (v *Volume) String() string {
	numLvls := len(v.Levels)
	if numLvls == 1 {
		return fmt.Sprintf("vol %d", v.Levels[0])
	}

	var lvlSet string
	for i, lvl := range v.Levels {
		lvlSet += fmt.Sprintf("%d", lvl)
		if i != numLvls {
			lvlSet += " "
		}
	}

	return fmt.Sprintf("vol \"%s\"", lvlSet)
}
