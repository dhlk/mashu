package main

import (
	"encoding/json"
	"os"
)

func decodeJsonFromFile(path string, o ...any) (err error) {
	var f *os.File
	if f, err = os.Open(path); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for _, x := range o {
		if err = d.Decode(x); err != nil {
			return
		}
	}

	return f.Close()
}
