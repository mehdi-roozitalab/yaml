package yaml

import (
	"os"

	"github.com/mehdi-roozitalab/core_utils"
)

var (
	includeNames      = []string{CreateTagName("include"), "!include"}
	includePathReader = StringListReader{
		ItemSeparators: []string{"|", "&"},
		ItemChild:      "items",
		ExtraNodeParser: func(reader *StringListReader, list *StringList, nodeName string, node *Node) error {
			if nodeName == "all" {
				if b, err := ToBool(node); err != nil {
					return err
				} else {
					list.Data["all"] = b
					return nil
				}
			} else {
				return Err_InvalidChild
			}
		},
	}
)

type IncludeTag struct{}

func (tag IncludeTag) Names() []string { return includeNames }
func (tag IncludeTag) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	fl := fragmentLoader{SourceNode: node, Loader: loader}
	if err := fl.ReadOptions(); err != nil {
		return nil, err
	} else {
		return fl.Resolve()
	}
}

type includeFragment struct {
	node *Node
}

func (f *includeFragment) UnmarshalYAML(node *Node) error {
	f.node = node
	return nil
}

type fragmentLoader struct {
	Loader           *Loader
	SourceNode       *Node
	LoadedNodes      []*Node
	LoadedPaths      []string
	IncludeList      *StringList
	ShouldIncludeAll core_utils.Bool3
}

func (fl *fragmentLoader) ReadIncludePaths() error {
	includeList, failedNode, err := includePathReader.ReadStringList(fl.Loader, fl.SourceNode)
	if err != nil {
		return NewYamlError(failedNode, err)
	}
	fl.IncludeList = includeList
	return nil
}
func (fl *fragmentLoader) ValidateIncludePaths() error {
	if len(fl.IncludeList.Values) == 0 {
		return NewYamlConstError(fl.SourceNode, "at least one include path is required")
	}
	for _, path := range fl.IncludeList.Values {
		if path.Value == "" {
			return NewYamlConstError(path.Node, "empty include path is not valid")
		}
	}
	return nil
}
func (fl *fragmentLoader) ReadShouldIncludeAll() core_utils.Bool3 {
	if allv, ok := fl.IncludeList.Data["all"]; ok {
		return core_utils.B3FromBool(allv.(bool))
	} else if fl.IncludeList.UsedSeparator != "" {
		return core_utils.B3FromBool(fl.IncludeList.UsedSeparator == "&")
	} else {
		return core_utils.B3Null
	}
}
func (fl *fragmentLoader) IsPathLoaded(path string) bool {
	return core_utils.StringArrayContains(fl.LoadedPaths, path)
}
func (fl *fragmentLoader) LoadPath(node *Node, path string) error {
	var f includeFragment
	if err := fl.Loader.LoadPath(path, &f); err == nil {
		fl.LoadedPaths = append(fl.LoadedPaths, path)
		fl.LoadedNodes = append(fl.LoadedNodes, f.node)
	} else if !os.IsNotExist(err) || fl.ShouldIncludeAll.IsTrue() {
		return NewYamlErrorf(node, "failed to load the file from %s: %w", path, err)
	}
	return nil
}

func (fl *fragmentLoader) ReadOptions() error {
	if err := fl.ReadIncludePaths(); err != nil {
		return err
	} else if err = fl.ValidateIncludePaths(); err != nil {
		return err
	} else {
		fl.ShouldIncludeAll = fl.ReadShouldIncludeAll()
		return nil
	}
}

func (fl *fragmentLoader) GetResult() (*Node, error) {
	if len(fl.LoadedNodes) == 0 {
		if fl.ShouldIncludeAll.IsTrue() {
			return nil, NewYamlConstError(fl.SourceNode, "failed to load any of the included path")
		}
		fl.SourceNode.Value = ""
		fl.SourceNode.Tag = "!!nil"
		fl.SourceNode.Content = fl.LoadedNodes
	} else if (!fl.ShouldIncludeAll.HaveValue() && len(fl.LoadedNodes) == 1) || fl.ShouldIncludeAll.IsFalse() {
		return fl.LoadedNodes[0], nil
	} else {
		fl.SourceNode.Kind = SequenceNode
		fl.SourceNode.Value = ""
		fl.SourceNode.Tag = "!!seq"
		fl.SourceNode.Content = fl.LoadedNodes
	}
	return fl.SourceNode, nil
}
func (fl *fragmentLoader) Resolve() (*Node, error) {
	for _, n := range fl.IncludeList.Values {
		if !fl.IsPathLoaded(n.Value) {
			if err := fl.LoadPath(n.Node, n.Value); err != nil {
				return nil, err
			}
		}
	}
	return fl.GetResult()
}
