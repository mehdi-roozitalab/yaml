package yaml

type Tag interface {
	Names() []string
	Resolve(loader *Loader, node *Node) (*Node, error)
}
