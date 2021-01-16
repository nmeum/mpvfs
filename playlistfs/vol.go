package playlistfs

import (
	"fmt"
	"strconv"
)

type Volume struct {
	Levels []uint
}

func VolCmd(buf []byte) (*Volume, error) {
	var vol Volume

	fields, err := parseFields(buf, 2, -1)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case 0:
				if data != "volume" {
					return nil, ErrNoVol
				}
			default:
				lvl, err := strconv.ParseUint(field[i], 10, 8)
				if err != nil {
					return nil, err
				} else if lvl > 100 {
					return nil, ErrInvalidVol
				}

				vol.Levels = append(vol.Levels, uint(lvl))
			}
		}
	}

	return &vol, nil
}

func (v *Volume) String() string {
	numLvls := len(v.Levels)
	if numLvls == 1 {
		return fmt.Sprintf("volume %d", v.Levels[0])
	}

	var lvlSet string
	for i, lvl := range v.Levels {
		lvlSet += fmt.Sprintf("%d", lvl)
		if i != numLvls {
			lvlSet += " "
		}
	}

	return fmt.Sprintf("volume \"%s\"", lvlSet)
}
