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

	if err = decodeJsonFromFile(filepath.Join(path, "source.json"), &s); err != nil {
		return
	}

	return
}

func (c Catalog) Create(s Source) (err error) {
	var path string
	if path, err = c.fsmap.Lookup([]byte(s.Key), true); err != nil {
		return
	}

	sourcePath := Output(filepath.Join(path, "source.json"))
	if err = sourcePath.Valid(); err != nil {
		return fmt.Errorf("mashu.Catalog.Create: for key '%s': %w", s.Key, err)
	}

	var f *os.File
	if f, err = os.Create(string(sourcePath)); err != nil {
		return
	}
	defer f.Close()

	e := json.NewEncoder(f)
	if err = e.Encode(s); err != nil {
		os.Remove(f.Name())
		return fmt.Errorf("mashu.Catalog.Create: unable to encode source: %w", err)
	}

	var k *os.File
	if k, err = os.OpenFile(filepath.Join(c.path, "keys"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		return
	}
	defer k.Close()

	ke := json.NewEncoder(k)
	if err = ke.Encode(s.Key); err != nil {
		return fmt.Errorf("mashu.Catalog.Create: unable to encode key: %w", err)
	}

	return
}
