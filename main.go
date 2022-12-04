package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

var (
	catalogPath = flag.String("catalog-path", "/var/mashu/catalog", "source catalog")
	catalogAlgo = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")
	catalogMode = flag.Bool("catalog", false, "catalog inputs")
	planMode    = flag.Bool("plan", false, "execute specified plans")
)

func catalogMain(c Catalog, args []string) (err error) {
	for _, arg := range args {
		var s Source
		if s, err = buildSource(arg); err != nil {
			return fmt.Errorf("mashu: error building source for '%s': %w", arg, err)
		}
		if err = c.Create(s); err != nil {
			return fmt.Errorf("mashu: error cataloging source for '%s': %w", arg, err)
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

	if err := projectMain(*catalog, flag.Args()); err != nil {
		log.Fatal(err)
		return
	}
}
