package jujusvg

import (
	"image"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestGetPointOutside(t *testing.T) {
	c := qt.New(t)

	var tests = []struct {
		about    string
		vertices []image.Point
		expected image.Point
	}{
		{
			about:    "zero vertices",
			vertices: []image.Point{},
			expected: image.Point{0, 0},
		},
		{
			about:    "one vertex",
			vertices: []image.Point{{0, 0}},
			expected: image.Point{10, 10},
		},
		{
			about:    "two vertices",
			vertices: []image.Point{{0, 0}, {10, 10}},
			expected: image.Point{20, 20},
		},
		{
			about:    "three vertices (convexHull fall through)",
			vertices: []image.Point{{0, 0}, {0, 10}, {10, 0}},
			expected: image.Point{10, 20},
		},
		{
			about:    "four vertices",
			vertices: []image.Point{{0, 0}, {0, 10}, {10, 0}, {10, 10}},
			expected: image.Point{20, 20},
		},
	}
	for i := range tests {
		test := tests[i]
		c.Run(test.about, func(c *qt.C) {
			c.Assert(getPointOutside(test.vertices, image.Point{10, 10}), qt.Equals, test.expected)
		})
	}
}

func TestConvexHull(t *testing.T) {
	c := qt.New(t)

	// Zero vertices
	vertices := []image.Point{}
	c.Assert(convexHull(vertices), qt.DeepEquals, []image.Point{{0, 0}})

	// Identities
	vertices = []image.Point{{1, 1}}
	c.Assert(convexHull(vertices), qt.DeepEquals, vertices)

	vertices = []image.Point{{1, 1}, {2, 2}}
	c.Assert(convexHull(vertices), qt.DeepEquals, vertices)

	vertices = []image.Point{{1, 1}, {2, 2}, {1, 2}}
	c.Assert(convexHull(vertices), qt.DeepEquals, vertices)

	// > 3 vertices
	vertices = []image.Point{}
	for i := 0; i < 100; i++ {
		vertices = append(vertices, image.Point{i / 10, i % 10})
	}
	c.Assert(convexHull(vertices), qt.DeepEquals, []image.Point{
		{0, 0},
		{9, 0},
		{9, 9},
		{0, 9},
	})
}
