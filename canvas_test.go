package jujusvg

import (
	"bytes"
	"encoding/xml"
	"image"
	"io"

	"github.com/ajstarks/svgo"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"gopkg.in/juju/jujusvg.v1/assets"
)

type CanvasSuite struct{}

var _ = gc.Suite(&CanvasSuite{})

func (s *CanvasSuite) TestServiceRender(c *gc.C) {
	// Ensure that the Service's definition and usage methods output the
	// proper SVG elements.
	var tests = []struct {
		about    string
		service  service
		expected string
	}{
		{
			about: "Service without iconSrc, no def created",
			service: service{
				name: "foo",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconUrl: "foo",
			},
			expected: `<g transform="translate(0,0)" >
<circle cx="90" cy="90" r="90" class="service-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<circle cx="47" cy="49" r="45" id="service-icon-mask-foo" fill="none" />
<clipPath id="clip-foo" ><use x="0" y="0" xlink:href="#service-icon-mask-foo" />
</clipPath>
<image x="42" y="42" width="96" height="96" xlink:href="foo" clip-path="url(#clip-foo)" />
</g>
`,
		},
		{
			about: "Service with iconSrc",
			service: service{
				name:      "bar",
				charmPath: "bar",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconSrc: []byte("<svg>bar</svg>"),
			},
			expected: `<svg:svg xmlns:svg="http://www.w3.org/2000/svg" id="icon-1">bar</svg:svg><g transform="translate(0,0)" >
<circle cx="90" cy="90" r="90" class="service-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<circle cx="47" cy="49" r="45" id="service-icon-mask-bar" fill="none" />
<clipPath id="clip-bar" ><use x="0" y="0" xlink:href="#service-icon-mask-bar" />
</clipPath>
<use x="42" y="42" xlink:href="#icon-1" width="96" height="96" clip-path="url(#clip-bar)" />
</g>
`,
		},
		{
			about: "Service with already def'd icon",
			service: service{
				name:      "baz",
				charmPath: "bar",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconSrc: []byte("<svg>bar</svg>"),
			},
			expected: `<g transform="translate(0,0)" >
<circle cx="90" cy="90" r="90" class="service-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<circle cx="47" cy="49" r="45" id="service-icon-mask-baz" fill="none" />
<clipPath id="clip-baz" ><use x="0" y="0" xlink:href="#service-icon-mask-baz" />
</clipPath>
<use x="42" y="42" xlink:href="#icon-1" width="96" height="96" clip-path="url(#clip-baz)" />
</g>
`,
		},
	}
	// Maintain our list of rendered icons outside the loop.
	iconsRendered := make(map[string]bool)
	iconIds := make(map[string]string)
	for _, test := range tests {
		var buf bytes.Buffer
		svg := svg.New(&buf)
		test.service.definition(svg, iconsRendered, iconIds)
		test.service.usage(svg, iconIds)
		c.Log(test.about)
		c.Log(buf.String())
		c.Assert(buf.String(), gc.Equals, test.expected)
	}
}

func (s *CanvasSuite) TestRelationRender(c *gc.C) {
	// Ensure that the Relation's definition and usage methods output the
	// proper SVG elements.
	var buf bytes.Buffer
	svg := svg.New(&buf)
	relation := serviceRelation{
		serviceA: &service{
			point: image.Point{
				X: 0,
				Y: 0,
			},
		},
		serviceB: &service{
			point: image.Point{
				X: 100,
				Y: 100,
			},
		},
	}
	relation.definition(svg)
	relation.usage(svg)
	c.Assert(buf.String(), gc.Equals,
		`<line x1="90" y1="90" x2="190" y2="190" stroke="#a7a7a7" stroke-width="1px" stroke-dasharray="62.71, 16" />
<use x="132" y="132" xlink:href="#healthCircle" />
<circle cx="153" cy="153" r="4" fill="#a7a7a7" />
<circle cx="126" cy="126" r="4" fill="#a7a7a7" />
`)
}

func (s *CanvasSuite) TestLayout(c *gc.C) {
	// Ensure that the SVG is sized exactly around the positioned services.
	canvas := Canvas{}
	canvas.addService(&service{
		point: image.Point{
			X: 0,
			Y: 0,
		},
	})
	canvas.addService(&service{
		point: image.Point{
			X: 100,
			Y: 100,
		},
	})
	width, height := canvas.layout()
	c.Assert(width, gc.Equals, 281)
	c.Assert(height, gc.Equals, 281)
	canvas.addService(&service{
		point: image.Point{
			X: -100,
			Y: -100,
		},
	})
	canvas.addService(&service{
		point: image.Point{
			X: -100,
			Y: 100,
		},
	})
	canvas.addService(&service{
		point: image.Point{
			X: 200,
			Y: -100,
		},
	})
	width, height = canvas.layout()
	c.Assert(width, gc.Equals, 481)
	c.Assert(height, gc.Equals, 381)
}

func (s *CanvasSuite) TestMarshal(c *gc.C) {
	// Ensure that the internal representation of the canvas can be marshalled
	// to SVG.
	var buf bytes.Buffer
	canvas := Canvas{}
	serviceA := &service{
		name:      "service-a",
		charmPath: "trusty/svc-a",
		point: image.Point{
			X: 0,
			Y: 0,
		},
		iconSrc: []byte(`
			<svg xmlns="http://www.w3.org/2000/svg" class="blah">
				<circle cx="20" cy="20" r="20" style="fill:#000" />
			</svg>`),
	}
	serviceB := &service{
		name: "service-b",
		point: image.Point{
			X: 100,
			Y: 100,
		},
	}
	canvas.addService(serviceA)
	canvas.addService(serviceB)
	canvas.addRelation(&serviceRelation{
		serviceA: serviceA,
		serviceB: serviceB,
	})
	canvas.Marshal(&buf)
	c.Logf("%s", buf)
	assertXMLEqual(c, buf.Bytes(), []byte(`
<?xml version="1.0"?>
<!-- Generated by SVGo -->
<svg width="281" height="281"
     style="font-family:Ubuntu, sans-serif;" viewBox="0 0 281 281"
     xmlns="http://www.w3.org/2000/svg"
     xmlns:xlink="http://www.w3.org/1999/xlink">
<defs>
<g id="healthCircle" transform="scale(1.1)" >`+assets.RelationIconHealthy+`
</g>
<svg xmlns="http://www.w3.org/2000/svg" class="blah" id="icon-1">
&#x9;&#x9;&#x9;&#x9;<circle cx="20" cy="20" r="20" style="fill:#000"></circle>
&#x9;&#x9;&#x9;</svg></defs>
<g id="relations">
<line x1="90" y1="90" x2="190" y2="190" stroke="#a7a7a7" stroke-width="1px" stroke-dasharray="62.71, 16" />
<use x="132" y="132" xlink:href="#healthCircle" />
<circle cx="153" cy="153" r="4" fill="#a7a7a7" />
<circle cx="126" cy="126" r="4" fill="#a7a7a7" />
</g>
<g id="services">
<g transform="translate(0,0)" >
<circle cx="90" cy="90" r="90" class="service-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<circle cx="47" cy="49" r="45" id="service-icon-mask-service-a" fill="none" />
<clipPath id="clip-service-a" ><use x="0" y="0" xlink:href="#service-icon-mask-service-a" />
</clipPath>
<use x="42" y="42" xlink:href="#icon-1" width="96" height="96" clip-path="url(#clip-service-a)" />
</g>
<g transform="translate(100,100)" >
<circle cx="90" cy="90" r="90" class="service-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<circle cx="47" cy="49" r="45" id="service-icon-mask-service-b" fill="none" />
<clipPath id="clip-service-b" ><use x="0" y="0" xlink:href="#service-icon-mask-service-b" />
</clipPath>
<image x="42" y="42" width="96" height="96" xlink:href="" clip-path="url(#clip-service-b)" />
</g>
</g>
</svg>
`))
}

func assertXMLEqual(c *gc.C, obtained, expected []byte) {
	toksObtained := xmlTokens(c, obtained)
	toksExpected := xmlTokens(c, expected)
	c.Assert(toksObtained, jc.DeepEquals, toksExpected)
}

func xmlTokens(c *gc.C, data []byte) []xml.Token {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var toks []xml.Token
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return toks
		}
		c.Assert(err, gc.IsNil)

		if cdata, ok := tok.(xml.CharData); ok {
			// It's char data - trim all white space and ignore it
			// if it's all blank.
			cdata = bytes.TrimSpace(cdata)
			if len(cdata) == 0 {
				continue
			}
			tok = cdata
		}
		toks = append(toks, xml.CopyToken(tok))
	}
}
