jujusvg
=======

A library for generating SVGs from Juju bundles and environments.

Usage
-----

Given the YAML specification of a basket (a series of bundles, in short), you
can convert this to an SVG like so:

```go
package main

// This is a demo application that uses the jujusvg library to build a bundle SVG
// from a given bundle.yaml file.

import (
        "io/ioutil"
        "log"
        "os"
        "strings"

        // Import the jujusvg library and the juju charm library
        "github.com/juju/jujusvg"
        "gopkg.in/juju/charm.v4"
)

// iconURL takes a reference to a charm and returns the URL for that charm's icon.
// In this case, we're using the manage.jujucharms.com API to provide the icon's URL.
func iconURL(ref *charm.Reference) string {
        return "https://manage.jujucharms.com/api/3/charm/" + ref.Path() + "/file/icon.svg"
}

func main() {
        // First, we need to read our bundle data into a []byte
        bundle_data, err := ioutil.ReadFile("bundle.yaml")
        if err != nil {
                log.Fatalf("Error reading bundle: %s\n", err)
        }

        // Next, generate a charm.Bundle from the bytearray by passing it to ReadNewBundleData.
        // This gives us an in-memory object representation of the bundle that we can pass to jujusvg
        bundle, err := charm.ReadBundleData(strings.NewReader(string(bundle_data)))
        if err != nil {
                log.Fatalf("Error parsing bundle: %s\n", err)
        }

        // Next, build a canvas of the bundle.  This is a simplified version of a charm.Bundle
        // that contains just the position information and charm icon URLs necessary to build
        // the SVG representation of the bundle
        canvas, err := jujusvg.NewFromBundle(bundle, iconURL)
        if err != nil {
                log.Fatalf("Error generating canvas: %s\n", err)
        }

        // Finally, marshal that canvas as SVG to os.Stdout; this will print the SVG data
        // required to generate an image of the bundle.
        canvas.Marshal(os.Stdout)
}

```

This generates a simple SVG representation of a bundle or bundles that can then
be included in a webpage as a visualization.
