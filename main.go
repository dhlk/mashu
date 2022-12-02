package main

import (
	"context"
	"flag"
	"log"
)

var (
	catalogPath = flag.String("catalog", "/var/mashu/catalog", "source catalog")
	catalogAlgo = flag.String("catalog-algorithm", "SHA-512", "source catalog algorithm")

	stream = flag.String("stream", "", "path to json stream of definitions to render")
)

func main() {
	flag.Parse()

	if *stream != "" {
		if err := renderStreamFromFile(context.TODO(), *stream); err != nil {
			log.Printf("mashu: error rendering stream: %v", err)
		}
		return
	}

	log.Print("nothing to do")
}
