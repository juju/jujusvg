package parsers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/makyo/jujusvg"
	"gopkg.in/yaml.v1"
)

type Bundle struct {
	Relations [][]string
	Services  map[string]struct {
		Annotations map[string]string
		Charm       string
	}
}

type BundleParser struct {
	Parser
}

// Parse takes a byte array of the bundles.yaml file and converts it to a
// Canvas object.
func (r *BundleParser) Parse(bundleData []byte) (map[string]jujusvg.Canvas, error) {
	basket := make(map[string]Bundle)
	canvases := make(map[string]jujusvg.Canvas)
	err := yaml.Unmarshal(bundleData, &basket)
	if err != nil {
		return nil, err
	}
	for bundleName, bundle := range basket {
		canvases[bundleName] = r.parseBundle(bundle)
	}
	return canvases, nil
}

// parseBundle creates the actual Canvas element from the parsed YAML.
func (r *BundleParser) parseBundle(bundle Bundle) jujusvg.Canvas {
	canvas := jujusvg.Canvas{}
	services := make(map[string]*jujusvg.Service)
	for serviceName, serviceData := range bundle.Services {
		x, _ := strconv.ParseFloat(serviceData.Annotations["gui-x"], 64)
		y, _ := strconv.ParseFloat(serviceData.Annotations["gui-y"], 64)
		services[serviceName] = &jujusvg.Service{
			X:        int(x),
			Y:        int(y),
			CharmUrl: serviceData.Charm,
			IconUrl: fmt.Sprintf(
				"https://manage.jujucharms.com/api/3/charm/%s/file/icon.svg",
				strings.Split(serviceData.Charm, ":")[1]),
		}
		canvas.AddService(services[serviceName])
	}
	for _, relation := range bundle.Relations {
		canvas.AddRelation(&jujusvg.Relation{
			ServiceA: services[strings.Split(relation[0], ":")[0]],
			ServiceB: services[strings.Split(relation[1], ":")[0]],
		})
	}
	return canvas
}
