jujusvg
=======

A library for generating SVGs from Juju bundles and environments.

Installation
------------

To start using the charm store, first ensure you have a valid Go environment,
then run the following:

    go get github.com/juju/jujusvg

Dependencies
------------

The project uses godeps (https://launchpad.net/godeps) to manage Go
dependencies. To install this, run:


    go get launchpad.net/godeps

After installing it, you can update the dependencies to the revision specified
in the `dependencies.tsv` file with the following:

    make deps

Use `make create-deps` to update the dependencies file.

Usage
-----

Given a Juju bundle, you can convert this to an SVG programatically.  This
generates a simple SVG representation of a bundle or bundles that can then be
included in a webpage as a visualization.

For an example of how to use this library, please see `examples/generatesvg.go`.
You can run this example like:

    go run generatesvg.go bundle.yaml > bundle.svg

The examples directory also includes three sample bundles that you can play
around with, or you can use the [Juju GUI](https://demo.jujucharms.com) to
generate your own bundles.

Design-related assets
---------------------

Some assets are specified based on assets provided by the design team. These
assets are specified in the defs section of the generated SVG, and can thus
be found in the Canvas.definition() method. Should these assets be updated,
the SVGo code will need to be updated to reflect these changes. Unfortunately,
this can only be done by hand, so care must be made to match the SVGs provided
by design exactly.  These original SVG assets live in the `assets` directory.

Current assets in use:

* The service block
* The relation health indicator
