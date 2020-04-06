package modules

// ShellModule stores our state
type ShellModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *ShellModule) Check(args map[string]interface{}) error {
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *ShellModule) Execute(args map[string]interface{}) (bool, error) {
	return false, nil
}

// init is used to dynamically register our module.
func init() {
	Register("shell", func() ModuleAPI {
		return &ShellModule{}
	})
}
