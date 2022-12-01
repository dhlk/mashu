package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fsmap"
	"log"
	"os"
	"path/filepath"
)

type Catalog struct {
	path  string
	fsmap *fsmap.Fsmap
}

func NewCatalog(path, algorithm string) (c *Catalog, err error) {
	c = &Catalog{path: path}
	c.fsmap, err = fsmap.New(path, algorithm)
	return
}

func (c Catalog) Keys(ctx context.Context) (o <-chan string, err error) {
	var f *os.File
	if f, err = os.Open(filepath.Join(c.path, "keys")); err != nil {
		return
	}

	ch := make(chan string)
	d := json.NewDecoder(f)
	go func() {
		defer f.Close()
		defer close(ch)

		for d.More() {
			var s string
			if err := d.Decode(&s); err != nil {
				log.Printf("mashu.Catalog.Keys: error decoding keys: %v", err)
				return
			}
			select {
			case <-ctx.Done():
				return
			case ch <- s:
			}
		}
	}()

	o = ch
	return
}

func (c Catalog) Lookup(key string) (s Source, err error) {
	var path string
	if path, err = c.fsmap.Lookup([]byte(key), false); err != nil {
		return
	}

	var f *os.File
	if f, err = os.Open(filepath.Join(path, "source.json")); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	err = d.Decode(&s)
	return
}

func (c Catalog) Create(s Source) (err error) {
	var path string
	if path, err = c.fsmap.Lookup([]byte(s.Key), true); err != nil {
		return
	}

	if _, err = os.Stat(filepath.Join(path, "source.json")); !os.IsNotExist(err) {
		if err != nil {
			return err
		} else {
			return fmt.Errorf("mashu.Catalog.Create: for key '%s': %w", s.Key, os.ErrExist)
		}
	}

	var f *os.File
	if f, err = os.Create(filepath.Join(path, "source.json")); err != nil {
		return
	}
	defer f.Close()

	e := json.NewEncoder(f)
	err = e.Encode(s)
	return
}
