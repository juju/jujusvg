package jujusvg

import (
	"io"
	"strings"

	"github.com/juju/xml"
	"gopkg.in/errgo.v1"
)

const svgNamespace = "http://www.w3.org/2000/svg"

// Process an icon SVG file from a reader, removing anything surrounding
// the <svg></svg> tags, which would be invalid in this context (such as
// <?xml...?> decls, directives, etc), writing out to a writer.  In
// addition, loosely check that the icon is a valid SVG file.
func processIcon(r io.Reader, w io.Writer, id string) error {
	dec := xml.NewDecoder(r)
	dec.DefaultSpace = svgNamespace

	enc := xml.NewEncoder(w)

	svgStartFound := false
	svgEndFound := false
	depth := 0
	for depth < 1 {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errgo.Notef(err, "cannot get token")
		}
		tag, ok := tok.(xml.StartElement)
		if ok && tag.Name.Space == svgNamespace && tag.Name.Local == "svg" {
			svgStartFound = true
			depth++
			st := tok.(xml.StartElement)
			for i := range st.Attr {
				if strings.ToLower(st.Attr[i].Name.Local) == "id" {
					st.Attr = append(st.Attr[:i], st.Attr[i+1:]...)
					break
				}
			}
			st.Attr = append(st.Attr, xml.Attr{
				Name: xml.Name{
					Local: "id",
				},
				Value: id,
			})
			if err := enc.EncodeToken(st); err != nil {
				return errgo.Notef(err, "cannot encode token %#v", st)
			}
		}
	}
	for depth > 0 {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errgo.Notef(err, "cannot get token")
		}
		switch tag := tok.(type) {
		case xml.StartElement:
			if tag.Name.Space == svgNamespace && tag.Name.Local == "svg" {
				depth++
			}
		case xml.EndElement:
			if tag.Name.Space == svgNamespace && tag.Name.Local == "svg" {
				depth--
				if depth == 0 {
					svgEndFound = true
				}
			}
		}
		if err := enc.EncodeToken(tok); err != nil {
			return errgo.Notef(err, "cannot encode token %#v", tok)
		}
	}

	if !svgStartFound || !svgEndFound {
		return errgo.Newf("icon does not appear to be a valid SVG")
	}

	if err := enc.Flush(); err != nil {
		return err
	}

	return nil
}
