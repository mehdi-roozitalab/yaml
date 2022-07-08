package yaml

var flatternNames = []string{CreateTagName("flattern"), "!flattern"}

// TagFlattern tag will be applied to a node of type sequence and will flattern sub lists to a flat list
type TagFlattern struct{}

func (tag TagFlattern) Names() []string { return flatternNames }
func (tag TagFlattern) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	if node.Kind != SequenceNode {
		return nil, NewYamlConstError(node, "flattern should only applied to a sequence")
	}

	content := make([]*Node, 0, len(node.Content))
	for _, c := range node.Content {
		if c.Kind == SequenceNode {
			content = append(content, c.Content...)
		} else {
			content = append(content, c)
		}
	}
	node.Content = content
	return node, nil
}
