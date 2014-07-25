package jujusvg

import (
	"io"
	"math"

	"github.com/ajstarks/svgo"
)

const (
	// IconSize is always 96px per Juju charm conventions.
	IconSize = 96
)

// Canvas contains a list of services and a list of relations which can be
// rendered to SVG.
type Canvas struct {
	services  []*Service
	relations []*Relation
}

// Service represents a service deployed to an environment and contains its
// position information, icon URL, and additional metadata.
type Service struct {
	Name     string
	CharmUrl string
	IconUrl  string
	X        int
	Y        int
}

// Relation represents a relation created between two services.
type Relation struct {
	ServiceA *Service
	ServiceB *Service
}

// definition creates any necessary defs that can be used later in the SVG.
func (s *Service) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the service in the SVG.
func (s *Service) usage(canvas *svg.SVG) {
	canvas.Image(s.X, s.Y, IconSize, IconSize, s.IconUrl)
}

// definition creates any necessary defs that can be used later in the SVG.
func (r *Relation) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the relation in the SVG.
func (r *Relation) usage(canvas *svg.SVG) {
	canvas.Line(
		r.ServiceA.X+(IconSize/2),
		r.ServiceA.Y+(IconSize/2),
		r.ServiceB.X+(IconSize/2),
		r.ServiceB.Y+(IconSize/2),
		"stroke:black")
}

// AddService adds a new service to the canvas.
func (c *Canvas) AddService(s *Service) {
	c.services = append(c.services, s)
}

// AddRelation adds a new relation to the canvas.
func (c *Canvas) AddRelation(r *Relation) {
	c.relations = append(c.relations, r)
}

// getRect retrieves the width and height of the canvas, as well as modifying
// the coordinates of the services to ensure that everything is positioned with
// (0, 0) as the minimum coordinates.
func (c *Canvas) getRect() (int, int) {
	minWidth := int(^uint(0) >> 1)
	minHeight := int(^uint(0) >> 1)
	maxWidth := -(minWidth - 1)
	maxHeight := -(minHeight - 1)

	for _, service := range c.services {
		if service.Y < minWidth {
			minWidth = service.Y
		}
		if service.X < minHeight {
			minHeight = service.X
		}
		if service.Y > maxWidth {
			maxWidth = service.Y
		}
		if service.X > maxHeight {
			maxHeight = service.X
		}
	}
	for _, service := range c.services {
		service.Y = service.Y - minWidth
		service.X = service.X - minHeight
	}
	return int(math.Abs(float64(maxWidth-minWidth))) + IconSize,
		int(math.Abs(float64(maxHeight-minHeight))) + IconSize
}

// Marshal renders the SVG to the given io.Writer
func (c *Canvas) Marshal(w io.Writer) {
	width, height := c.getRect()
	canvas := svg.New(w)
	canvas.Start(width, height)
	canvas.Def()
	for _, relation := range c.relations {
		relation.definition(canvas)
	}
	for _, service := range c.services {
		service.definition(canvas)
	}
	canvas.DefEnd()
	canvas.Gid("relations")
	for _, relation := range c.relations {
		relation.usage(canvas)
	}
	canvas.Gend()
	canvas.Gid("services")
	for _, service := range c.services {
		service.usage(canvas)
	}
	canvas.Gend()
	canvas.End()
}
