package main

import (
	"flag"
	"log"
)

var (
	catalogPath = flag.String("catalog", "/var/mashu/catalog", "source catalog")
	catalogAlgo = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")
)

func main() {
	flag.Parse()

	catalog, err := NewCatalog(*catalogPath, *catalogAlgo)
	if err != nil {
		log.Fatal(err)
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
