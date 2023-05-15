package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

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
