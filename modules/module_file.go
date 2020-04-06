package modules

// FileModule stores our state
type FileModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *FileModule) Check(args map[string]interface{}) error {
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FileModule) Execute(args map[string]interface{}) (bool, error) {
	return false, nil
}

// init is used to dynamically register our file-module
func init() {
	Register("file", func() ModuleAPI {
		return &FileModule{}
	})
}
