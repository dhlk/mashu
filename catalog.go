package main

import (
	"context"
	"encoding/json"
	"fmt"
	"fsmap"
	"math/rand"
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

func (c Catalog) Keys(ctx context.Context) (keys []string, err error) {
	var f *os.File
	if f, err = os.Open(filepath.Join(c.path, "keys")); err != nil {
		return
	}
	defer f.Close()

	d := json.NewDecoder(f)
	for d.More() {
		var s string
		if err = d.Decode(&s); err != nil {
			err = fmt.Errorf("mashu.Catalog.Keys: error decoding keys: %w", err)
			return
		}
		select {
		case <-ctx.Done():
			err = context.Canceled
			return
		default:
			keys = append(keys, s)
		}
	}

	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
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
