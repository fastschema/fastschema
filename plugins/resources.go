package plugins

import (
	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/fs"
)

type Resource struct {
	fsResource *fs.Resource
	program    *Program
	set        map[string]any
}

func NewResource(
	fsResource *fs.Resource,
	program *Program,
	set map[string]any,
) *Resource {
	return &Resource{
		fsResource: fsResource,
		program:    program,
		set:        set,
	}
}

func (r *Resource) Group(name string, metas ...*fs.Meta) *Resource {
	return &Resource{
		program:    r.program,
		set:        r.set,
		fsResource: r.fsResource.Group(name, metas...),
	}
}

func (r *Resource) Add(handler goja.Value, metas ...*fs.Meta) (*Resource, error) {
	return r, r.program.WithFuncName(handler, func(fnName string) {
		r.fsResource.Add(fs.NewResource(fnName, func(c fs.Context, _ any) (_ any, err error) {
			return r.program.CallFunc(fnName, r.set, c)
		}, metas...))
	})
}
