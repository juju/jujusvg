package jujusvg

import (
	"bytes"
	"encoding/xml"
	"image"
	"io"
	"testing"

	svg "github.com/ajstarks/svgo"
	qt "github.com/frankban/quicktest"

	"github.com/juju/jujusvg/v5/assets"
)

func TestApplicationRender(t *testing.T) {
	c := qt.New(t)

	// Ensure that the Application's definition and usage methods output the
	// proper SVG elements.
	var tests = []struct {
		about       string
		application application
		expected    string
	}{
		{
			about: "Application without iconSrc, no def created",
			application: application{
				name: "foo",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconUrl: "foo",
			},
			expected: `<g transform="translate(0,0)" >
<title>foo</title>
<circle cx="90" cy="90" r="90" class="application-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<image x="42" y="42" width="96" height="96" xlink:href="foo" clip-path="url(#clip-mask)" />
<rect x="0" y="135" width="180" height="32" rx="2" ry="2" fill="rgba(220, 220, 220, 0.8)" />
<text x="90" y="157" text-anchor="middle" style="font-weight:200" >foo</text>
</g>
`,
		},
		{
			about: "Application with iconSrc",
			application: application{
				name:      "bar",
				charmPath: "bar",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconSrc: []byte("<svg>bar</svg>"),
			},
			expected: `<svg:svg xmlns:svg="http://www.w3.org/2000/svg" id="icon-1">bar</svg:svg><g transform="translate(0,0)" >
<title>bar</title>
<circle cx="90" cy="90" r="90" class="application-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<use x="0" y="0" xlink:href="#icon-1" transform="translate(42,42)" width="96" height="96" clip-path="url(#clip-mask)" />
<rect x="0" y="135" width="180" height="32" rx="2" ry="2" fill="rgba(220, 220, 220, 0.8)" />
<text x="90" y="157" text-anchor="middle" style="font-weight:200" >bar</text>
</g>
`,
		},
		{
			about: "Application with already def'd icon",
			application: application{
				name:      "baz",
				charmPath: "bar",
				point: image.Point{
					X: 0,
					Y: 0,
				},
				iconSrc: []byte("<svg>bar</svg>"),
			},
			expected: `<g transform="translate(0,0)" >
<title>baz</title>
<circle cx="90" cy="90" r="90" class="application-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<use x="0" y="0" xlink:href="#icon-1" transform="translate(42,42)" width="96" height="96" clip-path="url(#clip-mask)" />
<rect x="0" y="135" width="180" height="32" rx="2" ry="2" fill="rgba(220, 220, 220, 0.8)" />
<text x="90" y="157" text-anchor="middle" style="font-weight:200" >baz</text>
</g>
`,
		},
	}
	// Maintain our list of rendered icons outside the loop.
	iconsRendered := make(map[string]bool)
	iconIds := make(map[string]string)
	for i := range tests {
		test := tests[i]
		c.Run(test.about, func(c *qt.C) {
			var buf bytes.Buffer
			svg := svg.New(&buf)
			test.application.definition(svg, iconsRendered, iconIds)
			test.application.usage(svg, iconIds)
			c.Log(test.about)
			c.Log(buf.String())
			c.Assert(buf.String(), qt.Equals, test.expected)
		})
	}
}

func TestRelationRender(t *testing.T) {
	c := qt.New(t)

	// Ensure that the Relation's definition and usage methods output the
	// proper SVG elements.
	var buf bytes.Buffer
	svg := svg.New(&buf)
	relation := applicationRelation{
		name: "foo",
		applicationA: &application{
			point: image.Point{
				X: 0,
				Y: 0,
			},
		},
		applicationB: &application{
			point: image.Point{
				X: 100,
				Y: 100,
			},
		},
	}
	relation.definition(svg)
	relation.usage(svg)
	c.Assert(buf.String(), qt.Equals,
		`<g >
<title>foo</title>
<line x1="90" y1="90" x2="190" y2="190" stroke="#a7a7a7" stroke-width="1px" stroke-dasharray="62.71, 16" />
<use x="132" y="132" xlink:href="#healthCircle" />
<circle cx="153" cy="153" r="4" fill="#a7a7a7" />
<circle cx="126" cy="126" r="4" fill="#a7a7a7" />
</g>
`)
}

func TestIconClipPath(t *testing.T) {
	c := qt.New(t)

	// Ensure that the icon ClipPath returns the correctly sizes clipping Circle
	var buf bytes.Buffer
	svg := svg.New(&buf)
	canvas := Canvas{}
	canvas.iconClipPath(svg)
	c.Assert(buf.String(), qt.Equals,
		`<circle cx="47" cy="49" r="45" id="application-icon-mask" fill="none" />
<clipPath id="clip-mask" ><use x="0" y="0" xlink:href="#application-icon-mask" />
</clipPath>
`)
}

func TestLayout(t *testing.T) {
	c := qt.New(t)

	// Ensure that the SVG is sized exactly around the positioned applications.
	canvas := Canvas{}
	canvas.addApplication(&application{
		name: "application1",
		point: image.Point{
			X: 0,
			Y: 0,
		},
	})
	canvas.addApplication(&application{
		name: "application2",
		point: image.Point{
			X: 100,
			Y: 100,
		},
	})
	width, height := canvas.layout()
	c.Assert(width, qt.Equals, 281)
	c.Assert(height, qt.Equals, 281)
	canvas.addApplication(&application{
		name: "application3",
		point: image.Point{
			X: -100,
			Y: -100,
		},
	})
	canvas.addApplication(&application{
		name: "application4",
		point: image.Point{
			X: -100,
			Y: 100,
		},
	})
	canvas.addApplication(&application{
		name: "application5",
		point: image.Point{
			X: 200,
			Y: -100,
		},
	})
	width, height = canvas.layout()
	c.Assert(width, qt.Equals, 481)
	c.Assert(height, qt.Equals, 381)
}

func TestMarshal(t *testing.T) {
	c := qt.New(t)

	// Ensure that the internal representation of the canvas can be marshalled
	// to SVG.
	var buf bytes.Buffer
	canvas := Canvas{}
	applicationA := &application{
		name:      "application-a",
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
	applicationB := &application{
		name: "application-b",
		point: image.Point{
			X: 100,
			Y: 100,
		},
	}
	canvas.addApplication(applicationA)
	canvas.addApplication(applicationB)
	canvas.addRelation(&applicationRelation{
		name:         "relation",
		applicationA: applicationA,
		applicationB: applicationB,
	})
	canvas.Marshal(&buf)
	c.Logf("%s", buf.Bytes())
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
<circle cx="47" cy="49" r="45" id="application-icon-mask" fill="none" />
<clipPath id="clip-mask" ><use x="0" y="0" xlink:href="#application-icon-mask" />
</clipPath>
<g id="relations">
<g >
<title>relation</title>
<line x1="90" y1="90" x2="190" y2="190" stroke="#a7a7a7" stroke-width="1px" stroke-dasharray="62.71, 16" />
<use x="132" y="132" xlink:href="#healthCircle" />
<circle cx="153" cy="153" r="4" fill="#a7a7a7" />
<circle cx="126" cy="126" r="4" fill="#a7a7a7" />
</g>
</g>
<g id="applications">
<g transform="translate(0,0)" >
<title>application-a</title>
<circle cx="90" cy="90" r="90" class="application-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<use x="0" y="0" xlink:href="#icon-1" transform="translate(42,42)" width="96" height="96" clip-path="url(#clip-mask)" />
<rect x="0" y="135" width="180" height="32" rx="2" ry="2" fill="rgba(220, 220, 220, 0.8)" />
<text x="90" y="157" text-anchor="middle" style="font-weight:200" >application-a</text>
</g>
<g transform="translate(100,100)" >
<title>application-b</title>
<circle cx="90" cy="90" r="90" class="application-block" fill="#f5f5f5" stroke="#888" stroke-width="1" />
<image x="42" y="42" width="96" height="96" xlink:href="" clip-path="url(#clip-mask)" />
<rect x="0" y="135" width="180" height="32" rx="2" ry="2" fill="rgba(220, 220, 220, 0.8)" />
<text x="90" y="157" text-anchor="middle" style="font-weight:200" >application-b</text>
</g>
</g>
</svg>
`))
}

func assertXMLEqual(c *qt.C, obtained, expected []byte) {
	toksObtained := xmlTokens(c, obtained)
	toksExpected := xmlTokens(c, expected)
	c.Assert(toksObtained, qt.DeepEquals, toksExpected)
}

func xmlTokens(c *qt.C, data []byte) []xml.Token {
	dec := xml.NewDecoder(bytes.NewReader(data))
	var toks []xml.Token
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return toks
		}
		c.Assert(err, qt.IsNil)

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
