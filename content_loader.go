package yaml

import "fmt"

type contentLoader struct {
	loader   *Loader
	filename string
	target   interface{}
}

func (c *contentLoader) fixNodeLocation(node, parent *Node) {
	if !isCommentsFixed(node) {
		fixNodeComment(node, c.filename)
	}

	if node.Kind == SequenceNode {
		for i, ch := range node.Content {
			fixNodeComment(ch, c.filename)
			ch.LineComment = fmt.Sprintf("%s[%d]", node.LineComment, i)
			c.fixNodeLocation(ch, node)
		}
	} else if node.Kind == MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			fixNodeComment(node.Content[i], c.filename)
			node.Content[i].LineComment = ""

			fixNodeComment(node.Content[i+1], c.filename)
			node.Content[i+1].LineComment = fmt.Sprintf("%s.%s", node.LineComment, node.Content[i].Value)

			c.fixNodeLocation(node.Content[i+1], node)
		}
	}
}

func (c *contentLoader) UnmarshalYAML(node *Node) error {
	// first of all fix location of the node
	c.fixNodeLocation(node, nil)

	var err error
	node, err = c.loader.ResolveTags(node)
	if err != nil {
		return err
	}
	// possibly fix location of any added node
	c.fixNodeLocation(node, nil)
	return node.Decode(c.target)
}
