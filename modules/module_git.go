package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// GitModule stores our state
type GitModule struct {
}

// Check is part of the module-api, and checks arguments.
func (g *GitModule) Check(args map[string]interface{}) error {

	required := []string{"repository", "path"}

	for _, key := range required {

		_, ok := args[key]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", key)
		}
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (g *GitModule) Execute(args map[string]interface{}) (bool, error) {

	// Repository location
	c := args["repository"]
	repo, check := c.(string)
	if !check {
		return false, fmt.Errorf("failed to convert 'repository' to string")
	}

	// Repository location
	path := ""
	c = args["path"]
	path, check = c.(string)
	if !check {
		return false, fmt.Errorf("failed to convert 'path' to string")
	}

	// optional branch to checkout
	branch := ""
	c = args["branch"]
	branch, check = c.(string)

	// Have we changed?
	changed := false

	// If we don't have "path/.git" then we need to fetch it
	tmp := filepath.Join(path, ".git")
	if !g.FileExists(tmp) {

		// Clone since it is missing.
		_, err := git.PlainClone(path, false, &git.CloneOptions{
			URL:      repo,
			Progress: os.Stdout,
		})

		if err != nil {
			return false, err
		}

		changed = true
	}

	//
	// OK now we need to pull in any changes.
	//

	// Open the repo.
	r, err := git.PlainOpen(path)
	if err != nil {
		return false, err
	}

	// Get the head-commit
	ref, err := r.Head()
	if err != nil {
		return false, err
	}

	// Get the work tree
	w, err := r.Worktree()
	if err != nil {
		return false, err
	}

	// If we're to switch branch do that
	if branch != "" {

		// fetch references
		err = r.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{"refs/*:refs/*", "HEAD:refs/heads/HEAD"},
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return false, err
		}

		// checkout the branch
		err = w.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
			Force:  true,
		})
		if err != nil {
			return false, err
		}
	}

	// Update the work-tree.  Note that we have to set the
	// reference to the branch if we're using one.
	options := &git.PullOptions{RemoteName: "origin"}
	if branch != "" {
		options.ReferenceName = plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))
	}

	// Do the pull
	err = w.Pull(options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return false, err
	}

	// Get the second ref
	ref2, err := r.Head()
	if err != nil {
		return false, err
	}

	// If the hashes differ we've updated, and thus changed
	if ref2.Hash() != ref.Hash() {
		changed = true
	}

	return changed, err
}

// FileExists reports whether the named file or directory exists.
func (g *GitModule) FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// init is used to dynamically register our module.
func init() {
	Register("git", func() ModuleAPI {
		return &GitModule{}
	})
}
