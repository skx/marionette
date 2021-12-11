//go:build !windows
// +build !windows

package file

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// ChangeMode changes the mode of the given file/directory to the
// specified value.
//
// If the mode was changed, this function will return true.
func ChangeMode(path string, mode string) (bool, error) {

	// Get the mode as an integer.
	//
	// NOTE: We expect octal input.
	m, _ := strconv.ParseInt(mode, 8, 64)

	// Get the details of the file, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// If the mode doesn't match what we expect then change it
	if info.Mode().Perm() != os.FileMode(m) {
		err = os.Chmod(path, os.FileMode(m))
		if err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// ChangeOwner changes the owner of the given file/directory to
// the specified value.
//
// If the ownership was changed this function will return true.
//
func ChangeOwner(path string, owner string) (bool, error) {

	// Get the details of the file, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// Get the user-details of who we should change to.
	var data *user.User
	data, err = user.Lookup(owner)
	if err != nil {
		return false, err
	}

	// Existing values
	UID := int(info.Sys().(*syscall.Stat_t).Uid)
	GID := int(info.Sys().(*syscall.Stat_t).Gid)

	// proposed owner
	uid, _ := strconv.Atoi(data.Uid)

	if uid != UID {
		err = os.Chown(path, uid, GID)
		return true, err
	}

	return false, nil
}

// ChangeGroup changes the group of the given file/directory to
// the specified value.
//
// If the ownership was changed this function will return true.
//
func ChangeGroup(path string, group string) (bool, error) {

	// Get the details of the file, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// Get the user-details of who we should change to.
	var data *user.User
	data, err = user.Lookup(group)
	if err != nil {
		return false, err
	}

	// Existing values
	UID := int(info.Sys().(*syscall.Stat_t).Uid)
	GID := int(info.Sys().(*syscall.Stat_t).Gid)

	// proposed owner
	gid, _ := strconv.Atoi(data.Gid)

	if gid != GID {
		err = os.Chown(path, UID, gid)

		return true, err
	}

	return false, nil
}
