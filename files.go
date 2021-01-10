package main

import (
	"fmt"
	"os"
)

type ControlFile struct {
	Name string
}

func (c ControlFile) Read(int64, []byte) (int, error) {
	fmt.Println("Reading", c.Name)
	return 0, nil
}

func (c ControlFile) Write(int64, []byte) (int, error) {
	fmt.Println("Writing", c.Name)
	return 0, os.ErrInvalid
}
