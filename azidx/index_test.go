package azidx

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"testing"

	"github.com/go-openapi/jsonreference"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex(t *testing.T) {
	specRoot := "../testdata/spec"
	idx, err := BuildIndex(specRoot, "")
	require.NoError(t, err)
	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err)
	expected := fmt.Sprintf(`{
  "resource_providers": {
    "MICROSOFT.DUMMY": {
      "2023-05-01-preview": {
        "DELETE": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/delete"
            }
          }
        },
        "GET": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/get"
            }
          }
        },
        "PUT": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/put"
            }
          }
        }
      },
      "2023-05-15": {
        "DELETE": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/delete"
            }
          }
        },
        "GET": {
          "/": {
            "actions": {
              "FOOS": {
                "/PROVIDERS/MICROSOFT.DUMMY/FOOS": "dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos/get"
              }
            }
          },
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/get"
            }
          },
          "/FOOS/BARS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}/BARS/{}": "dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D~1bars~1%%7BbarName%%7D/get"
            }
          }
        },
        "PUT": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#/paths/~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/put"
            }
          }
        }
      }
    }
  }
}`)
	require.Equal(t, expected, string(b))
}

func TestIndex_Lookup(t *testing.T) {
	index := Index{
		ResourceProviders: ResourceProviders{
			"*": APIVersions{
				"ver1": APIMethods{
					"GET": ResourceTypes{
						"/FOOS": &OperationInfo{
							OperationRefs: OperationRefs{
								"/PROVIDERS/{}/FOOS/{}": jsonreference.MustCreateRef("#*:VER1:GET:/FOOS::P1"),
							},
						},
					},
				},
			},
			"RP1": APIVersions{
				"ver1": APIMethods{
					"GET": ResourceTypes{
						"/": &OperationInfo{
							OperationRefs: OperationRefs{
								"/PROVIDERS/RP1":                  jsonreference.MustCreateRef("#RP1:VER1:GET:/::P1"),
								"/SUBSCRIPTIONS/{}/PROVIDERS/RP1": jsonreference.MustCreateRef("#RP1:VER1:GET:/::P2"),
								"/{*}/PROVIDERS/RP1":              jsonreference.MustCreateRef("#RP1:VER1:GET:/::P3"),
							},
						},
						"/FOOS": &OperationInfo{
							OperationRefs: OperationRefs{
								"/PROVIDERS/RP1/FOOS/{}":      jsonreference.MustCreateRef("#RP1:VER1:GET:/FOOS::P1"),
								"/PROVIDERS/RP1/FOOS/DEFAULT": jsonreference.MustCreateRef("#RP1:VER1:GET:/FOOS::P2"),
							},
						},
						"/FOOS/BARS": &OperationInfo{
							OperationRefs: OperationRefs{
								"/PROVIDERS/RP1/FOOS/{}/BARS/{}": jsonreference.MustCreateRef("#RP1:VER1:GET:/FOOS/BARS::P1"),
							},
						},
						"/FOOS/*": &OperationInfo{
							OperationRefs: OperationRefs{
								"/PROVIDERS/RP1/FOOS/{}/{}/{}": jsonreference.MustCreateRef("#RP1:VER1:GET:/FOOS/*::P1"),
							},
						},
					},
					"POST": ResourceTypes{
						"/": &OperationInfo{
							Actions: map[string]OperationRefs{
								"ACT1": {
									"/PROVIDERS/RP1/ACT1":                  jsonreference.MustCreateRef("#RP1:VER1:POST:/:ACT1:P1"),
									"/SUBSCRIPTIONS/{}/PROVIDERS/RP1/ACT1": jsonreference.MustCreateRef("#RP1:VER1:POST:/:ACT1:P2"),
								},
							},
						},
						"/FOOS": &OperationInfo{
							Actions: map[string]OperationRefs{
								"*": {
									"/PROVIDERS/RP1/FOOS/{}/{}": jsonreference.MustCreateRef("#RP1:VER1:POST:/FOOS:*:P1"),
								},
							},
						},
					},
				},
			},
		},
	}

	mustParseURL := func(input string) url.URL {
		uRL, err := url.Parse(input)
		if err != nil {
			t.Fatalf("parsing url %s: %v", input, err)
		}
		return *uRL
	}

	cases := []struct {
		url        url.URL
		method     string
		expect     string
		errPattern string
	}{
		{
			url:    mustParseURL("/providers/rp1/act1?api-version=ver1"),
			method: "post",
			expect: "#RP1:VER1:POST:/:ACT1:P1",
		},
		{
			url:    mustParseURL("/subscriptions/sub1/providers/rp1/act1?api-version=ver1"),
			method: "post",
			expect: "#RP1:VER1:POST:/:ACT1:P2",
		},
		{
			url:        mustParseURL("/subscriptions/sub1/resourceGroups/rg1/providers/rp1/act1?api-version=ver1"),
			method:     "get",
			errPattern: "matches nothing",
		},
		{
			url:    mustParseURL("/subscriptions/sub1/resourceGroups/rg1/providers/rp1?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/::P3",
		},
		{
			url:    mustParseURL("/providers/rp0/foos/foo1?api-version=ver1"),
			method: "get",
			expect: "#*:VER1:GET:/FOOS::P1",
		},
		{
			url:    mustParseURL("/providers/rp0/foos/foo1?api-version=ver1"),
			method: "get",
			expect: "#*:VER1:GET:/FOOS::P1",
		},
		{
			url:    mustParseURL("/providers/rp1/foos/foo1?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/FOOS::P1",
		},
		{
			url:    mustParseURL("/providers/rp1/foos/default?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/FOOS::P2",
		},
		{
			url:    mustParseURL("/providers/rp1/foos/foo1/sleep?api-version=ver1"),
			method: "post",
			expect: "#RP1:VER1:POST:/FOOS:*:P1",
		},
		{
			url:    mustParseURL("/providers/rp1/foos/foo1/bars/bar1?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/FOOS/BARS::P1",
		},
		{
			url:    mustParseURL("/providers/rp1/foos/foo1/bazs/baz1?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/FOOS/*::P1",
		},
	}

	for _, tt := range cases {
		t.Run(tt.method+" "+tt.url.String(), func(t *testing.T) {
			ref, err := index.Lookup(tt.method, tt.url)
			if tt.errPattern != "" {
				require.Error(t, err)
				require.Regexp(t, regexp.MustCompile(tt.errPattern), err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expect, ref.String())
		})
	}
}
