package jujusvg

import (
	"fmt"
	"image"
	"strconv"
	"sort"
	"strings"

	"gopkg.in/errgo.v1"
	"gopkg.in/yaml.v1"
)

type bundle struct {
	Relations [][]string
	Services  map[string]struct {
		Annotations map[string]string
		Charm       string
	}
}

// parseBasket parses the contents of a bundles.yaml file and returns
// a map with one Canvas for each bundle inside it.
func parseBasket(basketData []byte) (map[string]*Canvas, error) {
	basket := make(map[string]bundle)
	canvases := make(map[string]*Canvas)
	err := yaml.Unmarshal(basketData, &basket)
	if err != nil {
		return nil, err
	}
	for name, b := range basket {
		canvases[name], err = parseBasketBundle(b)
		if err != nil {
			return nil, err
		}
	}
	return canvases, nil
}

// parseBasketBundle creates the actual Canvas element from the parsed YAML.
func parseBasketBundle(b bundle) (*Canvas, error) {
	canvas := &Canvas{}
	services := make(map[string]*service)

	// Go through all services in alphabetical order so that
	// we get consistent results.
	serviceNames := make([]string, 0, len(b.Services))
	for name := range b.Services {
		serviceNames = append(serviceNames, name)
	}
	sort.Strings(serviceNames)
	for _, serviceName := range serviceNames {
		serviceData := b.Services[serviceName]
		x, xerr := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, yerr := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		if xerr != nil || yerr != nil {
			return nil, errgo.Newf("service %q does not have a valid position", serviceName)
		}
		services[serviceName] = &service{
			point: image.Point{
				X: int(x),
				Y: int(y),
			},
			iconUrl: fmt.Sprintf(
				"https://manage.jujucharms.com/api/3/charm/%s/file/icon.svg",
				strings.Split(serviceData.Charm, ":")[1]),
		}
		canvas.addService(services[serviceName])
	}
	for _, relation := range b.Relations {
		canvas.addRelation(&serviceRelation{
			serviceA: services[strings.Split(relation[0], ":")[0]],
			serviceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return canvas, nil
}
