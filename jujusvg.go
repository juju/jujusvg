package jujusvg

import (
	"image"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/errgo.v1"
	"gopkg.in/juju/charm.v5-unstable"
)

func NewFromBundleEmbedIcons(b *charm.BundleData, iconURL func(*charm.Reference) string) (*Canvas, error) {
	alreadyFetched := make(map[string]bool)
	icons := make(map[string]string)
	for _, serviceData := range b.Services {
		charmId, err := charm.ParseReference(serviceData.Charm)
		if err != nil {
			return nil, errgo.Notef(err, "cannot parse charm %q", serviceData.Charm)
		}
		if _, ok := alreadyFetched[charmId.Path()]; ok == false {
			alreadyFetched[charmId.Path()] = true
			icon, err := fetchIcon(iconURL(charmId))
			if err != nil {
				return nil, err
			}
			icons[charmId.Path()] = icon
		}
	}
	return NewFromBundleWithMap(b, iconURL, icons)
}

func fetchIcon(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errgo.Newf("URL %s was not valid", url)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errgo.Newf("could not read icon data from url %s", url)
	}
	return string(body), nil
}

func NewFromBundle(b *charm.BundleData, iconURL func(*charm.Reference) string) (*Canvas, error) {
	return NewFromBundleWithMap(b, iconURL, map[string]string{})
}

// NewFromBundleWithMap returns a new Canvas that can be used
// to generate a graphical representation of the given bundle
// data. The iconURL function is used to generate a URL
// that refers to an SVG for the supplied charm URL. If a map
// of charms to icon SVGs is provided, then those SVGs will be
// embedded in the bundle diagram and used instead of an image
// tag.
func NewFromBundleWithMap(b *charm.BundleData, iconURL func(*charm.Reference) string, iconMap map[string]string) (*Canvas, error) {
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
