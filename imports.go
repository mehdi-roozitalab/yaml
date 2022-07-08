package yaml

import y "gopkg.in/yaml.v3"

type Node = y.Node
type Kind = y.Kind

const (
	DocumentNode = y.DocumentNode
	SequenceNode = y.SequenceNode
	MappingNode  = y.MappingNode
	ScalarNode   = y.ScalarNode
	AliasNode    = y.AliasNode
)

func UnmarshalYaml(data []byte, target interface{}) error { return y.Unmarshal(data, target) }
func MarshalYaml(data interface{}) ([]byte, error)        { return y.Marshal(data) }
