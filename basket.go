package jujusvg

import (
	"fmt"
	"image"
	"strconv"
	"strings"

	"gopkg.in/yaml.v1"
)

type basketBundle struct {
	Relations [][]string
	Services  map[string]struct {
		Annotations map[string]string
		Charm       string
	}
}

type basketParser struct{}

// Parse takes a byte array of the bundles.yaml file and converts it to a
// Canvas object.
func (r *basketParser) Parse(basketData []byte) (map[string]*Canvas, error) {
	basket := make(map[string]basketBundle)
	canvases := make(map[string]*Canvas)
	err := yaml.Unmarshal(basketData, &basket)
	if err != nil {
		return nil, err
	}
	for bundleName, bundle := range basket {
		canvases[bundleName] = r.parseBasketBundle(bundle)
	}
	return canvases, nil
}

// parseBasketBundle creates the actual Canvas element from the parsed YAML.
func (r *basketParser) parseBasketBundle(bundle basketBundle) *Canvas {
	canvas := &Canvas{}
	services := make(map[string]*service)
	for serviceName, serviceData := range bundle.Services {
		x, _ := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, _ := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		services[serviceName] = &service{
			Point: image.Point{
				X: int(x),
				Y: int(y),
			},
			CharmUrl: serviceData.Charm,
			IconUrl: fmt.Sprintf(
				"https://manage.jujucharms.com/api/3/charm/%s/file/icon.svg",
				strings.Split(serviceData.Charm, ":")[1]),
		}
		canvas.AddService(services[serviceName])
	}
	for _, relation := range bundle.Relations {
		canvas.AddRelation(&serviceRelation{
			ServiceA: services[strings.Split(relation[0], ":")[0]],
			ServiceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return canvas
}
