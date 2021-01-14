package playlistfs

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNoCmd = errors.New("not a playlistfs command")
)

const (
	posCmd = iota
	posName
	posArg
)

type Command struct {
	Name string
	Arg  uint
}

func CtlCmd(buf []byte) (*Command, error) {
	var cmd Command

	fields, err := parseFields(buf, 2, 3)
	if err != nil {
		return nil, err
	}

	for _, field := range fields {
		for i := 0; i < len(field); i++ {
			data := field[i]
			switch i {
			case posCmd:
				if data != "cmd" {
					return nil, ErrNoCmd
				}
			case posName:
				cmd.Name = data
			case posArg:
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

func (c *Command) String() string {
	return fmt.Sprintf("cmd %s %d", c.Name, c.Arg)
}
