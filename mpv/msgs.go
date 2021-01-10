package mpv

import (
	"encoding/json"
	"io"
)

const noError = "success"

type msgID int32

type response struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
	ID    msgID       `json:"request_id"`
}

type request struct {
	Cmd []interface{} `json:"command"`
	ID  msgID         `json:"request_id"`
}

func (r *request) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)

	err := enc.Encode(r)
	if err != nil {
		return err
	}

	return nil
}
