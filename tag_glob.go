package yaml

import "path/filepath"

var globNames = []string{CreateTagName("glob"), "!glob"}

// TagGlob tag that will be applied to a string and return list of all files that match specified glob pattern
type TagGlob struct{}

func (tag TagGlob) Names() []string { return globNames }
func (tag TagGlob) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	if node.Kind != ScalarNode {
		return nil, NewYamlErrorf(node, "%s must applied to a string value", node.Tag)
	}

	if matches, err := filepath.Glob(node.Value); err != nil {
		return nil, NewYamlError(node, err)
	} else {
		return StringListToSequenceNode(node, matches), nil
	}
}
