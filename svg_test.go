package jujusvg

import (
	"bytes"

	gc "gopkg.in/check.v1"
)

type SVGSuite struct{}

var _ = gc.Suite(&SVGSuite{})

func (s *SVGSuite) TestProcessIcon(c *gc.C) {
	tests := []struct {
		about    string
		icon     []byte
		expected []byte
		err      string
	}{
		{
			about: "Nothing stripped",
			icon: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "ProcInst at start stripped",
			icon: []byte(`
				<?xml version="1.0"?>
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "Directive at start stripped",
			icon: []byte(`
				<!DOCTYPE svg>
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "ProcInst at end stripped",
			icon: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				<?xml foo="bar"?>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "Directive at end stripped",
			icon: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>
				<!DOCTYPE svg>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "ProcInsts/Directives inside svg left in place",
			icon: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<!DOCTYPE svg>
					<?proc foo="bar"?>
					<g id="foo"></g>
				</svg>
				`),
			expected: []byte(`
				<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100">
					<!DOCTYPE svg>
					<?proc foo="bar"?>
					<g id="foo"></g>
				</svg>`),
		},
		{
			about: "Not an SVG",
			icon: []byte(`
				<html xmlns="foo">
					<body>bad-wolf</body>
				</html>
				`),
			err: "Icon does not appear to be a valid svg.",
		},
	}
	for _, test := range tests {
		in := bytes.NewBuffer(test.icon)
		out := bytes.Buffer{}
		err := processIcon(in, &out)
		if test.err != "" {
			c.Assert(err, gc.ErrorMatches, test.err)
		} else {
			c.Assert(err, gc.IsNil)
			assertXMLEqual(c, out.Bytes(), test.expected)
		}
	}
}
