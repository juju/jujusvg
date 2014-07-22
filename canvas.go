package jujusvg

import (
	"github.com/ajstarks/svgo"

	"io"
)

type Canvas struct {
	out io.Writer
	services []*Service
	relations []*Relation
}

type Service struct {
	name string
	url string
	iconUrl string
	x int
	y int
}

type Relation struct {
	serviceA *Service
	serviceB *Service
}

func (r *Relation ) definition(canvas *svg.SVG) {
}

func (r *Relation) usage(canvas *svg.SVG) {
	canvas.Line(r.serviceA.x, r.serviceA.y, r.serviceB.x, r.serviceB.y)
}

func (s *Service) definition(canvas *svg.SVG) {
}

func (s *Service) usage(canvas *svg.SVG) {
	canvas.Image(s.x, s.y, 96, 96, s.iconUrl)
}

func (c *Canvas) AddService(s *Service) {
	c.services = append(c.services, s)
}

func (c *Canvas) AddRelation(r *Relation) {
	c.relations = append(c.relations, r)
}

func (c *Canvas) getRect() (int, int) {
	minWidth := int(^uint(0) >> 1)
	minHeight := int(^uint(0) >> 1)
	maxWidth := -(minWidth - 1)
	maxHeight := -(minHeight - 1)

	for _, service := range c.services {
		if service.y < minWidth {
			minWidth = service.y
		}
		if service.x < minHeight {
			minHeight = service.x
		}
		if service.y > maxWidth {
			maxWidth = service.y
		}
		if service.x > maxHeight {
			maxHeight = service.y
		}
	}
	for _, service := range c.services {
		service.y = service.y - minWidth
		service.x = service.x - minHeight
	}
	return maxWidth - minWidth, maxHeight - minHeight
}

func (c *Canvas) Marshal() (string) {
	width, height := c.getRect()
	canvas := svg.New(c.out)
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
	return "done"
}
