package yaml

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mehdi-roozitalab/core_utils"
)

type NodeComments struct {
	Head string `json:"head"`
	Line string `json:"line"`
	Foot string `json:"foot"`
}

var loaderId = uuid.NewString()

func isCommentsFixed(node *Node) bool { return strings.HasPrefix(node.FootComment, loaderId) }
func fixNodeComment(node *Node, filename string) {
	if isCommentsFixed(node) {
		return
	}

	comments := map[string]string{}
	if node.LineComment != "" {
		comments["line"] = node.LineComment
	}
	if node.HeadComment != "" {
		comments["head"] = node.HeadComment
	}
	if node.FootComment != "" {
		comments["foot"] = node.FootComment
	}

	if len(comments) != 0 {
		c, _ := json.Marshal(comments)
		node.HeadComment = string(c)
	}
	node.FootComment = loaderId + "-" + filename
	node.LineComment = ""
}

func nodeFilename(node *Node) string { return node.FootComment[len(loaderId)+1:] }
func nodeName(node *Node) string     { return node.LineComment }
func nodeFullName(node *Node) string {
	return fmt.Sprintf("%s(%d:%d):%s", nodeFilename(node), node.Line, node.Column, nodeName(node))
}
func nodeLocation(node *Node) Location {
	return Location{nodeFilename(node), node.Line, node.Column, nodeName(node)}
}

func NodeFilename(node *Node) string {
	if isCommentsFixed(node) {
		return nodeFilename(node)
	}
	return ""
}
func NodeName(node *Node) string {
	if isCommentsFixed(node) {
		return nodeName(node)
	}
	return ""
}
func NodeFullName(node *Node) string {
	if isCommentsFixed(node) {
		return nodeFullName(node)
	}
	return fmt.Sprintf("(%d:%d)", node.Line, node.Column)
}
func NodeLocation(node *Node) Location {
	if isCommentsFixed(node) {
		return nodeLocation(node)
	}
	return Location{Line: node.Line, Column: node.Column}
}

func GetNodeComments(node *Node) NodeComments {
	if isCommentsFixed(node) {
		var nodeComments NodeComments
		if node.HeadComment == "" {
			return NodeComments{}
		}
		_ = json.Unmarshal([]byte(node.HeadComment), &nodeComments)
		return nodeComments
	} else {
		return NodeComments{
			Head: node.HeadComment,
			Line: node.LineComment,
			Foot: node.FootComment,
		}
	}
}

// ProcessMappingNode2 execute a function(``processor``) for each mapping node in a mapping node
// if ``node`` is not a ``MappingNode`` return ``Err_BadNodeKind``.
// if ``processor`` return false or return an error, this function immediately return and ignore
// processing rest of the nodes.
func ProcessMappingNode2(node *Node, processor func(index int, name string, node *Node) (bool, error)) error {
	if node.Kind != MappingNode {
		return NewYamlError(node, Err_BadNodeKind)
	}

	n := len(node.Content)
	for i := 0; i < n; i += 2 {
		if continue_, err := processor(i, node.Content[i].Value, node.Content[i+1]); err != nil {
			return err
		} else if !continue_ {
			return nil
		}
	}

	return nil
}

// ProcessMappingNode execute a function(``processor``) for each mapping node in a mapping node if ``node`` is
// not a ``MappingNode`` return ``Err_BadNodeKind``.
// if ``processor`` return an error, this function immediately return and ignore processing rest of the nodes.
func ProcessMappingNode(node *Node, processor func(index int, name string, node *Node) error) error {
	if node.Kind != MappingNode {
		return NewYamlError(node, Err_BadNodeKind)
	}

	n := len(node.Content)
	for i := 0; i < n; i += 2 {
		if err := processor(i, node.Content[i].Value, node.Content[i+1]); err != nil {
			return err
		}
	}

	return nil
}

// PopMappingNode pop a node with specified key from a ``MappingNode`` and return it to the caller.
// return ``Err_BadKind`` if node is not a ``MappingNode``.
func PopMappingNode(node *Node, key string) (*Node, error) {
	var res *Node
	err := ProcessMappingNode2(node, func(index int, name string, node *Node) (bool, error) {
		if name == key {
			res = node
			node.Content = append(node.Content[:index], node.Content[index+2:]...)
			return false, nil
		}

		return true, nil
	})

	return res, err
}

// PopMappingNodeTo pop a node with specified key from a ``MappingNode`` and decode it into ``target``.
// return ``Err_BadKind`` if node is not a ``MappingNode``.
// return ``Err_MissingRequiredNode`` if no node exists with specified ``key``.
func PopMappingNodeTo(node *Node, key string, target interface{}) error {
	if res, err := PopMappingNode(node, key); err != nil {
		return err
	} else if res == nil {
		return NewYamlErrorf(node, "missing required property(%s): %w", key, Err_MissingRequiredNode)
	} else if err = res.Decode(target); err != nil {
		return NewYamlError(res, err)
	} else {
		return nil
	}
}

func IsNullNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!null"
}
func IsStringNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!str"
}
func IsBoolNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!bool"
}
func IsIntNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!int"
}
func IsFloatNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!float"
}
func IsTimestampNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!timestamp"
}
func IsBinaryNode(node *Node) bool {
	return node.Kind == ScalarNode && node.ShortTag() == "!!binary"
}

func ToBool(node *Node) (res bool, err error) {
	err = node.Decode(&res)
	return
}
func ToInt(node *Node) (res int64, err error) {
	err = node.Decode(&res)
	return
}
func ToUint(node *Node) (res uint64, err error) {
	err = node.Decode(&res)
	return
}
func ToFloat(node *Node) (res float64, err error) {
	err = node.Decode(&res)
	return
}
func ToString(node *Node) (res string, err error) {
	err = node.Decode(&res)
	return
}

func CreateTagName(name string) string {
	return fmt.Sprintf("!<tag:github.com,2000:mehdi-roozitalab/yaml/%s>", name)
}
func IsTag(tag Tag, name string) bool {
	return core_utils.StringArrayContains(tag.Names(), name)
}

func CreateNodeFromTemplate(template *Node, kind Kind, tag, value string, content []*Node) *Node {
	node := *template
	node.Tag = tag
	node.Kind = kind
	node.Value = value
	node.Content = content
	return &node
}
func StringListToSequenceNode(template *Node, seq []string) *Node {
	result := CreateNodeFromTemplate(template, SequenceNode, "!!seq", "", make([]*Node, 0, len(seq)))
	for _, s := range seq {
		node := CreateNodeFromTemplate(template, ScalarNode, "!!str", s, nil)
		result.Content = append(result.Content, node)
	}
	return result
}
func StringToScalarNode(template *Node, s string) *Node {
	return CreateNodeFromTemplate(template, ScalarNode, "!!str", s, nil)
}
