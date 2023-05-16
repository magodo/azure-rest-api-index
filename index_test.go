package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/go-openapi/jsonreference"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex(t *testing.T) {
	specRoot := "./testdata/spec"
	idx, err := BuildIndex(specRoot, "")
	require.NoError(t, err)
	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err)
	pwd, err := os.Getwd()
	require.NoError(t, err)
	expected := fmt.Sprintf(`{
  "rootdir": "./testdata/spec",
  "resource_providers": {
    "MICROSOFT.DUMMY": {
      "2023-05-01-preview": {
        "DELETE": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/delete"
            }
          }
        },
        "GET": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/get"
            }
          }
        },
        "PUT": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/preview/2023-05-01-preview/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/put"
            }
          }
        }
      },
      "2023-05-15": {
        "DELETE": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/delete"
            }
          }
        },
        "GET": {
          "/": {
            "actions": {
              "FOOS": {
                "/PROVIDERS/MICROSOFT.DUMMY/FOOS": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#~1providers~1Microsoft.Dummy~1foos/get"
              }
            }
          },
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/get"
            }
          },
          "/FOOS/BARS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}/BARS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D~1bars~1%%7BbarName%%7D/get"
            }
          }
        },
        "PUT": {
          "/FOOS": {
            "operation_refs": {
              "/PROVIDERS/MICROSOFT.DUMMY/FOOS/{}": "%[1]s/testdata/spec/dummy/resource-manager/Microsoft.Dummy/stable/2023-05-15/foo.json#~1providers~1Microsoft.Dummy~1foos~1%%7BfooName%%7D/put"
            }
          }
        }
      }
    }
  }
}`, pwd)
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
								"/PROVIDERS/RP1/FOOS/{}": jsonreference.MustCreateRef("#*:VER1:GET:/FOOS::P1"),
							},
						},
					},
				},
			},
			"RP1": APIVersions{
				"ver1": APIMethods{
					"GET": ResourceTypes{
						"/": &OperationInfo{
							Actions: map[string]OperationRefs{
								"ACT1": {
									"/PROVIDERS/RP1/ACT1":                  jsonreference.MustCreateRef("#RP1:VER1:GET:/:ACT1:P1"),
									"/SUBSCRIPTIONS/{}/PROVIDERS/RP1/ACT1": jsonreference.MustCreateRef("#RP1:VER1:GET:/:ACT1:P2"),
								},
							},
							OperationRefs: OperationRefs{
								"/PROVIDERS/RP1":                  jsonreference.MustCreateRef("#RP1:VER1:GET:/::P1"),
								"/SUBSCRIPTIONS/{}/PROVIDERS/RP1": jsonreference.MustCreateRef("#RP1:VER1:GET:/::P2"),
								"/{*}/PROVIDERS/RP1":              jsonreference.MustCreateRef("#RP1:VER1:GET:/::P3"),
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
			method: "get",
			expect: "#RP1:VER1:GET:/:ACT1:P1",
		},
		{
			url:    mustParseURL("/subscriptions/sub1/providers/rp1/act1?api-version=ver1"),
			method: "get",
			expect: "#RP1:VER1:GET:/:ACT1:P2",
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
			url:    mustParseURL("/providers/rp1/foos/foo1?api-version=ver1"),
			method: "get",
			expect: "#*:VER1:GET:/FOOS::P1",
		},
	}

	for i, tt := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
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
