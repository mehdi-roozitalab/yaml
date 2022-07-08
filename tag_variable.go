package yaml

var defineVarNames = []string{CreateTagName("define_var"), "define_var"}

type DefinetVariableTag struct{}

func (tag DefinetVariableTag) Names() []string { return defineVarNames }
func (tag DefinetVariableTag) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	if node.Kind != MappingNode {
		return nil, NewYamlErrorf(node, "%s may only applied to a mapping node", node.Tag)
	}

	for i := 0; i < len(node.Content); i += 2 {
		k := node.Content[i].Value
		var val interface{}
		if v, err := loader.ResolveTags(node.Content[i+1]); err != nil {
			return nil, err
		} else if err = v.Decode(&val); err != nil {
			return nil, NewYamlErrorf(v, "failed to parse node's value")
		} else {
			loader.Variables[k] = v
		}
	}
	return nil, nil
}
