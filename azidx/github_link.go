package azidx

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"text/scanner"

	"github.com/go-git/go-git/v5"
	"github.com/go-openapi/jsonpointer"
)

func BuildGithubLink(ptr jsonpointer.Pointer, commit, specdir, fpath string) (string, error) {
	repo, err := git.PlainOpen(filepath.Dir(specdir))
	if err != nil {
		return "", err
	}
	ref, err := repo.Head()
	if err != nil {
		return "", err
	}
	if repoCommit := ref.Hash().String(); repoCommit != commit {
		return "", fmt.Errorf("repository commit %q not equals to the commit the index is built %q", repoCommit, commit)
	}

	b, err := os.ReadFile(fpath)
	if err != nil {
		return "", err
	}
	offset, err := JSONPointerOffset(ptr, string(b))
	if err != nil {
		return "", err
	}
	var sc scanner.Scanner
	sc.Init(bytes.NewBuffer(b))
	fmt.Println(offset)
	for i := 0; i < int(offset); i++ {
		sc.Next()
	}
	pos := sc.Pos()

	specdir, err = filepath.Abs(specdir)
	if err != nil {
		return "", err
	}

	relFile, err := filepath.Rel(specdir, fpath)
	if err != nil {
		return "", err
	}

	return "https://github.com/Azure/azure-rest-api-specs/blob/" + commit + "/specification/" + relFile + "#L" + strconv.Itoa(pos.Line), nil
}
