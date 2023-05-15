package specpath

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/magodo/azure-rest-api-index/tests"
	"github.com/stretchr/testify/require"
)

func TestSPecPathInfo(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		pathInfo *Info
		hasError bool
	}{
		{
			name: "regular stable spec file",
			path: "/compute/resource-manager/Microsoft.Compute/stable/2020-01-01/compute.json",
			pathInfo: &Info{
				ResourceProvider:   "compute",
				ResourceProviderMS: "Microsoft.Compute",
				IsPreview:          false,
				Version:            "2020-01-01",
				SpecName:           "compute.json",
			},
		},
		{
			name: "regular preview spec file",
			path: "/compute/resource-manager/Microsoft.Compute/preview/2020-01-01-preview/compute.json",
			pathInfo: &Info{
				ResourceProvider:   "compute",
				ResourceProviderMS: "Microsoft.Compute",
				IsPreview:          true,
				Version:            "2020-01-01-preview",
				SpecName:           "compute.json",
			},
		},
		{
			name: "regular stable spec file with sub service",
			path: "/mediaservices/resource-manager/Microsoft.Media/Accounts/preview/2019-05-01-preview/Accounts.json",
			pathInfo: &Info{
				ResourceProvider:   "mediaservices",
				ResourceProviderMS: "Microsoft.Media",
				IsPreview:          true,
				Version:            "2019-05-01-preview",
				SpecName:           "Accounts.json",
				subservice:         ptr("Accounts"),
			},
		},
		{
			name:     "regular stable spec file, wrong rootdir",
			path:     "/some/root/dir/compute/resource-manager/Microsoft.Compute/stable/2020-01-01/compute.json",
			hasError: true,
		},
		{
			name:     "spec not ends with .json",
			path:     "/compute/resource-manager/Microsoft.Compute/preview/2020-01-01-preview/compute",
			hasError: true,
		},
		{
			name:     "path has wrong segment count",
			path:     "/compute/resource-manager/Microsoft.Compute/2020-01-01-preview/compute",
			hasError: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			info, err := SpecPathInfo("/", tt.path)
			if tt.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, *tt.pathInfo, *info)
		})
	}
}

func TestInfo_ToPath(t *testing.T) {
	cases := []struct {
		input  Info
		expect string
	}{
		{
			input: Info{
				ResourceProvider:   "compute",
				ResourceProviderMS: "Microsoft.Compute",
				IsPreview:          false,
				Version:            "2020-01-01",
				SpecName:           "compute.json",
			},
			expect: "compute/resource-manager/Microsoft.Compute/stable/2020-01-01/compute.json",
		},
		{
			input: Info{
				ResourceProvider:   "compute",
				ResourceProviderMS: "Microsoft.Compute",
				IsPreview:          true,
				Version:            "2020-01-01-preview",
				SpecName:           "compute.json",
			},
			expect: "compute/resource-manager/Microsoft.Compute/preview/2020-01-01-preview/compute.json",
		},
		{
			input: Info{
				ResourceProvider:   "mediaservices",
				ResourceProviderMS: "Microsoft.Media",
				IsPreview:          true,
				Version:            "2019-05-01-preview",
				SpecName:           "Accounts.json",
				subservice:         ptr("Accounts"),
			},
			expect: "mediaservices/resource-manager/Microsoft.Media/Accounts/preview/2019-05-01-preview/Accounts.json",
		},
	}

	for idx, tt := range cases {
		t.Run(strconv.Itoa(idx), func(t *testing.T) {
			require.Equal(t, tt.expect, tt.input.ToPath())
		})
	}
}

func Test_E2E_SpecPathInfo(t *testing.T) {
	tests.E2EPrecheck(t)
	repoDir := os.Getenv("AZURE_REST_API_INDEX_E2E_SPEC_REPO")
	if repoDir == "" {
		t.Skip(`"AZURE_REST_API_INDEX_E2E_SPEC_REPO" not specified`)
	}
	specRootDir := filepath.Join(repoDir, "specification")
	if err := filepath.Walk(specRootDir,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info == nil {
				return nil
			}
			if info.IsDir() {
				if strings.EqualFold(filepath.Base(p), "data-plane") {
					return filepath.SkipDir
				}
				if strings.EqualFold(filepath.Base(p), "examples") {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip files not match the schema file patterns
			if _, err := SpecPathInfo(specRootDir, p); err != nil {
				if filepath.Ext(p) == ".json" {
					t.Logf("%s: %v", p, err)
				}
				return nil
			}
			return nil
		}); err != nil {
		t.Fatal(err)
	}
}

func ptr[T any](v T) *T {
	return &v
}
