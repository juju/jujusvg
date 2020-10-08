package main

// This is a demo application that uses the jujusvg library to build a bundle SVG
// from a given bundle.yaml file.

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/juju/charm/v8"

	// Import the jujusvg library and the juju charm library
	"github.com/juju/jujusvg/v5"
)

// iconURL takes a reference to a charm and returns the URL for that charm's icon.
// In this case, we're using the api.jujucharms.com API to provide the icon's URL.
func iconURL(ctx context.Context, ref *charm.URL) (string, error) {
	return "https://api.jujucharms.com/charmstore/v5/" + ref.Path() + "/icon.svg", nil
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Please provide the name of a bundle file as the first argument")
	}

	ctx := context.Background()

	// First, we need to read our bundle data into a []byte
	bundle_data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error reading bundle: %s\n", err)
	}

	// Next, generate a charm.Bundle from the bytearray by passing it to ReadNewBundleData.
	// This gives us an in-memory object representation of the bundle that we can pass to jujusvg
	bundle, err := charm.ReadBundleData(strings.NewReader(string(bundle_data)))
	if err != nil {
		log.Fatalf("Error parsing bundle: %s\n", err)
	}

	fetcher := &jujusvg.HTTPFetcher{
		IconURL: iconURL,
	}
	// Next, build a canvas of the bundle.  This is a simplified version of a charm.Bundle
	// that contains just the position information and charm icon URLs necessary to build
	// the SVG representation of the bundle
	canvas, err := jujusvg.NewFromBundle(ctx, bundle, iconURL, fetcher)
	if err != nil {
		log.Fatalf("Error generating canvas: %s\n", err)
	}

	// Finally, marshal that canvas as SVG to os.Stdout; this will print the SVG data
	// required to generate an image of the bundle.
	canvas.Marshal(os.Stdout)
}
