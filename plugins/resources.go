package plugins

import (
	"net/http"

	"github.com/dop251/goja"
	"github.com/fastschema/fastschema/fs"
)

type Resource struct {
	FsResource *fs.Resource
	program    *Program
	set        map[string]any
}

func NewResource(
	fsResource *fs.Resource,
	program *Program,
	set map[string]any,
) *Resource {
	return &Resource{
		FsResource: fsResource,
		program:    program,
		set:        set,
	}
}

func (r *Resource) Find(resourceID string) *Resource {
	fsResource := r.FsResource.Find(resourceID)
	if fsResource == nil {
		return nil
	}

	return &Resource{
		program:    r.program,
		set:        r.set,
		FsResource: fsResource,
	}
}
func (r *Resource) Group(name string, metas ...*fs.Meta) *Resource {
	return &Resource{
		program:    r.program,
		set:        r.set,
		FsResource: r.FsResource.Group(name, metas...),
	}
}

func (r *Resource) Add(handler goja.Value, metas ...*fs.Meta) (*Resource, error) {
	return r, r.program.WithFuncName(handler, func(fnName string) {
		r.FsResource.Add(fs.NewResource(fnName, func(c fs.Context, _ any) (_ any, err error) {
			result, err := r.program.CallFunc(fnName, r.set, c)
			if err != nil {
				return nil, err
			}

			if outputObj, ok := result.(map[string]any); ok {
				if outputObj["__name__"] == "HtmlResponse" {
					header := make(http.Header)
					header.Set("Content-Type", "text/html")
					return &fs.HTTPResponse{
						StatusCode: int(outputObj["status"].(int64)),
						Body:       []byte(outputObj["html"].(string)),
						Header:     header,
					}, nil
				}
			}

			return result, nil
		}, metas...))
	})
}
