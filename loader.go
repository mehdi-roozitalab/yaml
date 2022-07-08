package yaml

import (
	"os"

	"github.com/mehdi-roozitalab/core_utils"
)

type Loader struct {
	registry  TagRegistry
	Variables map[string]interface{}
}

func NewLoader(registry TagRegistry) *Loader {
	return &Loader{
		registry:  registry,
		Variables: map[string]interface{}{},
	}
}

func (loader *Loader) GetTagRegistry() TagRegistry { return loader.registry }
func (loader *Loader) ResolveTags(node *Node) (*Node, error) {
	if tag := loader.registry.GetTagByName(node.Tag); tag != nil {
		if resolved, err := tag.Resolve(loader, node); err != nil {
			return nil, err
		} else if resolved == nil {
			return nil, nil
		} else if err = loader.resolveChildTags(resolved); err != nil {
			return nil, err
		} else {
			return resolved, nil
		}
	} else if err := loader.resolveChildTags(node); err != nil {
		return nil, err
	}
	return node, nil
}

func (loader *Loader) Load(content []byte, target interface{}, filename string) error {
	cl := contentLoader{
		filename: filename,
		loader:   loader,
		target:   target,
	}
	return UnmarshalYaml(content, &cl)
}
func (loader *Loader) LoadPath(path string, target interface{}) error {
	if content, err := os.ReadFile(path); err != nil {
		return err
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		defer os.Chdir(wd)

		fullpath, err := core_utils.AbsolutePath(path)
		if err != nil {
			return err
		}

		err = loader.Load(content, target, fullpath)
		return err
	}
}

func (loader *Loader) resolveChildTags(node *Node) error {
	if node.Kind == MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			if ch, err := loader.ResolveTags(node.Content[i+1]); err != nil {
				return err
			} else if ch == nil {
				node.Content = append(node.Content[:i], node.Content[i+2:]...)
				i -= 2
			} else {
				node.Content[i+1] = ch
			}
		}
	} else {
		for i := 0; i < len(node.Content); i++ {
			if ch, err := loader.ResolveTags(node.Content[i]); err != nil {
				return err
			} else if ch == nil {
				node.Content = append(node.Content[:i], node.Content[i+1:]...)
				i -= 1
			} else {
				node.Content[i] = ch
			}
		}
	}
	return nil
}

func Unmarshal(content []byte, target interface{}) error {
	loader := NewLoader(NewChildRegistry(DefaultTagRegistry(), NewSimpleTagRegistry()))
	return loader.Load(content, target, "<input>")
}
func UnmarshalPath(path string, target interface{}) error {
	loader := NewLoader(NewChildRegistry(DefaultTagRegistry(), NewSimpleTagRegistry()))
	return loader.LoadPath(path, target)
}
