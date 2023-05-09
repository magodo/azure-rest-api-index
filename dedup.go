package main

import (
	"regexp"

	"github.com/go-openapi/jsonreference"
)

type DedupMatcher struct {
	RP             *regexp.Regexp
	Version        *regexp.Regexp
	RT             *regexp.Regexp
	ACT            *regexp.Regexp
	Method         *regexp.Regexp
	PathPatternStr *regexp.Regexp
}

func (key DedupMatcher) Match(loc OpLocator, pathp string) bool {
	return (key.RP == nil || key.RP.MatchString(loc.RP)) &&
		(key.Version == nil || key.Version.MatchString(loc.Version)) &&
		(key.RT == nil || key.RT.MatchString(loc.RT)) &&
		(key.ACT == nil || key.ACT.MatchString(loc.ACT)) &&
		(key.Method == nil || key.Method.MatchString(string(loc.Method))) &&
		(key.PathPatternStr == nil || key.PathPatternStr.MatchString(pathp))
}

type DedupPicker struct {
	SpecPath *regexp.Regexp
	Pointer  *regexp.Regexp
}

func (picker DedupPicker) Match(ref jsonreference.Ref) bool {
	specPath := ref.GetURL().Path
	pointer := ref.GetPointer().String()
	return (picker.SpecPath == nil || picker.SpecPath.MatchString(specPath)) &&
		(picker.Pointer == nil || picker.Pointer.MatchString(pointer))
}

type Deduplicator map[DedupMatcher]DedupPicker

type DeduplicateRecords []DedupRecord

type DedupRecord struct {
	Matcher DedupMatcherIn `json:"matcher"`
	Picker  DedupPickerIn  `json:"picker"`
}

type DedupMatcherIn struct {
	RP             string `json:"rp,omitempty"`
	Version        string `json:"versio,omitemptyn"`
	RT             string `json:"rt,omitempty"`
	ACT            string `json:"act,omitempty"`
	Method         string `json:"method,omitempty"`
	PathPatternStr string `json:"path_pattern_str,omitempty"`
}

type DedupPickerIn struct {
	SpecPath string `json:"spec_path,omitempty"`
	Pointer  string `json:"pointer,omitempty"`
}

func (records DeduplicateRecords) ToDeduplicator() Deduplicator {
	dup := Deduplicator{}
	for _, rec := range records {
		matcher, picker := rec.Matcher, rec.Picker
		m := DedupMatcher{}
		if matcher.RP != "" {
			m.RP = regexp.MustCompile(matcher.RP)
		}
		if matcher.Version != "" {
			m.Version = regexp.MustCompile(matcher.Version)
		}
		if matcher.RT != "" {
			m.RT = regexp.MustCompile(matcher.RT)
		}
		if matcher.ACT != "" {
			m.ACT = regexp.MustCompile(matcher.ACT)
		}
		if matcher.Method != "" {
			m.Method = regexp.MustCompile(matcher.Method)
		}
		if matcher.PathPatternStr != "" {
			m.Method = regexp.MustCompile(matcher.PathPatternStr)
		}
		p := DedupPicker{}
		if picker.SpecPath != "" {
			p.SpecPath = regexp.MustCompile(picker.SpecPath)
		}
		if picker.Pointer != "" {
			p.Pointer = regexp.MustCompile(picker.Pointer)
		}
		dup[m] = p
	}
	return dup
}
