package jujusvg

import (
	"errors"

	"gopkg.in/juju/charm.v4"
)

// NewFromBasket generates canvases for each bundle in a basket, mapped to
// the name provided in the basket
func NewFromBasket(basket []byte) (map[string]*Canvas, error) {
	return parseBasket(basket)
}

func NewFromBundle(bundle *charm.BundleData) (*Canvas, error) {
	return nil, errors.New("Not implemented yet.")
}
