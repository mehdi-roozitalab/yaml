package yaml

import (
	"fmt"

	"github.com/mehdi-roozitalab/core_utils"
)

const (
	Err_InvalidObject       = core_utils.ConstError("value must be a string or an array of strings, object value is not value")
	Err_InvalidChild        = core_utils.ConstError("unknown child")
	Err_MissingItems        = core_utils.ConstError("missing any items")
	Err_BadNodeKind         = core_utils.ConstError("bad kind of node")
	Err_MissingRequiredNode = core_utils.ConstError("missing required node")
	Err_InvalidCase         = core_utils.ConstError("invalid case, case value must be a boolean")
)

type YamlError struct {
	Location
	Err error
}

func (e *YamlError) Error() string {
	return fmt.Sprintf("%s: %v", e.Location, e.Err)
}

func NewYamlError(node *Node, err error) error {
	if _, ok := err.(*YamlError); ok {
		return err
	}

	return &YamlError{
		Location: NodeLocation(node),
		Err:      err,
	}
}
func NewYamlConstError(node *Node, err string) error {
	return NewYamlError(node, core_utils.ConstError(err))
}
func NewYamlErrorf(node *Node, format string, a ...interface{}) error {
	return NewYamlError(node, fmt.Errorf(format, a...))
}
