package jujusvg // import "gopkg.in/juju/jujusvg.v1"

import (
	"image"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v6-unstable"
)

// NewFromBundle returns a new Canvas that can be used
// to generate a graphical representation of the given bundle
// data. The iconURL function is used to generate a URL
// that refers to an SVG for the supplied charm URL.
// If fetcher is non-nil, it will be used to fetch icon
// contents for any icons embedded within the charm,
// allowing the generated bundle to be self-contained. If fetcher
// is nil, a default fetcher which refers to icons by their
// URLs as svg <image> tags will be used.
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
		icon := iconMap[charmId.Path()]
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
