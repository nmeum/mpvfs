package mpv

import (
	"encoding/json"
	"io"
)

type msgID int32

type message struct {
	command
	response
}

func (m message) isCommand() bool {
	return len(m.Cmd) != 0
}

func (m message) isResponse() bool {
	return len(m.Error) != 0
}

type response struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
	ID    msgID       `json:"request_id"`
}

type command struct {
	Cmd []string `json:"command"`
	ID  msgID    `json:"request_id"`
}

func (c *command) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)

	err := enc.Encode(c)
	if err != nil {
		return err
	}

	return nil
}
