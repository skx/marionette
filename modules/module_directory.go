package modules

// DirectoryModule stores our state
type DirectoryModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *DirectoryModule) Check(args map[string]interface{}) error {
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DirectoryModule) Execute(args map[string]interface{}) (bool, error) {
	return false, nil
}

// init is used to dynamically register our module.
func init() {
	Register("directory", func() ModuleAPI {
		return &DirectoryModule{}
	})
}
