package playlistfs

import (
	"bufio"
	"bytes"
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
	reader := bytes.NewBuffer(buf)

	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)

	var cmd Command
	for i := 0; scanner.Scan(); i++ {
		data := scanner.Text()
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

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}

func (c *Command) String() string {
	return fmt.Sprintf("cmd %s %d", c.Name, c.Arg)
}
