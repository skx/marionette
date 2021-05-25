package modules

import (
	"fmt"
	"os"
	"path/filepath"

	mcfg "github.com/skx/marionette/config"
	"github.com/skx/marionette/file"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// GitModule stores our state
type GitModule struct {

	// cfg contains our configuration object.
	cfg *mcfg.Config
}

// Check is part of the module-api, and checks arguments.
func (g *GitModule) Check(args map[string]interface{}) error {

	// Required keys for this module
	required := []string{"repository", "path"}

	// Ensure they exist.
	for _, key := range required {
		_, ok := args[key]
		if !ok {
			return fmt.Errorf("missing '%s' parameter", key)
		}

		val := StringParam(args, key)
		if val == "" {
			return fmt.Errorf("'%s' wasn't a simple string", key)

		}

	}
	return nil
}

// verbose will show the message if the verbose flag is set
func (g *GitModule) verbose(msg string) {
	if g.cfg.Verbose {
		fmt.Printf("%s\n", msg)
	}
}

// Execute is part of the module-api, and is invoked to run a rule.
func (g *GitModule) Execute(args map[string]interface{}) (bool, error) {

	// Repository location - we've already confirmed these are valid
	// in our check function.
	repo := StringParam(args, "repository")
	path := StringParam(args, "path")

	// optional branch to checkout
	branch := StringParam(args, "branch")

	// Have we changed?
	changed := false

	// If we don't have "path/.git" then we need to fetch it
	tmp := filepath.Join(path, ".git")
	if !file.Exists(tmp) {

		g.verbose("\tRepository not present at destination; cloning")

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

	options := &git.PullOptions{RemoteName: "origin"}

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
		g.verbose("\tRepository updated.")
		changed = true
	} else {
		g.verbose("\tNo changes to local repository.\n")
	}

	return changed, err
}

// init is used to dynamically register our module.
func init() {
	Register("git", func(cfg *mcfg.Config) ModuleAPI {
		return &GitModule{cfg: cfg}
	})
}
