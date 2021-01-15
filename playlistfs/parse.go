package playlistfs

import (
	"bufio"
	"bytes"
	"errors"
)

type Fields []string

// TODO: Maybe pass a callback or refactor this into its own type.
func parseFields(buf []byte) ([]Fields, error) {
	lineRdr := bytes.NewBuffer(buf)
	lineScr := bufio.NewScanner(lineRdr)

	var result []Fields
	for lineScr.Scan() {
		fieldRdr := bytes.NewBuffer(lineScr.Bytes())
		fieldScr := bufio.NewScanner(fieldRdr)
		fieldScr.Split(bufio.ScanWords)

		var fields Fields
		for fieldScr.Scan() {
			field := fieldScr.Text()
			fields = append(fields, field)
		}

		if len(fields) == 0 {
			return []Fields{}, errors.New("empty line")
		}

		result = append(result, fields)
	}

	err := lineScr.Err()
	if err != nil {
		return []Fields{}, err
	}

	return result, nil
}
