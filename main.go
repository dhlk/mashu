package main

import (
	"flag"
	"log"
)

var (
	catalogPath = flag.String("catalog", "/var/mashu/catalog", "source catalog")
	catalogAlog = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")

	clips = flag.String("clips", "", "path to clip definition stream json; renders clips")
)

func main() {
	flag.Parse()

	if *clips != "" {
		if err := renderClipsFromFile(*clips); err != nil {
			log.Printf("mashu: error rendering clips: %v", err)
		}
		return
	}

	log.Print("nothing to do")
}
