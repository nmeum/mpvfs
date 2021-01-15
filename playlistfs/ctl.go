package playlistfs

import (
	"fmt"
	"strconv"
)

type Control struct {
	Name string
	Arg  uint
}

func CtlCmd(buf []byte) (*Control, error) {
	var cmd Control

	fields, err := parseFields(buf, 2, 3)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case 0:
				if data != "cmd" {
					return nil, ErrNoCtl
				}
			case 1:
				cmd.Name = data
			case 2:
				arg, err := strconv.ParseUint(data, 10, 32)
				if err != nil {
					return nil, err
				}
				cmd.Arg = uint(arg)
			}
		}
	}

	return &cmd, nil
}

func (c *Control) String() string {
	return fmt.Sprintf("cmd %s %d", c.Name, c.Arg)
}
