//go:build windows
// +build windows

package file

// ChangeMode is a NOP on Microsoft Windows
func ChangeMode(path string, mode string) (bool, error) {
	return false, nil
}

// ChangeOwner is a NOP on Microsoft Windows.
func ChangeOwner(path string, group string) (bool, error) {
	return false, nil
}

// ChangeGroup is a NOP on Microsoft Windows.
func ChangeGroup(path string, group string) (bool, error) {
	return false, nil
}
