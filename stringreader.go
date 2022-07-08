package yaml

import (
	"fmt"
	"strings"
)

type StringNode struct {
	Value string
	Node  *Node
}
type ListDefaultValue struct {
	Node  *Node
	Value string
}
type StringList struct {
	UsedSeparator string
	Values        []StringNode
	DefaultValue  *ListDefaultValue
	Data          map[string]interface{}
}

type StringListReader struct {
	DefaultSeparator byte
	ItemSeparators   []string
	ItemChild        string
	DefaultChild     string
	ExtraNodeParser  func(reader *StringListReader, list *StringList, nodeName string, node *Node) error
}

func (reader *StringListReader) AcceptObject() bool { return reader.ItemChild != "" }

func (reader *StringListReader) ReadStringList(loader *Loader, node *Node) (list *StringList, failedNode *Node, err error) {
	var context readStringListContext

	context.Init()

	switch node.Kind {
	case ScalarNode:
		context.ReadScalarNode(node)

	case SequenceNode:
		context.ReadSequenceNode(node)

	case MappingNode:
		context.ReadMappingNode(node)

	default:
		return nil, node, Err_BadNodeKind
	}
	return context.Result()
}

type readStringListContext struct {
	Reader     *StringListReader
	Loader     *Loader
	ItemsRead  bool
	List       *StringList
	FailedNode *Node
	Err        error
}

func (c *readStringListContext) Init() {
	c.List = &StringList{}
}
func (c *readStringListContext) IsFailed() bool { return c.Err != nil }
func (c *readStringListContext) Result() (*StringList, *Node, error) {
	if c.IsFailed() {
		return c.List, c.FailedNode, c.Err
	}
	return c.List, nil, nil
}

func (c *readStringListContext) AppendStringNode(value string, node *Node) {
	c.List.Values = append(c.List.Values, StringNode{Value: value, Node: node})
}
func (c *readStringListContext) ConvertNodeToStringAndAppendValue(node *Node) {
	var s string
	if s, c.Err = ToString(node); c.Err != nil {
		c.FailedNode = node
	} else {
		c.AppendStringNode(s, node)
	}
}
func (c *readStringListContext) ResolveNode(node *Node) *Node {
	if resolved, err := c.Loader.ResolveTags(node); err != nil {
		c.FailedNode = node
		c.Err = err
		return nil
	} else {
		return resolved
	}
}

func (c *readStringListContext) ReadDefaultValueFromString(s string) string {
	if n := strings.IndexByte(s, c.Reader.DefaultSeparator); n != -1 {
		c.List.DefaultValue = &ListDefaultValue{Value: s[n+1:]}
		return s[:n]
	}
	return s
}
func (c *readStringListContext) TryReadDefaultValueFromString(s string) string {
	if c.Reader.DefaultSeparator != 0 {
		return c.ReadDefaultValueFromString(s)
	}
	return s
}
func (c *readStringListContext) ReadPossiblyConcatenatedValuesFromString(s string, node *Node) {
	if len(c.Reader.ItemSeparators) != 0 {
		for _, sep := range c.Reader.ItemSeparators {
			if !strings.Contains(s, sep) {
				continue
			}

			c.List.UsedSeparator = sep
			for _, item := range strings.Split(s, sep) {
				c.AppendStringNode(item, node)
			}
		}
	} else {
		c.List.Values = []StringNode{{Value: s, Node: node}}
	}
}
func (c *readStringListContext) ReadScalarNode(node *Node) {
	s := c.TryReadDefaultValueFromString(node.Value)
	c.ReadPossiblyConcatenatedValuesFromString(s, node)
}

func (c *readStringListContext) ReadSequenceNodeValue(node *Node) {
	if node = c.ResolveNode(node); node == nil {
		return
	}

	c.ConvertNodeToStringAndAppendValue(node)
}
func (c *readStringListContext) ReadSequenceNode(node *Node) {
	for _, n := range node.Content {
		c.ReadSequenceNode(n)
		if c.Err != nil {
			return
		}
	}
}

func (c *readStringListContext) ReadMappingItemValue(node *Node) {
	switch node.Kind {
	case ScalarNode:
		c.ConvertNodeToStringAndAppendValue(node)

	case SequenceNode:
		for _, n := range node.Content {
			c.ConvertNodeToStringAndAppendValue(n)
		}

	default:
		c.FailedNode = node
		c.Err = fmt.Errorf("node must be a string or a list of strings: %w", Err_BadNodeKind)
	}
}
func (c *readStringListContext) ReadExtraNode(key string, node *Node) {
	if c.Reader.ExtraNodeParser != nil {
		if c.Err = c.Reader.ExtraNodeParser(c.Reader, c.List, key, node); c.Err != nil {
			c.FailedNode = node
			return
		}
	} else {
		c.FailedNode = node
		c.Err = fmt.Errorf("%s is an invalid key: %w", key, Err_InvalidChild)
	}
}
func (c *readStringListContext) ReadMappingNodeValue(key string, node *Node) {
	c.FailedNode = node
	if key == c.Reader.DefaultChild {
		c.List.DefaultValue = &ListDefaultValue{Node: node}
	} else if node = c.ResolveNode(node); node == nil {
		return
	} else if key == c.Reader.ItemChild {
		c.ReadMappingItemValue(node)
		c.ItemsRead = true
	} else {
		c.ReadExtraNode(key, node)
	}
}
func (c *readStringListContext) ReadMappingNode(node *Node) {
	if !c.Reader.AcceptObject() {
		c.FailedNode = node
		c.Err = Err_InvalidObject
		return
	}

	for i := 0; i < len(node.Content); i++ {
		key := node.Content[i].Value
		c.ReadMappingNodeValue(key, node.Content[i+1])
		if c.Err != nil {
			return
		}
	}

	if !c.ItemsRead {
		c.FailedNode = node
		c.Err = fmt.Errorf("missing %s: %w", c.Reader.ItemChild, Err_MissingItems)
	}
}
