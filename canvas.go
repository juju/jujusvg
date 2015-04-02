package jujusvg

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"math"
	"regexp"

	svg "github.com/ajstarks/svgo"

	"github.com/juju/jujusvg/assets"
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

	fontColor     = "#505050"
	relationColor = "#38B44A"
)

// Canvas holds the parsed form of a bundle or environment.
type Canvas struct {
	services      []*service
	relations     []*serviceRelation
	iconsRendered map[string]bool
}

// service represents a service deployed to an environment and contains the
// point of the top-left corner of the icon, icon URL, and additional metadata.
type service struct {
	name      string
	charmPath string
	iconUrl   string
	iconSrc   string
	point     image.Point
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

var (
	pathSanitizer = regexp.MustCompile(`\W+`)
	piSanitizer   = regexp.MustCompile(`<\?xml[^>]*>`)
)

// sanitizeCharmPath ensures that a given string will be safe to use as a
// CSS selector attribute (id or class).
func sanitizeSelector(selector string) string {
	return pathSanitizer.ReplaceAllString(selector, "-")
}

// dePI removes any processing instructions to ensure a valid SVG.
func dePI(svg string) string {
	return piSanitizer.ReplaceAllString(svg, "")
}

// definition creates any necessary defs that can be used later in the SVG.
func (s *service) definition(canvas *svg.SVG, iconsRendered map[string]bool) error {
	if _, ok := iconsRendered[s.charmPath]; s.iconSrc != "" && ok == false {
		iconsRendered[s.charmPath] = true

		canvas.Group(fmt.Sprintf(`id="icon-%s"`, sanitizeSelector(s.charmPath)))
		defer canvas.Gend()

		// Temporary solution:
		iconBuf := bytes.NewBufferString(s.iconSrc)
		return processIcon(iconBuf, canvas.Writer)
	}
	return nil
}

// usage creates any necessary tags for actually using the service in the SVG.
func (s *service) usage(canvas *svg.SVG) {
	canvas.Use(
		s.point.X,
		s.point.Y,
		"#serviceBlock",
		fmt.Sprintf(`id="%s"`, s.name))
	if s.iconSrc != "" {
		canvas.Use(
			s.point.X+serviceBlockSize/2-iconSize/2,
			s.point.Y+serviceBlockSize/2-iconSize/2,
			fmt.Sprintf("#icon-%s", sanitizeSelector(s.charmPath)),
			fmt.Sprintf(`width="%d" height="%d"`, iconSize, iconSize))
	} else {
		canvas.Image(
			s.point.X+serviceBlockSize/2-iconSize/2,
			s.point.Y+serviceBlockSize/2-iconSize/2,
			iconSize,
			iconSize,
			s.iconUrl)
	}
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
	canvas.Group(`id="serviceBlock"`,
		`transform="scale(0.8)"`)
	io.WriteString(canvas.Writer, assets.ServiceModule)
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
	c.iconsRendered = make(map[string]bool)
	for _, service := range c.services {
		service.definition(canvas, c.iconsRendered)
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

// Marshal renders the SVG to the given io.Writer.
func (c *Canvas) Marshal(w io.Writer) {

	// TODO check write errors and return an error from
	// Marshal if the write fails. The svg package does not
	// itself check or return write errors; a possible work-around
	// is to wrap the writer in a custom writer that panics
	// on error, and catch the panic here.
	width, height := c.layout()

	canvas := svg.New(w)
	canvas.Start(
		width,
		height,
		fmt.Sprintf(`style="font-family:Ubuntu, sans-serif;" viewBox="0 0 %d %d"`,
			width, height))
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
