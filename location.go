package yaml

import "fmt"

type Location struct {
	Filename string
	Line     int
	Column   int
	Path     string
}

func (loc Location) String() string {
	if loc.Path != "" {
		return fmt.Sprintf("%s@%s(%d:%d)", loc.Path, loc.Filename, loc.Line, loc.Column)
	}
	return fmt.Sprintf("(%d:%d)", loc.Line, loc.Column)
}
