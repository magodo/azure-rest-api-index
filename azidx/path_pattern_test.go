package azidx

import (
	"testing"

	"github.com/go-openapi/loads"
	"github.com/stretchr/testify/require"
)

func TestParsePathPatternFromSwagger(t *testing.T) {
	cases := []struct {
		path     string
		expect   []PathPattern
		hasError bool
	}{
		{
			path: "/providers/Microsoft.Resources/operations",
			expect: []PathPattern{
				{
					Segments: []PathSegment{
						{
							FixedName: "providers",
						},
						{
							FixedName: "Microsoft.Resources",
						},
						{
							FixedName: "operations",
						},
					},
				},
			},
		},
		{
			path: "/{scope}/providers/Microsoft.Resources/deployments/{deploymentName}",
			expect: []PathPattern{
				{
					Segments: []PathSegment{
						{
							IsParameter: true,
							IsMulti:     true,
						},
						{
							FixedName: "providers",
						},
						{
							FixedName: "Microsoft.Resources",
						},
						{
							FixedName: "deployments",
						},
						{
							IsParameter: true,
						},
					},
				},
			},
		},
		{
			path: "/{resourceId}",
			expect: []PathPattern{
				{
					Segments: []PathSegment{
						{
							IsParameter: true,
							IsMulti:     true,
						},
					},
				},
			},
		},
		{
			path:     "/{nonexist}",
			hasError: true,
		},
		{
			path: "/providers/Microsoft.CostManagement/{externalCloudProviderType}/{externalCloudProviderId}/alerts",
			expect: []PathPattern{
				{
					Segments: []PathSegment{
						{
							FixedName: "providers",
						},
						{
							FixedName: "Microsoft.CostManagement",
						},
						{
							FixedName: "externalBillingAccounts",
						},
						{
							IsParameter: true,
						},
						{
							FixedName: "alerts",
						},
					},
				},
				{
					Segments: []PathSegment{
						{
							FixedName: "providers",
						},
						{
							FixedName: "Microsoft.CostManagement",
						},
						{
							FixedName: "externalSubscriptions",
						},
						{
							IsParameter: true,
						},
						{
							FixedName: "alerts",
						},
					},
				},
			},
		},
	}

	doc, err := loads.Spec("../testdata/path_pattern/resources.json")
	if err != nil {
		t.Fatal(err)
	}
	swagger := doc.Spec()
	for _, tt := range cases {
		t.Run(tt.path, func(t *testing.T) {
			p, err := ParsePathPatternFromSwagger("../testdata/path_pattern/resources.json", swagger, tt.path, OperationKindGet)
			if tt.hasError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expect, p)
		})
	}
}

func TestParsePathPatternFromString(t *testing.T) {
	cases := []struct {
		input  string
		expect PathPattern
	}{
		{
			input: "/providers/Microsoft.Resources/operations",
			expect: PathPattern{
				Segments: []PathSegment{
					{
						FixedName: "providers",
					},
					{
						FixedName: "Microsoft.Resources",
					},
					{
						FixedName: "operations",
					},
				},
			},
		},
		{
			input: "/{*}/providers/Microsoft.Resources/deployments/{}",
			expect: PathPattern{
				Segments: []PathSegment{
					{
						IsParameter: true,
						IsMulti:     true,
					},
					{
						FixedName: "providers",
					},
					{
						FixedName: "Microsoft.Resources",
					},
					{
						FixedName: "deployments",
					},
					{
						IsParameter: true,
					},
				},
			},
		},
		{
			input: "/{*}",
			expect: PathPattern{
				Segments: []PathSegment{
					{
						IsParameter: true,
						IsMulti:     true,
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expect, *ParsePathPatternFromString(tt.input))
		})
	}
}

func TestParsePathPatternString(t *testing.T) {
	cases := []struct {
		input  PathPattern
		expect string
	}{
		{
			input: PathPattern{
				Segments: []PathSegment{
					{
						FixedName: "providers",
					},
					{
						FixedName: "Microsoft.Resources",
					},
					{
						FixedName: "operations",
					},
				},
			},
			expect: "/providers/Microsoft.Resources/operations",
		},
		{
			input: PathPattern{
				Segments: []PathSegment{
					{
						IsParameter: true,
						IsMulti:     true,
					},
					{
						FixedName: "providers",
					},
					{
						FixedName: "Microsoft.Resources",
					},
					{
						FixedName: "deployments",
					},
					{
						IsParameter: true,
					},
				},
			},
			expect: "/{*}/providers/Microsoft.Resources/deployments/{}",
		},
		{
			input: PathPattern{
				Segments: []PathSegment{
					{
						IsParameter: true,
						IsMulti:     true,
					},
				},
			},
			expect: "/{*}",
		},
	}

	for _, tt := range cases {
		t.Run(tt.expect, func(t *testing.T) {
			require.Equal(t, tt.input.String(), tt.expect)
		})
	}
}
