package modules

import (
	"fmt"
	"testing"

	"github.com/skx/marionette/config"
	"github.com/skx/marionette/environment"
)

func TestModules(t *testing.T) {

	// Create our configuration & environment objects
	cfg := &config.Config{Verbose: false}
	env := &environment.Environment{}

	// Get all modules
	modules := Modules()

	for _, module := range modules {

		mod := Lookup(module, cfg, env)
		if mod == nil {
			t.Fatalf("failed to load module")
		}
	}

	count := len(modules)
	if count != 16 {
		t.Fatalf("unexpected number of modules: %d", len(modules))
	}

	// Register an alias
	RegisterAlias("cmd", "shell")

	// Now "cmd" is an alias for "shell"
	if len(Modules()) != count+1 {
		t.Fatalf("unexpected number of modules: %d", len(Modules()))
	}

	a := fmt.Sprintf("%v", Lookup("cmd", cfg, env))
	b := fmt.Sprintf("%v", Lookup("shell", cfg, env))
	if a != b {
		t.Fatalf("alias didn't seem to work?: %s != %s", a, b)
	}

}
