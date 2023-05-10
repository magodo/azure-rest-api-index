package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

type Index struct {
	rootdir string
	ops     map[OpLocator]OperationRefs
}

const Wildcard = "*"

type OpLocator struct {
	// Upper cased RP name, e.g. MICROSOFT.COMPUTE. This might be "" for API path that has no explicit RP defined (e.g. /subscriptions/{subscriptionId})
	// This can be "*" to indicate it maps any RP
	RP string
	// API version, e.g. 2020-10-01-preview
	Version string
	// Upper cased resource type, e.g. /VIRTUALNETWORKS/SUBNETS
	RT string
	// Upper cased potential action/collection type, e.g. LISTKEYS (action), SUBNETS
	ACT string
	// HTTP operation kind, e.g. GET
	Method OperationKind
}

// OperationRefs represents a set of operation defintion (in form of JSON reference) that are mapped by the same operation locator.
// Since for a given operation locator, there might maps to multiple operation definition, only differing by the contained path pattern, there fore the actual operation ref is keyed by the containing path pattern.
// The value is a JSON reference to the operation, e.g. <dir>/foo.json#/paths/~1subscriptions~1{subscriptionId}~1providers~1{resourceProviderNamespace}~1register/post
type OperationRefs map[PathPatternStr]jsonreference.Ref

// PathPatternStr represents an API path pattern, with all the fixed segment upper cased, and all the parameterized segment as a literal "{}", or "{*}" (for x-ms-skip-url-encoding).
type PathPatternStr string

func BuildIndex(rootdir string, dedupFile string) (*Index, error) {
	index := &Index{
		rootdir: rootdir,
		ops:     map[OpLocator]OperationRefs{},
	}

	var deduplicator Deduplicator

	if dedupFile != "" {
		b, err := os.ReadFile(dedupFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %v", dedupFile, err)
		}
		var records DeduplicateRecords
		if err := json.Unmarshal(b, &records); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %v", dedupFile, err)
		}
		deduplicator, err = records.ToDeduplicator()
		if err != nil {
			return nil, fmt.Errorf("converting the dedup file: %v", err)
		}
	}

	logger.Info("Collecting specs", "dir", rootdir)
	l, err := collectSpecs(rootdir)
	if err != nil {
		return nil, fmt.Errorf("collecting specs: %v", err)
	}
	logger.Info(fmt.Sprintf("%d specs collected", len(l)))

	logger.Info("Parsing specs")

	type dupkey struct {
		OpLocator
		PathPatternStr
	}

	dups := map[dupkey][]jsonreference.Ref{}

	for _, spec := range l {
		m, err := parseSpec(spec)
		if err != nil {
			return nil, fmt.Errorf("parsing spec %s: %v", spec, err)
		}
		for k, mm := range m {
			if len(index.ops[k]) == 0 {
				index.ops[k] = OperationRefs{}
			}
			for ppattern, ref := range mm {
				if exist, ok := index.ops[k][ppattern]; ok {
					// Temporarily record duplicate operation definitions and resolve it later
					k := dupkey{
						OpLocator:      k,
						PathPatternStr: ppattern,
					}
					if len(dups[k]) == 0 {
						dups[k] = append(dups[k], exist)
					}
					dups[k] = append(dups[k], ref)
					continue
				}
				index.ops[k][ppattern] = ref
			}
		}
	}

	// resolving any duplicates
	if deduplicator != nil {
		for k, refs := range dups {
			var dedupOp *DedupOp
			for matcher, op := range deduplicator {
				op := op
				if matcher.Match(k.OpLocator, string(k.PathPatternStr)) {
					if dedupOp != nil {
						panic(fmt.Sprintf("Duplicate matchers in duplicator that match %s", k))
					}
					dedupOp = &op
				}
			}

			var refStrs []string
			for _, ref := range refs {
				refStrs = append(refStrs, ref.String())
			}

			if dedupOp != nil {
				if picker := dedupOp.Picker; picker != nil {
					var pickCnt int
					var pickRef jsonreference.Ref
					for _, ref := range refs {
						if picker.Match(ref) {
							pickCnt++
							pickRef = ref
						}
					}

					if pickCnt == 0 {
						panic(fmt.Sprintf("Nothing get deduplicated for %s. refs: %v", k, refStrs))
					}

					if pickCnt > 1 {
						panic(fmt.Sprintf("Still have duplicates after deduplicating %s. refs: %v", k, refStrs))
					}
					index.ops[k.OpLocator][k.PathPatternStr] = pickRef
					continue
				} else if dedupOp.Ignore {
					delete(index.ops[k.OpLocator], k.PathPatternStr)
					if len(index.ops[k.OpLocator]) == 0 {
						delete(index.ops, k.OpLocator)
					}
					continue
				}
			}

			logger.Warn("duplicate definition", "oploc", k.OpLocator, "path", k.PathPatternStr, "refs", refStrs)
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
func parseSpec(specpath string) (map[OpLocator]OperationRefs, error) {
	doc, err := loads.Spec(specpath)
	if err != nil {
		return nil, fmt.Errorf("loading spec: %v", err)
	}
	swagger := doc.Spec()

	// Skipping swagger specs that have no "paths" defined
	if swagger.Paths == nil || len(swagger.Paths.Paths) == 0 {
		return nil, nil
	}
	if swagger.Info == nil {
		return nil, fmt.Errorf(`spec has no "Info"`)
	}
	if swagger.Info.Version == "" {
		return nil, fmt.Errorf(`spec has no "Info.Version"`)
	}

	version := swagger.Info.Version
	infoMap := map[OpLocator]OperationRefs{}
	for path, pathItem := range swagger.Paths.Paths {
		for _, opKind := range PossibleOperationKinds {
			if PathItemOperation(pathItem, opKind) == nil {
				continue
			}
			logger.Debug("Parsing spec", "spec", specpath, "path", path, "operation", opKind)
			pathPatterns, err := ParsePathPatternFromSwagger(specpath, swagger, path, opKind)
			if err != nil {
				return nil, fmt.Errorf("parsing path pattern for %s (%s): %v", path, opKind, err)
			}
			for _, pathPattern := range pathPatterns {
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
					// TODO: ignore the too general ones, but keep the implicit RP Microsoft.Resources
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
							// TODO: shall we resolve some of these violations
							logger.Warn("action-like segment is parameterized", "path", path, "operation", opKind)
							continue
							//return nil, nil, fmt.Errorf("action-like segment is parameterized, in %s (%s)", path, opKind)
						}
						lastIdx = lastIdx - 1
						act = seg.FixedName
					}
					var rts []string
					for i := providerIdx + 2; i < lastIdx; i += 2 {
						seg := pathPattern.Segments[i]
						if seg.IsParameter {
							// TODO: shall we resolve some of these violations
							logger.Warn("resource type is parameterized", "path", path, "operation", opKind, "index", i)
							continue
							//return nil, nil, fmt.Errorf("resource type %dth segment is parameterized, in %s (%s)", i, path, opKind)
						}
						rts = append(rts, seg.FixedName)
					}
					rt = "/" + strings.Join(rts, "/")
				}

				absSpecPath, err := filepath.Abs(specpath)
				if err != nil {
					return nil, fmt.Errorf("failed to get abs path for %s: %v", specpath, err)
				}

				opRef := jsonreference.MustCreateRef(absSpecPath + "#" + jsonpointer.Escape(path) + "/" + strings.ToLower(string(opKind)))

				pathPatternStr := PathPatternStr(strings.ToUpper(pathPattern.String()))

				opLoc := OpLocator{
					RP:      strings.ToUpper(rp),
					Version: version,
					RT:      strings.ToUpper(rt),
					ACT:     strings.ToUpper(act),
					Method:  opKind,
				}

				if rpIsGlob {
					opLoc.RP = Wildcard
				}

				if _, ok := infoMap[opLoc]; !ok {
					infoMap[opLoc] = map[PathPatternStr]jsonreference.Ref{}
				}
				if exist, ok := infoMap[opLoc][pathPatternStr]; ok {
					return nil, fmt.Errorf(
						"operation locator %#v for path pattern %s already applied with operation %s, conflicts to the new operation %s", opLoc, pathPatternStr, &exist, &opRef)
				}
				infoMap[opLoc][pathPatternStr] = opRef
			}
		}
	}
	return infoMap, nil
}

func sortedKeys[K ~string, V any](input map[K]V) []K {
	keys := make([]K, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	return keys
}
