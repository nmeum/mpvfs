package mpv

import (
	"encoding/json"
	"io"
)

const noError = "success"

type response struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
	ReqID int32       `json:"request_id"`

	// Additional fields used by observe events
	ID           int32  `json:"id"`
	Event        string `json:"event"`
	PropertyName string `json:"name"`
}

type request struct {
	Cmd []interface{} `json:"command"`
	ID  int32         `json:"request_id"`
}

func (r *request) Encode(w io.Writer) error {
	enc := json.NewEncoder(w)

	err := enc.Encode(r)
	if err != nil {
		return err
	}

	return nil
}

func (r *request) String() string {
	data, err := json.Marshal(r)
	if err != nil {
		// This function is primarly intended for debugging
		// purposes and therefore panics on error.
		panic(err)
	}

	return string(data)
}
