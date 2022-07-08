package yaml

import (
	"os"

	"github.com/mehdi-roozitalab/core_utils"
)

var (
	fileNames       = []string{CreateTagName("file"), "!file"}
	fileItemsReader = StringListReader{
		ItemSeparators:   []string{"|"},
		DefaultSeparator: ':',
		ItemChild:        "items",
		DefaultChild:     "default",
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

const err_NoFile = core_utils.ConstError("none of the files exists and no default value is provided")

// FileTag tag that will applied to a string or a sequence of strings and will read the content of the
// file or files.
// By default it will read first existing file and return a ScalarNode of type string but you may set
// `all` to true to force it read all the files and return a sequence of strings
type FileTag struct{}

func (tag FileTag) Names() []string { return fileNames }
func (tag FileTag) Resolve(loader *Loader, node *Node) (*Node, error) {
	if IsTag(tag, node.Tag) {
		return node, nil
	}

	reader := fileReader{
		Loader:     loader,
		SourceNode: node,
	}
	if err := reader.ReadOptions(); err != nil {
		return nil, err
	} else {
		return reader.Resolve()
	}
}

type fileReader struct {
	Loader        *Loader
	SourceNode    *Node
	Files         *StringList
	ShouldReadAll bool
	ReadFiles     []string
}

func (f *fileReader) ReadFileNames() error {
	includeList, failedNode, err := includePathReader.ReadStringList(f.Loader, f.SourceNode)
	if err != nil {
		return NewYamlError(failedNode, err)
	}
	f.Files = includeList
	return nil
}
func (f *fileReader) ValidateFiles() error {
	if len(f.Files.Values) == 0 {
		return NewYamlConstError(f.SourceNode, "at least one path is required")
	}
	for _, path := range f.Files.Values {
		if path.Value == "" {
			return NewYamlConstError(path.Node, "empty path is not valid")
		}
	}
	return nil
}
func (f *fileReader) ReadShouldReadAll() bool {
	if allv, ok := f.Files.Data["all"]; ok {
		return allv.(bool)
	} else {
		return false
	}
}

func (f *fileReader) ReadOptions() error {
	if err := f.ReadFileNames(); err != nil {
		return err
	} else if err = f.ValidateFiles(); err != nil {
		return err
	} else {
		f.ShouldReadAll = f.ReadShouldReadAll()
		return nil
	}
}
func (f *fileReader) GetResult() (*Node, error) {
	if f.ShouldReadAll {
		return StringListToSequenceNode(f.SourceNode, f.ReadFiles), nil
	} else {
		return StringToScalarNode(f.SourceNode, f.ReadFiles[0]), nil
	}
}
func (f *fileReader) LoadDefault() (*Node, error) {
	if f.Files.DefaultValue != nil {
		if f.Files.DefaultValue.Node != nil {
			var s string
			if resolved, err := f.Loader.ResolveTags(f.Files.DefaultValue.Node); err != nil {
				return nil, err
			} else if err = resolved.Decode(&s); err != nil {
				return nil, NewYamlError(resolved, err)
			} else {
				f.ReadFiles = []string{s}
			}
		} else {
			f.ReadFiles = []string{f.Files.DefaultValue.Value}
		}
		return f.GetResult()
	}

	return nil, NewYamlError(f.SourceNode, err_NoFile)
}
func (f *fileReader) Resolve() (*Node, error) {
	for _, file := range f.Files.Values {
		if content, err := os.ReadFile(file.Value); err != nil {
			if !os.IsNotExist(err) || f.ShouldReadAll {
				return nil, NewYamlErrorf(file.Node, "failed to read the file at %q: %w", file.Value, err)
			}
		} else {
			f.ReadFiles = append(f.ReadFiles, string(content))
			if !f.ShouldReadAll {
				return f.GetResult()
			}
		}
	}
	if len(f.ReadFiles) == 0 {
		return f.LoadDefault()
	}
	return f.GetResult()
}
