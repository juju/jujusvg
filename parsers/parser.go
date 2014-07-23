// Parsers package contains parsers that take a given data format and output
// an internal representation of a Canvas with Services and Relations.
package parsers

import (
	"github.com/makyo/jujusvg"
)

// Parser is an interface that various parsers implement
type Parser interface {
	Parse([]byte) (map[string]jujusvg.Canvas, error)
}
