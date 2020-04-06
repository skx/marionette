package modules

// PackageModule stores our state
type PackageModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *PackageModule) Check(args map[string]interface{}) error {
	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *PackageModule) Execute(args map[string]interface{}) (bool, error) {
	return false, nil
}

// init is used to dynamically register our module.
func init() {
	Register("package", func() ModuleAPI {
		return &PackageModule{}
	})
}
