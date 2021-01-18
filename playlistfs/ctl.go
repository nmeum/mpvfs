package playlistfs

import (
	"fmt"
	"strconv"
)

type Control struct {
	Name string
	Arg  *uint
}

func CtlCmd(buf []byte) (*Control, error) {
	var cmd Control

	fields, err := parseFields(buf, 1, 2)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case 0:
				cmd.Name = data
			case 1:
				arg, err := strconv.ParseUint(data, 10, 32)
				if err != nil {
					return nil, err
				}

				uarg := uint(arg)
				cmd.Arg = &uarg
			}
		}
	}

	return &cmd, nil
}

func (c *Control) String() string {
	return fmt.Sprintf("%s %d", c.Name, *c.Arg)
}
