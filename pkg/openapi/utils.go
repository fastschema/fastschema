package openapi

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/ogen-go/ogen"
)

// CreateParameters creates openapi parameters from app args
func CreateParameters(args app.Args, params []string) ([]*ogen.Parameter, error) {
	parameters := make([]*ogen.Parameter, 0, len(args))
	for argName, arg := range args {
		example, err := json.Marshal(arg.Example)
		if err != nil {
			return nil, err
		}

		parameter := &ogen.Parameter{
			Name:        argName,
			In:          "query",
			Description: arg.Description,
			Required:    arg.Required,
			Schema: &ogen.Schema{
				Type:   arg.Type.Common(),
				Format: arg.Type.String(),
			},
			Examples: map[string]*ogen.Example{
				argName: {
					Summary: argName + " example",
					Value:   example,
				},
			},
		}

		if utils.Contains(params, argName) {
			parameter.In = "path"
			parameter.Required = true
		}

		parameters = append(parameters, parameter)
	}

	return parameters, nil
}

// MergeParameters merges two slices of parameters
func MergeParameters(first []*ogen.Parameter, second []*ogen.Parameter) []*ogen.Parameter {
	result := make([]*ogen.Parameter, 0, len(first)+len(second))
	result = append(result, first...)
	result = append(result, second...)
	return result
}

// MergeArgs merges two app.Args maps and returns a new app.Args map.
//
//	If there are duplicate keys in the second map, an error is returned.
func MergeArgs(first app.Args, second app.Args) (app.Args, error) {
	result := first.Clone()
	for key, value := range second {
		if _, ok := result[key]; ok {
			return nil, errors.InternalServerError("duplicate key %s in args", key)
		}

		result[key] = value
	}

	return result, nil
}

// MergePathItems merges two maps of path items into a single map.
//
//	It takes two maps of type map[string]*ogen.PathItem as input and returns a new map of type map[string]*ogen.PathItem.
//	The function iterates over the second map and checks if each key exists in the first map.
//	If the key exists, it merges the corresponding path item properties from the second map into the first map.
//	If the key does not exist, it adds the key-value pair from the second map to the first map.
//	The resulting map contains all the path items from both input maps, with any overlapping keys merged.
func MergePathItems(first map[string]*ogen.PathItem, second map[string]*ogen.PathItem) map[string]*ogen.PathItem {
	result := make(map[string]*ogen.PathItem, len(first))
	for key, value := range first {
		result[key] = value
	}

	for key, value := range second {
		if _, ok := result[key]; ok {
			if value.Get != nil {
				result[key].Get = value.Get
			}

			if value.Put != nil {
				result[key].Put = value.Put
			}

			if value.Post != nil {
				result[key].Post = value.Post
			}

			if value.Delete != nil {
				result[key].Delete = value.Delete
			}

			if value.Options != nil {
				result[key].Options = value.Options
			}

			if value.Head != nil {
				result[key].Head = value.Head
			}

			if value.Patch != nil {
				result[key].Patch = value.Patch
			}

			if value.Trace != nil {
				result[key].Trace = value.Trace
			}
		} else {
			result[key] = value
		}
	}

	return result
}

// JoinPaths concatenates multiple path segments into a single path string.
//
//	It removes any duplicate slashes in the resulting path.
func JoinPaths(paths ...string) string {
	if len(paths) > 0 && paths[len(paths)-1] == "" {
		paths = paths[:len(paths)-1]
	}

	return regexp.MustCompile(`/{2,}`).ReplaceAllString(strings.Join(paths, "/"), "/")
}

// NormalizePath normalizes the given path by converting it into the OpenAPI format.
//
//	It replaces any path parameters in the form of ":param" with "{param}".
//	The function returns the normalized path, a slice of extracted path parameters, and any error encountered.
func NormalizePath(path string) (string, []string) {
	params := ExtractPathParams(path)

	// path is in form of /path/:id/other
	// we need to convert it into openapi format /path/{id}/other
	for _, param := range params {
		path = strings.ReplaceAll(path, ":"+param, "{"+param+"}")
	}

	return path, params
}

// ExtractPathParams extracts path parameters from a given string.
//
//	It uses a regular expression to find all occurrences of ":param" pattern
//	and returns a slice of the captured parameter names.
func ExtractPathParams(s string) []string {
	re := regexp.MustCompile(`:(\w+)`)
	matches := re.FindAllStringSubmatch(s, -1)

	var words []string
	for _, match := range matches {
		words = append(words, match[1]) // match[1] is the captured group
	}

	return words
}

type methodPath struct {
	method string
	path   string
}

// FlattenResources takes a slice of resources, a prefix string, and app.Args as input,
//
//	and returns a flattened list of ResourceInfo structs and an error.
//	It recursively flattens the resources and generates a path for each method in the resource.
//	If a resource is a group, it appends the group prefix to the path.
//	If a resource has no method, it uses "GET" as the default method.
//	The ResourceInfo struct contains information about the resource, such as ID, signature, method, path, args, and public status.
//	If there is an error during the flattening process, it returns nil and the error.
func FlattenResources(resources []*app.Resource, prefix string, args app.Args) ([]*ResourceInfo, error) {
	var infos []*ResourceInfo
	for _, resource := range resources {
		meta := resource.Meta()
		if meta == nil {
			meta = &app.Meta{}
		}

		args, err := MergeArgs(args, meta.Args)
		if err != nil {
			return nil, err
		}

		if resource.IsGroup() {
			groupBasePath := resource.Name()
			if meta.Prefix != "" {
				groupBasePath = meta.Prefix
			}

			groupPrefix := JoinPaths(prefix, groupBasePath)
			subInfos, err := FlattenResources(resource.Resources(), groupPrefix, args)
			if err != nil {
				return nil, err
			}

			infos = append(infos, subInfos...)
			continue
		}

		// resource may contains many http methods, we need to create a path for each method
		noMethod := true
		methodPaths := []methodPath{
			{method: "GET", path: meta.Get},
			{method: "HEAD", path: meta.Head},
			{method: "POST", path: meta.Post},
			{method: "PUT", path: meta.Put},
			{method: "DELETE", path: meta.Delete},
			{method: "CONNECT", path: meta.Connect},
			{method: "OPTIONS", path: meta.Options},
			{method: "TRACE", path: meta.Trace},
			{method: "PATCH", path: meta.Patch},
		}

		for _, method := range methodPaths {
			if method.path == "" {
				continue
			}

			noMethod = false
			path := JoinPaths(prefix, method.path)
			infos = append(infos, &ResourceInfo{
				ID:         resource.ID(),
				Signatures: resource.Signature(),
				Method:     method.method,
				Path:       path,
				Args:       args,
				Public:     resource.IsPublic(),
			})
		}

		// if there is no method, use get as default
		if noMethod {
			path := JoinPaths(prefix, resource.Name())
			infos = append(infos, &ResourceInfo{
				ID:         resource.ID(),
				Signatures: resource.Signature(),
				Method:     "GET",
				Path:       path,
				Args:       args,
				Public:     resource.IsPublic(),
			})
		}
	}

	return infos, nil
}

// RefSchema returns a new instance of ogen.Schema with the given name as the reference.
//
//	The reference is constructed using the provided name and the "#/components/schemas/" prefix.
func RefSchema(name string) *ogen.Schema {
	return &ogen.Schema{Ref: "#/components/schemas/" + name}
}
