package parsers

import (
	"github.com/makyo/jujusvg"
)

type Parser interface {
	Parse([]byte) (map[string]jujusvg.Canvas, error)
}
