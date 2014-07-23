package jujusvg

import (
	"io"

	"github.com/ajstarks/svgo"
)

const (
	IconSize = 96
)

type Canvas struct {
	services []*Service
	relations []*Relation
}

type Service struct {
	Name string
	Url string
	IconUrl string
	X int
	Y int
}

type Relation struct {
	ServiceA *Service
	ServiceB *Service
}

func (r *Relation ) definition(canvas *svg.SVG) {
}

func (r *Relation) usage(canvas *svg.SVG) {
	canvas.Line(r.ServiceA.X, r.ServiceA.Y, r.ServiceB.X, r.ServiceB.Y)
}

func (s *Service) definition(canvas *svg.SVG) {
}

func (s *Service) usage(canvas *svg.SVG) {
	canvas.Image(s.X, s.Y, IconSize, IconSize, s.IconUrl)
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
			maxHeight = service.Y
		}
	}
	for _, service := range c.services {
		service.Y = service.Y - minWidth
		service.X = service.X - minHeight
	}
	return maxWidth - minWidth + IconSize, maxHeight - minHeight + IconSize
}

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
