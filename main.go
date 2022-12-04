package main

import (
	"flag"
	"log"
)

var (
	catalogPath = flag.String("catalog-path", "/var/mashu/catalog", "source catalog")
	catalogAlgo = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")
	catalogMode = flag.Bool("catalog", false, "catalog inputs")
)

func main() {
	flag.Parse()

	catalog, err := NewCatalog(*catalogPath, *catalogAlgo)
	if err != nil {
		log.Fatal(err)
		return
	}

	if *catalogMode {
		for _, arg := range flag.Args() {
			var s Source
			var err error
			if s, err = buildSource(arg); err != nil {
				log.Fatalf("mashu: error building source for '%s': %v", arg, err)
				return
			}
			if err = catalog.Create(s); err != nil {
				log.Fatalf("mashu: error cataloging source for '%s': %v", arg, err)
				return
			}
		}
		return
	}

	for _, arg := range flag.Args() {
		project, err := NewProject(arg, *catalog)
		if err != nil {
			log.Fatal(err)
			return
		}

		if err := project.Execute(); err != nil {
			log.Fatal(err)
			return
		}
	}
}
