package yaml

import (
	"strconv"
	"strings"

	"github.com/mehdi-roozitalab/core_utils"
)

var (
	switchNames = []string{CreateTagName("switch"), "switch"}
	trueValues  = []string{"true", "yes", "ok", "1", "y"}
	falseValues = []string{"false", "no", "0", "n"}
)

type SwitchTag struct{}

func (tag SwitchTag) Names() []string { return switchNames }
func (tag SwitchTag) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	reader := switchReader{Loader: loader, SourceNode: node}
	return reader.Resolve()
}

type switchCase struct {
	Node *Node
	Case *Node
	Then *Node
}

func (sc *switchCase) Parse(node *Node) bool {
	if node.Kind != MappingNode {
		return false
	}

	sc.Node = node
	switch len(node.Content) {
	case 2:
		if node.Content[0].Value == "else" {
			sc.Then = node.Content[1]
			return true
		}
	case 4:
		if node.Content[0].Value == "case" && node.Content[2].Value == "then" {
			sc.Case = node.Content[1]
			sc.Then = node.Content[3]
			return true
		} else if node.Content[0].Value == "then" && node.Content[2].Value == "case" {
			sc.Case = node.Content[3]
			sc.Then = node.Content[1]
			return true
		}
	}

	return false
}
func (sc *switchCase) Match(loader *Loader) (bool, error) {
	if sc.Case == nil {
		return true, nil
	}

	if c, err := loader.ResolveTags(sc.Case); err != nil {
		return false, NewYamlError(sc.Case, err)
	} else {
		if c.Kind == ScalarNode {
			switch c.ShortTag() {
			case "!!bool":
				return c.Value == "true", nil
			case "!!str":
				v := strings.ToLower(c.Value)
				if core_utils.StringArrayContains(trueValues, v) {
					return true, nil
				} else if core_utils.StringArrayContains(falseValues, v) {
					return false, nil
				}
			case "!!int":
				if n, err := strconv.Atoi(c.Value); err == nil {
					return n != 0, nil
				}
			}
		}
		return false, NewYamlError(sc.Case, Err_InvalidCase)
	}
}

type switchReader struct {
	Loader      *Loader
	SourceNode  *Node
	SwitchCases []switchCase
}

func (r *switchReader) ValidateSourceNode() error {
	if r.SourceNode.Kind != SequenceNode {
		return NewYamlErrorf(r.SourceNode, "switch tag may only applied to a sequence of cases: %w", Err_BadNodeKind)
	}
	if len(r.SourceNode.Content) == 0 {
		return NewYamlConstError(r.SourceNode, "at least one case is required")
	}
	return nil
}
func (r *switchReader) ReadSwitchCases() error {
	if r.SourceNode.Kind != SequenceNode {
		return NewYamlErrorf(r.SourceNode, "switch tag may only applied to a sequence of cases: %w", Err_BadNodeKind)
	}

	r.SwitchCases = make([]switchCase, len(r.SourceNode.Content))
	for i := range r.SourceNode.Content {
		if !r.SwitchCases[i].Parse(r.SourceNode.Content[i]) {
			return NewYamlConstError(r.SourceNode.Content[i], "invalid case")
		}
	}
	return nil
}
func (r *switchReader) MoveElseCaseToEndOfCases() error {
	var elseCase *switchCase
	for i := range r.SwitchCases {
		if r.SwitchCases[i].Case == nil {
			if elseCase != nil {
				return NewYamlConstError(r.SwitchCases[i].Node, "multiple else in a single ")
			}
			elseCase = &switchCase{}
			*elseCase = r.SwitchCases[i]
			r.SwitchCases = append(r.SwitchCases[:i], r.SwitchCases[i+1:]...)
		}
	}
	if elseCase != nil {
		r.SwitchCases = append(r.SwitchCases, *elseCase)
	}
	return nil
}
func (r *switchReader) ResolveActiveNode() (*Node, error) {
	for _, sc := range r.SwitchCases {
		if match, err := sc.Match(r.Loader); err != nil {
			return nil, err
		} else if match {
			return r.Loader.ResolveTags(sc.Node)
		}
	}

	return nil, NewYamlConstError(r.SourceNode, "none of the cases match input value")
}
func (r *switchReader) Resolve() (*Node, error) {
	if err := r.ValidateSourceNode(); err != nil {
		return nil, err
	} else if err = r.ReadSwitchCases(); err != nil {
		return nil, err
	} else if err = r.MoveElseCaseToEndOfCases(); err != nil {
		return nil, err
	} else {
		return r.ResolveActiveNode()
	}
}
