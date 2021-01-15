package playlistfs

import (
	"bufio"
	"bytes"
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

		err := fieldScr.Err()
		if err != nil {
			return []Fields{}, err
		}

		numFields := len(fields)
		if numFields < min {
			return []Fields{}, ErrTooFewArgs
		} else if max != -1 && numFields > max {
			return []Fields{}, ErrTooManyArgs
		}

		result = append(result, fields)
	}

	err := lineScr.Err()
	if err != nil {
		return []Fields{}, err
	}

	return result, nil
}
