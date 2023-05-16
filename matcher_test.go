package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatcher_Match(t *testing.T) {
	cases := []struct {
		name    string
		matcher Matcher
		input   string
		expect  bool
	}{
		{
			name: "literal matching string",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			input:  "/foo",
			expect: true,
		},
		{
			name: "literal matching string (empty value)",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "",
					},
				},
			},
			input:  "/",
			expect: true,
		},
		{
			name: "literal matching string (no prefix)",
			matcher: Matcher{
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			input:  "foo",
			expect: true,
		},
		{
			name: "literal non matching string",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			input:  "/bar",
			expect: false,
		},
		{
			name: "matching string with wildcard",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			input:  "/foo/bar",
			expect: true,
		},
		{
			name: "matching string with wildcard in the middle",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
					{
						Value: "baz",
					},
				},
			},
			input:  "/foo/bar/baz",
			expect: true,
		},
		{
			name: "non matching string with wildcard in the middle",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
					{
						Value: "baz",
					},
				},
			},
			input:  "/foo/a/b/baz",
			expect: false,
		},
		{
			name: "matching string with any wildcard in the middle",
			matcher: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
						IsAny:      true,
					},
					{
						Value: "baz",
					},
				},
			},
			input:  "/foo/a/b/baz",
			expect: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, tt.matcher.Match(tt.input))
		})
	}
}

func TestMatchers_Less(t *testing.T) {
	cases := []struct {
		name   string
		m1     Matcher
		m2     Matcher
		isLess bool
	}{
		{
			name: "Equal literally",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			isLess: false,
		},
		{
			name: "Equal with wildcard",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			isLess: false,
		},
		{
			name: "Equal with any",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						IsWildcard: true,
						IsAny:      true,
					},
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						IsWildcard: true,
						IsAny:      true,
					},
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			isLess: false,
		},
		{
			name: "Less literally",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "xfoo",
					},
				},
			},
			isLess: true,
		},
		{
			name: "Less with shorter length (wildcard)",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "xfoo",
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			isLess: true,
		},
		{
			name: "Less with shorter length (any)",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "xfoo",
					},
					{
						IsWildcard: true,
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
					{
						IsWildcard: true,
						IsAny:      true,
					},
				},
			},
			isLess: true,
		},
		{
			name: "Less with per segment comparing",
			m1: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
					},
					{
						IsWildcard: true,
						IsAny:      true,
					},
				},
			},
			m2: Matcher{
				PrefixSep: true,
				Separater: "/",
				Segments: []MatchSegment{
					{
						Value: "foo",
					},
					{
						IsWildcard: true,
						IsAny:      true,
					},
					{
						IsWildcard: true,
					},
				},
			},
			isLess: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isLess, Matchers{tt.m1, tt.m2}.Less(0, 1))
		})
	}
}
