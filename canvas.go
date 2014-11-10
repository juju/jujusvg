package jujusvg

import (
	"fmt"
	"image"
	"io"
	"math"

	svg "github.com/ajstarks/svgo"
)

const (
	// iconSize is always 96px per Juju charm conventions.
	iconSize           = 96
	healthCircleRadius = 10
	maxInt             = int(^uint(0) >> 1)
	minInt             = -(maxInt - 1)
)

// Canvas holds the parsed form of a bundle or environment.
type Canvas struct {
	services  []*service
	relations []*serviceRelation
}

// service represents a service deployed to an environment and contains the
// point of the top-left corner of the icon, icon URL, and additional metadata.
type service struct {
	name    string
	iconUrl string
	point   image.Point
}

// serviceRelation represents a relation created between two services.
type serviceRelation struct {
	serviceA *service
	serviceB *service
}

// definition creates any necessary defs that can be used later in the SVG.
func (s *service) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the service in the SVG.
func (s *service) usage(canvas *svg.SVG) {
	canvas.Image(s.point.X, s.point.Y, iconSize, iconSize, s.iconUrl)
}

// definition creates any necessary defs that can be used later in the SVG.
func (r *serviceRelation) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the relation in the SVG.
func (r *serviceRelation) usage(canvas *svg.SVG) {
	points := r.getShortestRelation()
	canvas.Line(
		points[0].X,
		points[0].Y,
		points[1].X,
		points[1].Y,
		`stroke="#38B44A"`,
		`stroke-width="2px"`,
		fmt.Sprintf(`stroke-dasharray="%s"`, strokeDashArray(points[0], points[1])))
	mid := points[0].Add(points[1]).Div(2).Add(point(-healthCircleRadius, -healthCircleRadius))
	canvas.Use(mid.X, mid.Y, "#healthCircle")
}

// getShortestRelation finds the shortest line between two services, assuming
// that each service can be connected on one of four cardinal points only.
func (r *serviceRelation) getShortestRelation() []image.Point {
	aConnectors, bConnectors := r.serviceA.generateConnectors(), r.serviceB.generateConnectors()
	shortestDistance := float64(maxInt)
	shortestPair := []image.Point{
		r.serviceA.point,
		r.serviceB.point,
	}
	for _, pointA := range aConnectors {
		for _, pointB := range bConnectors {
			distance := pointDistance(pointA, pointB)
			if distance < shortestDistance {
				shortestDistance = distance
				shortestPair = []image.Point{
					pointA,
					pointB,
				}
			}
		}
	}
	return shortestPair
}

// generateConnectors generates the points for each of the four cardinal points
// of each service.
func (s *service) generateConnectors() []image.Point {
	return []image.Point{
		point(s.point.X+iconSize/2, s.point.Y),
		point(s.point.X, s.point.Y+iconSize/2),
		point(s.point.X+iconSize/2, s.point.Y+iconSize),
		point(s.point.X+iconSize, s.point.Y+iconSize/2),
	}
}

// strokeDashArray generates the stroke-dasharray attribute content so that
// the relation health indicator is placed in an empty space.
func strokeDashArray(pointA, pointB image.Point) string {
	return fmt.Sprintf("%.2f, 20",
		pointDistance(pointA, pointB)/2-healthCircleRadius)
}

// addService adds a new service to the canvas.
func (c *Canvas) addService(s *service) {
	c.services = append(c.services, s)
}

// addRelation adds a new relation to the canvas.
func (c *Canvas) addRelation(r *serviceRelation) {
	c.relations = append(c.relations, r)
}

// layout adjusts all items so that they are positioned appropriately,
// and returns the overall size of the canvas.
func (c *Canvas) layout() (int, int) {
	minWidth := maxInt
	minHeight := maxInt
	maxWidth := minInt
	maxHeight := minInt

	for _, service := range c.services {
		if service.point.X < minWidth {
			minWidth = service.point.X
		}
		if service.point.Y < minHeight {
			minHeight = service.point.Y
		}
		if service.point.X > maxWidth {
			maxWidth = service.point.X
		}
		if service.point.Y > maxHeight {
			maxHeight = service.point.Y
		}
	}
	for _, service := range c.services {
		service.point = service.point.Sub(point(minWidth, minHeight))
	}
	return abs(maxWidth-minWidth) + iconSize,
		abs(maxHeight-minHeight) + iconSize
}

func (c *Canvas) definition(canvas *svg.SVG) {
	canvas.Def()
	defer canvas.DefEnd()

	// Relation health circle.
	canvas.Gid("healthCircle")
	canvas.Circle(
		healthCircleRadius,
		healthCircleRadius,
		healthCircleRadius,
		"stroke:#38B44A;fill:none;stroke-width:2px")
	canvas.Circle(
		healthCircleRadius,
		healthCircleRadius,
		healthCircleRadius/2,
		"fill:#38B44A")
	canvas.Gend()

	// Service and relation specific defs.
	for _, relation := range c.relations {
		relation.definition(canvas)
	}
	for _, service := range c.services {
		service.definition(canvas)
	}
}

func (c *Canvas) relationsGroup(canvas *svg.SVG) {
	canvas.Gid("relations")
	defer canvas.Gend()
	for _, relation := range c.relations {
		relation.usage(canvas)
	}
}

func (c *Canvas) servicesGroup(canvas *svg.SVG) {
	canvas.Gid("services")
	defer canvas.Gend()
	for _, service := range c.services {
		service.usage(canvas)
	}
}

// Marshal renders the SVG to the given io.Writer
func (c *Canvas) Marshal(w io.Writer) {
	width, height := c.layout()

	// TODO check write errors and return an error from
	// Marshal if the write fails. The svg package does not
	// itself check or return write errors; a possible workaround
	// is to wrap the writer in a custom writer that panics
	// on error, and catch the panic here.

	canvas := svg.New(w)
	canvas.Start(width, height)
	defer canvas.End()
	c.definition(canvas)
	c.relationsGroup(canvas)
	c.servicesGroup(canvas)
}

// abs returns the absolute value of a number.
// Utility function for readability.
func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

// pointDistance calculates the distance between two points.
func pointDistance(p1, p2 image.Point) float64 {
	dp := p1.Sub(p2)
	return math.Sqrt(square(float64(dp.X)) + square(float64(dp.Y)))
}

// square multiplies a number by itself.
// Utility function for readability
func square(x float64) float64 {
	return x * x
}

// point generates an image.Point given its coordinates.
// Utility function for readability.
func point(x, y int) image.Point {
	return image.Point{x, y}
}
