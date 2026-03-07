package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// GitDiff checks if the given file is currently modified in the git worktree.
// If the file is currently modified, a diff containing the changes between the
// committed and current file contents is returned.
func GitDiff(packagePath, fileName string) (string, error) {
	repo, err := git.PlainOpenWithOptions(packagePath, &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("error opening parent git repository: %w", err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("error opening repository worktree: %w", err)
	}

	status, err := wt.Status()
	if err != nil {
		return "", fmt.Errorf("error getting worktree status: %w", err)
	}

	pathedFileName := filepath.Join(packagePath, fileName)
	if status.File(pathedFileName).Worktree == git.Unmodified {
		// The file is unmodified since the last commit, no diff to return.
		return "", nil
	}

	committedContent, err := getCommittedFileContents(repo, pathedFileName)
	if err != nil {
		return "", fmt.Errorf("could not get file content of %q from HEAD: %w", pathedFileName, err)
	}

	currentContent, err := os.ReadFile(pathedFileName)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("could not read file %q: %w", pathedFileName, err)
	}

	return Diff(committedContent, currentContent, fileName), nil
}

// getCommittedFileContents retrieves the state of the file from the git repository
// as it is committed in the current HEAD commit.
func getCommittedFileContents(repo *git.Repository, fileName string) ([]byte, error) {
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("could not get repository HEAD: %w", err)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("could not get HEAD commit: %w", err)
	}

	file, err := commit.File(fileName)
	if errors.Is(err, object.ErrFileNotFound) {
		// The file didn't exist, so the diff is the entire file.
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("could not get file %q from commit: %w", fileName, err)
	}

	fileContent, err := file.Contents()
	if err != nil {
		return nil, fmt.Errorf("could not get file content: %w", err)
	}

	return []byte(fileContent), nil
}
