jujusvg
=======

A library for generating SVGs from Juju bundles and environments.

Usage
-----

Given the YAML specification of a basket (a series of bundles, in short), you
can convert this to an SVG like so:

```go
package main

import (
		"fmt"
		"io/ioutil"
        "os"

		"github.com/juju/jujusvg/parsers"
)

func main() {
	basket, err := ioutil.ReadFile("bundles.yaml")
	if err != nil {
		fmt.Printf("Error reading file: %s\n", err)
		return
	}
	bundleParser := parsers.BundleParser{}
	canvases, err := bundleParser.Parse(basket)
	if err != nil {
		fmt.Printf("Error reading basket: %s\n", err)
		return
	}
	for canvasName, canvas := range canvases {
		fmt.Printf("Found bundle: %s\n", canvasName)
		canvas.Marshal(os.Stdout)
	}
}
```

This generates a simple SVG representation of a bundle or bundles that can then
be included in a webpage as a visualization.
