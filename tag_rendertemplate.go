package yaml

import (
	"github.com/mehdi-roozitalab/template"
)

var renderTemplateNames = []string{CreateTagName("render_template"), "!render_template", "!t"}

type RenderTemplateTag struct{}

func (tag RenderTemplateTag) Names() []string { return renderTemplateNames }
func (tag RenderTemplateTag) Resolve(loader *Loader, node *Node) (*Node, error) {
	if !IsTag(tag, node.Tag) {
		return node, nil
	}

	if node.Kind != ScalarNode {
		return nil, NewYamlErrorf(node, "%s tag may only applied to string values", node.Tag)
	} else if tmpl, err := template.ParseTextTemplate(node.Value); err != nil {
		return nil, NewYamlErrorf(node, "failed to parse template: %w", err)
	} else if s, err := tmpl.Render(loader.Variables); err != nil {
		return nil, NewYamlErrorf(node, "failed to render template: %w", err)
	} else {
		return StringToScalarNode(node, s), nil
	}
}
