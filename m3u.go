package main

import (
	"bufio"
	"fmt"
	"os"
)

var ErrNoM3UHeader = fmt.Errorf("not an m3u file")

func getM3uEntries(name string) (e []string, err error) {
	var f *os.File
	if f, err = os.Open(name); err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	if !s.Scan() || s.Text() != "#EXTM3U" {
		err = ErrNoM3UHeader
		return
	}

	for s.Scan() {
		l := s.Text()
		if l[0] == '#' {
			continue
		}
		e = append(e, l)
	}

	err = s.Err()
	return
}
