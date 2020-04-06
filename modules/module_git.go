package modules

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
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

	// If we don't have "path/.git" then we're fetching
	tmp := filepath.Join(path, ".git")
	if g.FileExists(tmp) {
		return false, nil
	}

	// Clone since it is missing.
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	})

	return true, err
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
