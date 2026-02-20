package azidx

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/magodo/jsonpointerpos"
)

func BuildGithubLink(fpath string, fpos jsonpointerpos.JSONPointerPosition, commit, specdir string) (string, error) {
	repo, err := git.PlainOpen(filepath.Dir(specdir))
	if err != nil {
		if err != git.ErrRepositoryNotExists {
			return "", err
		}
	} else {
		head, err := repo.Head()
		if err != nil {
			return "", err
		}
		if repoCommit := head.Hash().String(); repoCommit != commit {
			return "", fmt.Errorf("repository commit %q not equals to the commit the index is built %q", repoCommit, commit)
		}
	}

	relFile, err := filepath.Rel(specdir, fpath)
	if err != nil {
		return "", err
	}

	return "https://github.com/Azure/azure-rest-api-specs/blob/" + commit + "/specification/" + relFile + "#L" + strconv.Itoa(fpos.Line), nil
}
