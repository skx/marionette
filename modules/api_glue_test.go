package modules

import (
	"fmt"
	"testing"

	"github.com/skx/marionette/config"
)

func TestModules(t *testing.T) {

	// Create our configuration object
	cfg := &config.Config{Verbose: false}

	// Get all modules
	modules := Modules()

	for _, module := range modules {

		mod := Lookup(module, cfg)
		if mod == nil {
			t.Fatalf("failed to load module")
		}
	}

	count := len(modules)
	if count != 12 {
		t.Fatalf("unexpected number of modules: %d", len(modules))
	}

	// Register an alias
	RegisterAlias("cmd", "shell")

	// Now "cmd" is an alias for "shell"
	if len(Modules()) != count+1 {
		t.Fatalf("unexpected number of modules: %d", len(Modules()))
	}

	a := fmt.Sprintf("%v", Lookup("cmd", cfg))
	b := fmt.Sprintf("%v", Lookup("shell", cfg))
	if a != b {
		t.Fatalf("alias didn't seem to work?: %s != %s", a, b)
	}

}
