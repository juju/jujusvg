package jujusvg

import (
	"image"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v5"
)

// NewFromBundle returns a new Canvas that can be used
// to generate a graphical representation of the given bundle
// data. The iconURL function is used to generate a URL
// that refers to an SVG for the supplied charm URL.  If
// requested, icon SVGs that are available will be embedded
// within the bundle SVG so that the bundle will be self-
// contained as much as possible.  If that is to be the case,
// utilize the provided IconFetcher (or default to HTTPFetcher
// if nil) to retrieve the icons.
func NewFromBundle(b *charm.BundleData, iconURL func(*charm.Reference) string, fetcher IconFetcher) (*Canvas, error) {
	if fetcher == nil {
		fetcher = &LinkFetcher{
			IconURL: iconURL,
		}
	}
	iconMap, err := fetcher.FetchIcons(b)
	if err != nil {
		return nil, err
	}
	return newFromBundleWithMap(b, iconURL, iconMap)
}

// newFromBundleWithMap returns a new Canvas that can be used
// to generate a graphical representation of the given bundle
// data. The iconURL function is used to generate a URL
// that refers to an SVG for the supplied charm URL. If a map
// of charms to icon SVGs is provided, then those SVGs will be
// embedded in the bundle diagram and used instead of an image
// tag.
func newFromBundleWithMap(b *charm.BundleData, iconURL func(*charm.Reference) string, iconMap map[string][]byte) (*Canvas, error) {
	var canvas Canvas

	// Verify the bundle to make sure that all the invariants
	// that we depend on below actually hold true.
	if err := b.Verify(nil); err != nil {
		return nil, errgo.Notef(err, "cannot verify bundle")
	}
	// Go through all services in alphabetical order so that
	// we get consistent results.
	serviceNames := make([]string, 0, len(b.Services))
	for name := range b.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	services := make(map[string]*service)
	for _, name := range serviceNames {
		serviceData := b.Services[name]
		x, xerr := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, yerr := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		if xerr != nil || yerr != nil {
			return nil, errgo.Newf("service %q does not have a valid position", name)
		}
		charmId, err := charm.ParseReference(serviceData.Charm)
		if err != nil {
			// cannot actually happen, as we've verified it.
			return nil, errgo.Notef(err, "cannot parse charm %q", serviceData.Charm)
		}
		icon, _ := iconMap[charmId.Path()]
		svc := &service{
			name:      name,
			charmPath: charmId.Path(),
			point:     image.Point{int(x), int(y)},
			iconUrl:   iconURL(charmId),
			iconSrc:   icon,
		}
		services[name] = svc
		canvas.addService(svc)
	}
	for _, relation := range b.Relations {
		canvas.addRelation(&serviceRelation{
			serviceA: services[strings.Split(relation[0], ":")[0]],
			serviceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return &canvas, nil
}
