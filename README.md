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
        "log"
        "os"

        "github.com/makyo/jujusvg"
)

func main() {
        basket, err := ioutil.ReadFile("bundles.yaml")
        if err != nil {
                log.Fatalf("Error reading file: %s\n", err)
        }
        canvases, err := jujusvg.NewFromBasket(basket)
        if err != nil {
                log.Fatalf("Error reading basket: %s\n", err)
        }
        for canvasName, canvas := range canvases {
                fmt.Printf("Found bundle: %s\n", canvasName)
                canvas.Marshal(os.Stdout)
        }
}
```

This generates a simple SVG representation of a bundle or bundles that can then
be included in a webpage as a visualization.
