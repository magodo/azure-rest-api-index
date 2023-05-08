package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

type Index struct {
	rootdir      string
	indexFixedRP map[OpLocatorFixedRP]CandidateOperations
	indexGlobRP  map[OpLocatorGlobRP]CandidateOperations
}

type OpLocatorFixedRP struct {
	// Upper cased RP name, e.g. MICROSOFT.COMPUTE. This might be "" for API path that has no explicit RP defined (e.g. /subscriptions/{subscriptionId})
	RP string
	OpLocatorGlobRP
}

type OpLocatorGlobRP struct {
	// API version, e.g. 2020-10-01-preview
	Version string
	// Upper cased resource type, e.g. /VIRTUALNETWORKS/SUBNETS
	RT string
	// Upper cased potential action/collection type, e.g. LISTKEYS (action), SUBNETS
	ACT string
	// HTTP operation kind, e.g. GET
	Method OperationKind
}

// CandidateOperations represents a map of path patterns that maps to the same operation locator.
type CandidateOperations map[PathPatternStr]OperationRefStr

// PathPatternStr represents an API path pattern, with all the fixed segment upper cased, and all the parameterized segment as a literal "{}", or "{*}" (for x-ms-skip-url-encode).
type PathPatternStr string

// The JSON reference to the operation, e.g. <dir>/foo.json#/paths/~1subscriptions~1{subscriptionId}~1providers~1{resourceProviderNamespace}~1register/post
type OperationRefStr string

func BuildIndex(rootdir string) (*Index, error) {
	index := &Index{
		rootdir:      rootdir,
		indexFixedRP: map[OpLocatorFixedRP]CandidateOperations{},
		indexGlobRP:  map[OpLocatorGlobRP]CandidateOperations{},
	}

	logger.Info("Collecting specs", "dir", rootdir)
	l, err := collectSpecs(rootdir)
	if err != nil {
		return nil, fmt.Errorf("collecting specs: %v", err)
	}
	logger.Info(fmt.Sprintf("%d specs collected", len(l)))

	logger.Info("Parsing specs")
	for _, spec := range l {
		mFix, mGlob, err := parseSpec(spec)
		if err != nil {
			return nil, fmt.Errorf("parsing spec %s: %v", spec, err)
		}
		for k, v := range mFix {
			if len(index.indexFixedRP[k]) == 0 {
				index.indexFixedRP[k] = CandidateOperations{}
			}
			index.indexFixedRP[k][v.PathPatternStr] = v.OperationRefStr
		}
		for k, v := range mGlob {
			if len(index.indexGlobRP[k]) == 0 {
				index.indexGlobRP[k] = CandidateOperations{}
			}
			index.indexGlobRP[k][v.PathPatternStr] = v.OperationRefStr
		}
	}

	return index, nil
}

// collectSpecs collects all Swagger specs based on the effective tags in each RP's readme.md.
func collectSpecs(rootdir string) ([]string, error) {
	var speclist []string

	if err := filepath.WalkDir(rootdir,
		func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if strings.EqualFold(d.Name(), "data-plane") {
					return filepath.SkipDir
				}
				if strings.EqualFold(d.Name(), "examples") {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name() != "readme.md" {
				return nil
			}
			content, err := os.ReadFile(p)
			if err != nil {
				return fmt.Errorf("reading file %s: %v", p, err)
			}
			l, err := SpecListFromReadmeMD(content)
			if err != nil {
				return fmt.Errorf("retrieving spec list from %s: %v", p, err)
			}
			for _, relp := range l {
				speclist = append(speclist, filepath.Join(filepath.Dir(p), relp))
			}
			return filepath.SkipDir
		}); err != nil {
		return nil, err
	}
	sort.Slice(speclist, func(i, j int) bool { return speclist[i] < speclist[j] })
	return speclist, nil
}

type OpInfo struct {
	PathPatternStr
	OperationRefStr
}

type OperationKind string

const (
	OperationKindGet     OperationKind = "GET"
	OperationKindPut                   = "PUT"
	OperationKindPost                  = "POST"
	OperationKindDelete                = "DELETE"
	OperationKindOptions               = "OPTIONS"
	OperationKindHead                  = "HEAD"
	OperationKindPatch                 = "PATCH"
)

var PossibleOperationKinds = []OperationKind{
	OperationKindGet,
	OperationKindPut,
	OperationKindPost,
	OperationKindDelete,
	OperationKindOptions,
	OperationKindHead,
	OperationKindPatch,
}

func PathItemOperation(pathItem spec.PathItem, op OperationKind) *spec.Operation {
	switch op {
	case OperationKindGet:
		return pathItem.Get
	case OperationKindPut:
		return pathItem.Put
	case OperationKindPost:
		return pathItem.Post
	case OperationKindDelete:
		return pathItem.Delete
	case OperationKindOptions:
		return pathItem.Options
	case OperationKindHead:
		return pathItem.Head
	case OperationKindPatch:
		return pathItem.Patch
	}
	return nil
}

// parseSpec parses one Swagger spec and returns back a per-spec index for it
func parseSpec(specpath string) (map[OpLocatorFixedRP]OpInfo, map[OpLocatorGlobRP]OpInfo, error) {
	doc, err := loads.Spec(specpath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading spec: %v", err)
	}
	swagger := doc.Spec()

	// Skipping swagger specs that have no "paths" defined
	if swagger.Paths == nil || len(swagger.Paths.Paths) == 0 {
		return nil, nil, nil
	}
	if swagger.Info == nil {
		return nil, nil, fmt.Errorf(`spec has no "Info"`)
	}
	if swagger.Info.Version == "" {
		return nil, nil, fmt.Errorf(`spec has no "Info.Version"`)
	}

	version := swagger.Info.Version
	infoMapFixedRP := map[OpLocatorFixedRP]OpInfo{}
	infoMapGlobRP := map[OpLocatorGlobRP]OpInfo{}
	for path, pathItem := range swagger.Paths.Paths {
		for _, opKind := range PossibleOperationKinds {
			if PathItemOperation(pathItem, opKind) == nil {
				continue
			}
			logger.Debug("Parsing spec", "spec", specpath, "path", path, "operation", opKind)
			pathPattern, err := ParsePathPatternFromSwagger(specpath, swagger, path, opKind)
			if err != nil {
				return nil, nil, fmt.Errorf("parsing path pattern for %s (%s): %v", path, opKind, err)
			}
			// path -> RP, RT, ACT
			// We look backwards for the first "providers" segment.
			providerIdx := -1
			for i := len(pathPattern.Segments) - 1; i >= 0; i-- {
				if strings.EqualFold(pathPattern.Segments[i].FixedName, "providers") {
					providerIdx = i
					break
				}
			}
			var (
				rp, rt, act string
				rpIsGlob    bool
			)

			if providerIdx == -1 {
				logger.Warn("no provider defined", "spec", specpath, "path", path, "operation", opKind)
			} else {
				// RP found, but can be glob
				providerSeg := pathPattern.Segments[providerIdx+1]
				rp = providerSeg.FixedName
				rpIsGlob = providerSeg.IsParameter
				lastIdx := len(pathPattern.Segments)
				if len(pathPattern.Segments[providerIdx:])%2 == 1 {
					seg := pathPattern.Segments[len(pathPattern.Segments)-1]
					if seg.IsParameter {
						return nil, nil, fmt.Errorf("action-like segment is parameterized, in %s (%s)", path, opKind)
					}
					lastIdx = lastIdx - 1
					act = seg.FixedName
				}
				var rts []string
				for i := providerIdx + 2; i < lastIdx; i += 2 {
					seg := pathPattern.Segments[i]
					if seg.IsParameter {
						return nil, nil, fmt.Errorf("resource type %dth segment is parameterized, in %s (%s)", i, path, opKind)
					}
					rts = append(rts, seg.FixedName)
				}
				rt = "/" + strings.Join(rts, "/")
			}

			absSpecPath, err := filepath.Abs(specpath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get abs path for %s: %v", specpath, err)
			}
			opRefStr := absSpecPath + "#" + jsonpointer.Escape(path) + "/" + strings.ToLower(string(opKind))

			info := OpInfo{
				PathPatternStr:  PathPatternStr(pathPattern.String()),
				OperationRefStr: OperationRefStr(opRefStr),
			}

			opLocGlobRP := OpLocatorGlobRP{
				Version: version,
				RT:      strings.ToUpper(rt),
				ACT:     strings.ToUpper(act),
				Method:  opKind,
			}

			if rpIsGlob {
				if exist, ok := infoMapGlobRP[opLocGlobRP]; ok {
					return nil, nil, fmt.Errorf("operation locator %#v already applied to %#v, conflicts to the new %#v", opLocGlobRP, exist, info)
				}
				infoMapGlobRP[opLocGlobRP] = info
			} else {
				opLocFixedRP := OpLocatorFixedRP{
					RP:              strings.ToUpper(rp),
					OpLocatorGlobRP: opLocGlobRP,
				}
				if exist, ok := infoMapFixedRP[opLocFixedRP]; ok {
					return nil, nil, fmt.Errorf("operation locator %#v already applied to %#v, conflicts to the new %#v", opLocFixedRP, exist, info)
				}
				infoMapFixedRP[opLocFixedRP] = info
			}
		}
	}
	return infoMapFixedRP, infoMapGlobRP, nil
}

func sortedKeys[K ~string, V any](input map[K]V) []K {
	keys := make([]K, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	return keys
}
