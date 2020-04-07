// +build windows

package file

// ChangeOwner is a NOP on Microsoft Windows.
func ChangeOwner(path string, group string) (bool, error) {
	return false, nil
}

// ChangeGroup is a NOP on Microsoft Windows.
func ChangeGroup(path string, group string) (bool, error) {
	return false, nil
}
