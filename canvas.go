package jujusvg

import (
	"fmt"
	"image"
	"io"
	"math"

	svg "github.com/ajstarks/svgo"
)

const (
	iconSize           = 96
	serviceBlockSize   = 189
	healthCircleRadius = 10
	relationLineWidth  = 2
	maxInt             = int(^uint(0) >> 1)
	minInt             = -(maxInt - 1)
	maxHeight          = 450
	maxWidth           = 1000
	viewBoxHeight      = 600

	fontColor     = "#505050"
	relationColor = "#38B44A"
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

// line represents a line segment with two endpoints.
type line struct {
	p0, p1 image.Point
}

// definition creates any necessary defs that can be used later in the SVG.
func (s *service) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the service in the SVG.
func (s *service) usage(canvas *svg.SVG) {
	canvas.Use(
		s.point.X,
		s.point.Y,
		"#serviceBlock")
	canvas.Image(
		s.point.X+serviceBlockSize/2-iconSize/2,
		s.point.Y+serviceBlockSize/2-iconSize/2,
		iconSize,
		iconSize,
		s.iconUrl)
	canvas.Textlines(
		s.point.X+serviceBlockSize/2,
		s.point.Y+serviceBlockSize/6,
		[]string{s.name},
		serviceBlockSize/10,
		0,
		"#505050",
		"middle")
}

// definition creates any necessary defs that can be used later in the SVG.
func (r *serviceRelation) definition(canvas *svg.SVG) {
}

// usage creates any necessary tags for actually using the relation in the SVG.
func (r *serviceRelation) usage(canvas *svg.SVG) {
	l := r.shortestRelation()
	canvas.Line(
		l.p0.X,
		l.p0.Y,
		l.p1.X,
		l.p1.Y,
		fmt.Sprintf(`stroke="%s"`, relationColor),
		fmt.Sprintf(`stroke-width="%dpx"`, relationLineWidth),
		fmt.Sprintf(`stroke-dasharray="%s"`, strokeDashArray(l)))
	mid := l.p0.Add(l.p1).Div(2).Sub(point(healthCircleRadius, healthCircleRadius))
	canvas.Use(mid.X, mid.Y, "#healthCircle")
}

// shortestRelation finds the shortest line between two services, assuming
// that each service can be connected on one of four cardinal points only.
func (r *serviceRelation) shortestRelation() line {
	aConnectors, bConnectors := r.serviceA.cardinalPoints(), r.serviceB.cardinalPoints()
	shortestDistance := float64(maxInt)
	shortestPair := line{
		p0: r.serviceA.point,
		p1: r.serviceB.point,
	}
	for _, pointA := range aConnectors {
		for _, pointB := range bConnectors {
			ab := line{p0: pointA, p1: pointB}
			distance := ab.length()
			if distance < shortestDistance {
				shortestDistance = distance
				shortestPair = ab
			}
		}
	}
	return shortestPair
}

// cardinalPoints generates the points for each of the four cardinal points
// of each service.
func (s *service) cardinalPoints() []image.Point {
	return []image.Point{
		point(s.point.X+serviceBlockSize/2, s.point.Y),
		point(s.point.X, s.point.Y+serviceBlockSize/2),
		point(s.point.X+serviceBlockSize/2, s.point.Y+serviceBlockSize),
		point(s.point.X+serviceBlockSize, s.point.Y+serviceBlockSize/2),
	}
}

// strokeDashArray generates the stroke-dasharray attribute content so that
// the relation health indicator is placed in an empty space.
func strokeDashArray(l line) string {
	return fmt.Sprintf("%.2f, %d", l.length()/2-healthCircleRadius, healthCircleRadius*2)
}

// length calculates the length of a line.
func (l *line) length() float64 {
	dp := l.p0.Sub(l.p1)
	return math.Sqrt(square(float64(dp.X)) + square(float64(dp.Y)))
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
	return abs(maxWidth-minWidth) + serviceBlockSize,
		abs(maxHeight-minHeight) + serviceBlockSize
}

func (c *Canvas) definition(canvas *svg.SVG) {
	canvas.Def()
	defer canvas.DefEnd()

	// Service block.
	// Note: this is implemented based off the service block SVG provided by
	// design; any changes to that will likely incur an entire rewrite of this
	// bit of SVGo.  See the README for more information.
	canvas.Group(`id="serviceBlock"`,
		`transform="translate(115.183,4.8),scale(0.8)"`)
	canvas.Gtransform("translate(-399.571,-251.207)")
	canvas.Path(`M410.565,479.165h-73.988c-38.324,0-57.56,0-68.272-10.713c-10.712-10.713-10.712-29.949-10.712-68.273
v-73.986c-0.001-38.324-0.001-57.561,10.711-68.273c10.713-10.713,29.949-10.713,68.274-10.713h73.988
c38.324,0,57.561,0,68.272,10.713c10.713,10.712,10.713,29.949,10.713,68.273v73.986c0,38.324,0,57.561-10.713,68.273
C468.126,479.165,448.889,479.165,410.565,479.165z M336.577,257.207c-34.445,0-53.419,0-61.203,7.784
s-7.783,26.757-7.782,61.202v73.986c0,34.444,0,53.419,7.784,61.202c7.784,7.784,26.757,7.784,61.201,7.784h73.988
c34.444,0,53.418,0,61.202-7.784c7.783-7.783,7.783-26.758,7.783-61.202v-73.986c0-34.444,0-53.418-7.783-61.202
c-7.784-7.784-26.758-7.784-61.202-7.784H336.577z`,
		`fill="#BBBBBB"`)
	canvas.Path(`M410.565,479.165h-73.988c-38.324,0-57.56,0-68.272-10.713c-10.712-10.713-10.712-29.949-10.712-68.273
v-73.986c0-38.324,0-57.561,10.712-68.273c10.713-10.713,29.949-10.713,68.272-10.713h73.988c38.324,0,57.561,0,68.272,10.713
c10.713,10.712,10.713,29.949,10.713,68.273v73.986c0,38.324,0,57.561-10.713,68.273
C468.126,479.165,448.889,479.165,410.565,479.165z M336.577,257.207c-34.444,0-53.417,0-61.201,7.784
s-7.784,26.758-7.784,61.202v73.986c0,34.444,0,53.419,7.784,61.202c7.784,7.784,26.757,7.784,61.201,7.784h73.988
c34.444,0,53.418,0,61.201-7.784c7.784-7.783,7.784-26.758,7.784-61.202v-73.986c0-34.444,0-53.418-7.784-61.202
c-7.783-7.784-26.757-7.784-61.201-7.784H336.577z`,
		`fill="#BBBBBB"`)
	canvas.Gend() // Gtransform
	canvas.Path(`M-42,219.958h32c2.209,0,4,1.791,4,4v2c0,2.209-1.791,4-4,4h-32
c-2.209,0-4-1.791-4-4v-2C-46,221.749-44.209,219.958-42,219.958z`,
		`fill-rule="evenodd"`,
		`clip-rule="evenodd"`,
		`fill="#BBBBBB"`)
	canvas.Path(`M-42-6h32c2.209,0,4,1.791,4,4v2c0,2.209-1.791,4-4,4h-32
c-2.209,0-4-1.791-4-4v-2C-46-4.209-44.209-6-42-6z`,
		`fill-rule="evenodd"`,
		`clip-rule="evenodd"`,
		`fill="#BBBBBB"`)
	canvas.Path(`M81.979,127.979v-32c0-2.209,1.791-4,4-4h2c2.209,0,4,1.791,4,4
v32c0,2.209-1.791,4-4,4h-2C83.771,131.979,81.979,130.188,81.979,127.979z`,
		`fill-rule="evenodd"`,
		`clip-rule="evenodd"`,
		`fill="#BBBBBB"`)
	canvas.Path(`M-143.979,127.979v-32c0-2.209,1.791-4,4-4h2c2.209,0,4,1.791,4,4
v32c0,2.209-1.791,4-4,4h-2C-142.188,131.979-143.979,130.188-143.979,127.979z`,
		`fill-rule="evenodd"`,
		`clip-rule="evenodd"`,
		`fill="#BBBBBB"`)
	canvas.Path(`M10.994-1h-73.988c-73.987,0-73.987,0-73.985,73.986v73.986c0,73.986,0,73.986,73.985,73.986h73.988
c73.985,0,73.985,0,73.985-73.986V72.986C84.979-1,84.979-1,10.994-1z`,
		`fill="#FFFFFF"`)
	canvas.Gend() // Gid

	// Relation health circle.
	canvas.Gid("healthCircle")
	canvas.Circle(
		healthCircleRadius,
		healthCircleRadius,
		healthCircleRadius,
		fmt.Sprintf("stroke:%s;fill:none;stroke-width:%dpx", relationColor, relationLineWidth))
	canvas.Circle(
		healthCircleRadius,
		healthCircleRadius,
		healthCircleRadius/2,
		fmt.Sprintf("fill:%s", relationColor))
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

// Compute the scale.
func (c *Canvas) computeScale(width, height int) float32 {
	scale := float32(1)
	if height > maxHeight {
		scale = maxHeight / float32(height)
	}
	if float32(width)*scale > maxWidth {
		scale = maxWidth / float32(width)
	}

	return scale
}

// Marshal renders the SVG to the given io.Writer.
func (c *Canvas) Marshal(w io.Writer) {

	// TODO check write errors and return an error from
	// Marshal if the write fails. The svg package does not
	// itself check or return write errors; a possible work-around
	// is to wrap the writer in a custom writer that panics
	// on error, and catch the panic here.
	width, height := c.layout()
	scale := c.computeScale(width, height)

	canvas := svg.New(w)
	newWidth := int((float32(viewBoxHeight) / float32(height)) * float32(width))
	canvas.Start(
		width,
		height,
		fmt.Sprintf(`style="font-family:Ubuntu, sans-serif;" viewBox="0 0 %d %d"`,
			newWidth, viewBoxHeight),
		fmt.Sprintf(`transform="scale(%f)"`, scale))
	defer canvas.End()
	c.definition(canvas)
	c.relationsGroup(canvas)
	c.servicesGroup(canvas)
}

// abs returns the absolute value of a number.
func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}

// square multiplies a number by itself.
func square(x float64) float64 {
	return x * x
}

// point generates an image.Point given its coordinates.
func point(x, y int) image.Point {
	return image.Point{x, y}
}
