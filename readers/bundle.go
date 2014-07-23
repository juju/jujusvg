package readers

import (
	"fmt"
	"strings"

	"github.com/makyo/jujusvg"
	"gopkg.in/yaml.v1"
)

type Bundle struct {
	Relations [][]string
	Services  map[string]struct {
		Annotations map[string]int
		Charm       string
	}
}

type Basket struct {
	Bundles map[string]Bundle
}

type BundleReader struct {
	Reader
}

func (r *BundleReader) Read(b []byte) (map[string]jujusvg.Canvas, error) {
	basket := Basket{}
	canvases := make(map[string]jujusvg.Canvas)
	err := yaml.Unmarshal(b, &basket)
	if err != nil {
		return nil, err
	}
	for bundleName, bundle := range basket.Bundles {
		canvases[bundleName] = r.parseBundle(bundle)
	}
	return canvases, nil
}

func (r *BundleReader) parseBundle(bundle Bundle) jujusvg.Canvas {
	canvas := jujusvg.Canvas{}
	services := make(map[string]*jujusvg.Service)
	for serviceName, serviceData := range bundle.Services {
		services[serviceName] = &jujusvg.Service{
			X:        serviceData.Annotations["gui-x"],
			Y:        serviceData.Annotations["gui-y"],
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
