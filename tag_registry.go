package yaml

import (
	"fmt"
	"sync"

	"github.com/mehdi-roozitalab/core_utils"
)

type TagRegistry interface {
	RegisterTags(tag ...Tag) error
	GetTagByName(name string) Tag
}

type simpleTagRegistry struct {
	tags map[string]Tag
}

func NewSimpleTagRegistry() TagRegistry {
	return &simpleTagRegistry{
		tags: map[string]Tag{},
	}
}

func (r *simpleTagRegistry) RegisterTags(tag ...Tag) error {
	for _, item := range tag {
		for _, s := range item.Names() {
			if _, ok := r.tags[s]; ok {
				return fmt.Errorf("another tag with same name(%s) already exists", s)
			}
		}
	}

	for _, item := range tag {
		for _, s := range item.Names() {
			r.tags[s] = item
		}
	}

	return nil
}
func (r *simpleTagRegistry) GetTagByName(name string) Tag {
	if tag, ok := r.tags[name]; ok {
		return tag
	}
	return nil
}

type threadSafeTagRegistry struct {
	lock sync.Mutex
	base TagRegistry
}

func NewThreadSafeTagRegistry(baseRegistry TagRegistry) TagRegistry {
	if baseRegistry == nil {
		panic(core_utils.ConstError("missing base registry"))
	}

	return &threadSafeTagRegistry{
		base: baseRegistry,
	}
}

func (r *threadSafeTagRegistry) RegisterTags(tag ...Tag) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.base.RegisterTags(tag...)
}
func (r *threadSafeTagRegistry) GetTagByName(name string) Tag {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.base.GetTagByName(name)
}

type childTagRegistry struct {
	child  TagRegistry
	parent TagRegistry
}

func NewChildRegistry(parent, child TagRegistry) TagRegistry {
	if parent == nil {
		panic(core_utils.ConstError("midding parent"))
	}
	if child == nil {
		child = NewSimpleTagRegistry()
	}
	return &childTagRegistry{
		parent: parent,
		child:  child,
	}
}

func (r *childTagRegistry) RegisterTags(tag ...Tag) error {
	for _, item := range tag {
		for _, s := range item.Names() {
			if r.GetTagByName(s) != nil {
				return fmt.Errorf("another tag with same name(%s) already exists", s)
			}
		}
	}

	return r.child.RegisterTags(tag...)
}
func (r *childTagRegistry) GetTagByName(name string) Tag {
	if tag := r.child.GetTagByName(name); tag != nil {
		return tag
	}
	return r.parent.GetTagByName(name)
}

var defaultTagRegistry = NewThreadSafeTagRegistry(NewSimpleTagRegistry())

func DefaultTagRegistry() TagRegistry { return defaultTagRegistry }
