package main

import (
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
)

type PathPattern struct {
	Segments []PathSegment
}

type PathSegment struct {
	FixedName   string
	IsParameter bool
	IsMulti     bool // indicates the x-ms-skip-url-encoding = true
}

func ParsePathPatternFromSwagger(specFile string, swagger *spec.Swagger, path string, operation OperationKind) (*PathPattern, error) {
	if swagger.Paths == nil {
		return nil, fmt.Errorf(`no "paths"`)
	}
	pathItem, ok := swagger.Paths.Paths[path]
	if !ok {
		return nil, fmt.Errorf(`no path %s found`, path)
	}
	parameterMap := map[string]spec.Parameter{}
	for _, param := range pathItem.Parameters {
		if param.Ref.String() != "" {
			pparam, err := spec.ResolveParameterWithBase(swagger, param.Ref, &spec.ExpandOptions{RelativeBase: specFile})
			if err != nil {
				return nil, fmt.Errorf("resolving ref %q: %v", param.Ref.String(), err)
			}
			param = *pparam
		}
		parameterMap[param.Name] = param
	}
	// Per operation parameter overrides the per path parameter
	if op := PathItemOperation(pathItem, operation); op != nil {
		for _, param := range op.Parameters {
			if param.Ref.String() != "" {
				pparam, err := spec.ResolveParameterWithBase(swagger, param.Ref, &spec.ExpandOptions{RelativeBase: specFile})
				if err != nil {
					return nil, fmt.Errorf("resolving ref %q: %v", param.Ref.String(), err)
				}
				param = *pparam
			}
			parameterMap[param.Name] = param
		}
	}

	var segments []PathSegment
	for _, seg := range strings.Split(strings.TrimLeft(path, "/"), "/") {
		if isParameterizedSegment(seg) {
			name := strings.Trim(seg, "{}")
			param, ok := parameterMap[name]
			if !ok {
				return nil, fmt.Errorf("undefined parameter name %q", name)
			}
			segment := PathSegment{
				IsParameter: true,
			}
			if v, ok := param.VendorExtensible.Extensions["x-ms-skip-url-encoding"]; ok && v.(bool) {
				segment.IsMulti = true
			}
			segments = append(segments, segment)
		} else {
			segments = append(segments, PathSegment{FixedName: seg})
		}
	}
	return &PathPattern{Segments: segments}, nil
}

func ParsePathPatternFromString(path string) *PathPattern {
	var segments []PathSegment
	for _, seg := range strings.Split(strings.TrimLeft(path, "/"), "/") {
		switch seg {
		case "{}":
			segments = append(segments, PathSegment{IsParameter: true})
		case "{*}":
			segments = append(segments, PathSegment{IsParameter: true, IsMulti: true})
		default:
			segments = append(segments, PathSegment{FixedName: seg})
		}
	}
	return &PathPattern{Segments: segments}
}

func isParameterizedSegment(seg string) bool {
	return strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
}

func (p PathPattern) String() string {
	var segs []string
	for _, seg := range p.Segments {
		if !seg.IsParameter {
			segs = append(segs, seg.FixedName)
			continue
		}
		if seg.IsMulti {
			segs = append(segs, "{*}")
		} else {
			segs = append(segs, "{}")
		}
	}
	return "/" + strings.Join(segs, "/")
}
