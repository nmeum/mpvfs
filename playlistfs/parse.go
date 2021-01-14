package playlistfs

import (
	"bufio"
	"bytes"
	"errors"
)

type Fields []string

func parseFields(buf []byte, min int, max int) ([]Fields, error) {
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

		numFields := len(fields)
		if numFields < min {
			return []Fields{}, errors.New("below minimum")
		} else if max != -1 && numFields > max {
			return []Fields{}, errors.New("above maximum")
		}

		result = append(result, fields)
	}

	err := lineScr.Err()
	if err != nil {
		return []Fields{}, err
	}

	return result, nil
}
