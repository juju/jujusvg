package readers

import (
	"github.com/makyo/jujusvg"
)

type Reader interface {
	Read([]byte) (map[string]jujusvg.Canvas, error)
}
