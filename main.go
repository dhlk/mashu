package main

import (
	"flag"
	"log"
)

var (
	catalogPath = flag.String("catalog", "/var/mashu/catalog", "source catalog")
	catalogAlog = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")

	clips  = flag.String("clips", "", "path to clip definition stream json; renders clips")
	stacks = flag.String("stacks", "", "path to stack definition stream json; renders stacks")
)

func main() {
	flag.Parse()

	if *clips != "" {
		if err := renderClipsFromFile(*clips); err != nil {
			log.Printf("mashu: error rendering clips: %v", err)
		}
		return
	}
	if *stacks != "" {
		if err := renderStacksFromFile(*stacks); err != nil {
			log.Printf("mashu: error rendering stacks: %v", err)
		}
		return
	}

	log.Print("nothing to do")
}
