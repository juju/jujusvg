package jujusvg

import (
	"fmt"
	"io"

	"github.com/juju/xml"
)

const svgNamespace = "http://www.w3.org/2000/svg"

// Process an icon SVG file from a reader, removing anything surrounding
// the <svg></svg> tags, which would be invalid in this context (such as
// <?xml...?> decls, directives, etc), writing out to a writer.  In
// addition, loosely check that the icon is a valid SVG file.
func processIcon(r io.Reader, w io.Writer) error {
	dec := xml.NewDecoder(r)
	dec.DefaultSpace = svgNamespace

	enc := xml.NewEncoder(w)

	svgStartFound := false
	svgEndFound := false
	for {
		tok, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("cannot get token: %v", err)
		}
		if !svgStartFound {
			tag, ok := tok.(xml.StartElement)
			if ok && tag.Name.Space == svgNamespace && tag.Name.Local == "svg" {
				svgStartFound = true
			} else {
				continue
			}
		}
		if !svgEndFound {
			tag, ok := tok.(xml.EndElement)
			if ok && tag.Name.Space == svgNamespace && tag.Name.Local == "svg" {
				svgEndFound = true
			}
			if err := enc.EncodeToken(tok); err != nil {
				return fmt.Errorf("cannot encode token %#v: %v", tok, err)
			}
		} else {
			break
		}
	}

	if !svgStartFound || !svgEndFound {
		return fmt.Errorf("Icon does not appear to be a valid svg.")
	}

	if err := enc.Flush(); err != nil {
		return err
	}

	return nil
}
