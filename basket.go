package jujusvg

import (
	"fmt"
	"image"
	"strconv"
	"strings"

	"gopkg.in/yaml.v1"
)

type bundle struct {
	Relations [][]string
	Services  map[string]struct {
		Annotations map[string]string
		Charm       string
	}
}

// Parse takes a byte array of the bundles.yaml file and converts it to a
// Canvas object.
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
	for serviceName, serviceData := range b.Services {
		x, err := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, err := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		if err != nil {
			return nil, err
		}
		services[serviceName] = &service{
			Point: image.Point{
				X: int(x),
				Y: int(y),
			},
			IconUrl: fmt.Sprintf(
				"https://manage.jujucharms.com/api/3/charm/%s/file/icon.svg",
				strings.Split(serviceData.Charm, ":")[1]),
		}
		canvas.AddService(services[serviceName])
	}
	for _, relation := range b.Relations {
		canvas.AddRelation(&serviceRelation{
			ServiceA: services[strings.Split(relation[0], ":")[0]],
			ServiceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return canvas, nil
}
