package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

var (
	catalogPath = flag.String("catalog-path", "/var/mashu/catalog", "source catalog")
	catalogAlgo = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")
	catalogMode = flag.Bool("catalog", false, "catalog inputs")
	planMode    = flag.Bool("plan", false, "execute specified plans")
	genMode     = flag.Bool("generate", false, "generate a plans for the specified projects")
)

func catalogMain(c Catalog, args []string) (err error) {
	targets := make([]string, 0)

	for _, arg := range args {
		if _, err = c.Lookup(arg); errors.Is(err, fs.ErrNotExist) {
			err = nil
		} else {
			err = nil
			continue
		}

		targets = append(targets, arg)
	}

	if len(targets) == 0 {
		return
	}

	var sources []Source
	if sources, err = buildSources(targets...); err != nil {
		return fmt.Errorf("mashu: error building sources: %w", err)
	}
	for i, s := range sources {
		if err = c.Create(s); err != nil {
			return fmt.Errorf("mashu: error cataloging source for '%s': %w", targets[i], err)
		}
	}

	return
}

func planMain(c Catalog, args []string) error {
	for _, arg := range args {
		projectDir := filepath.Dir(filepath.Dir(arg))
		project, err := NewProject(projectDir, c)
		if err != nil {
			return err
		}

		if err := project.executePlanByName(strings.TrimSuffix(filepath.Base(arg), ".json")); err != nil {
			return err
		}
	}

	return nil
}

func genMain(c Catalog, args []string) error {
	for _, arg := range args {
		project, err := NewProject(arg, c)
		if err != nil {
			return err
		}

		if err := project.Generate(); err != nil {
			return err
		}
	}

	return nil
}

func projectMain(c Catalog, args []string) error {
	for _, arg := range args {
		project, err := NewProject(arg, c)
		if err != nil {
			return err
		}

		if err := project.Execute(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	catalog, err := NewCatalog(*catalogPath, *catalogAlgo)
	if err != nil {
		log.Fatal(err)
		return
	}

	if *catalogMode {
		if err := catalogMain(*catalog, flag.Args()); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *planMode {
		if err := planMain(*catalog, flag.Args()); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *genMode {
		if err := genMain(*catalog, flag.Args()); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := projectMain(*catalog, flag.Args()); err != nil {
		log.Fatal(err)
		return
	}
}
