package jujusvg

// Parser is an interface that various parsers implement
type Parser interface {
	Parse([]byte) (map[string]*Canvas, error)
}
