package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SpecListFromReadmeMD(t *testing.T) {
	input := fmt.Sprintf(`
### Tag: package-preview-2023-04

These settings apply only when %[1]s--tag=package-preview-2023-04%[1]s is specified on the command line.

%[1]s%[1]s%[1]syaml $(tag) == 'package-preview-2023-04'
input-file:
  - x.json
%[1]s%[1]s%[1]s
### Tag: package-preview-2023-01

These settings apply only when %[1]s--tag=package-preview-2023-01%[1]s is specified on the command line.

%[1]s%[1]s%[1]syaml $(tag) == 'package-preview-2023-01'
input-file:
  - c.json
  - d.json
%[1]s%[1]s%[1]s


### Tag: package-2023-03

These settings apply only when %[1]s--tag=package-2023-03%[1]s is specified on the command line.

%[1]s%[1]s%[1]syaml $(tag) == 'package-2023-03'
input-file:
  - b.json
  - a.json
%[1]s%[1]s%[1]s
### Tag: package-2021-08

These settings apply only when %[1]s--tag=package-2021-08%[1]s is specified on the command line.

%[1]s%[1]s%[1]s yaml $(tag) == 'package-2021-08'
input-file:
  - e.json
  - c.json
  - foo/$(this-folder)/z.json
%[1]s%[1]s%[1]s
`, "`")
	speclist, err := SpecListFromReadmeMD([]byte(input))
	require.NoError(t, err)
	require.Equal(t, []string{
		"a.json",
		"b.json",
		"c.json",
		"d.json",
		"e.json",
		"foo/z.json",
		"x.json",
	}, speclist)
}
