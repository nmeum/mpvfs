package playlistfs

import (
	"bufio"
	"bytes"
	"errors"
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
	Arg  int
}

func ParseCtlCmd(buf []byte) (*Command, error) {
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
			var err error
			cmd.Arg, err = strconv.Atoi(data)
			if err != nil {
				return nil, err
			}
		}
	}

	err := scanner.Err()
	if err != nil {
		return nil, err
	}

	return &cmd, nil
}
