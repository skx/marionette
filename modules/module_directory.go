package modules

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// DirectoryModule stores our state
type DirectoryModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *DirectoryModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *DirectoryModule) Execute(args map[string]interface{}) (bool, error) {

	// Default to not having changed
	changed := false

	// Get the target
	target := args["target"]
	str, ok := target.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// Get the mode, if any.  We'll have a default here.
	mode := "0755"
	mode_param, mode_param_present := args["mode"]
	if mode_param_present {
		m, ok := mode_param.(string)
		if ok {
			mode = m
		}
	}

	// User and group changes
	username := ""
	user_val, user_present := args["owner"]
	if user_present {
		_, ok = user_val.(string)
		if ok {
			username = user_val.(string)
		}
	}

	groupname := ""
	group_val, group_present := args["group"]
	if group_present {
		_, ok = group_val.(string)
		if ok {
			groupname = group_val.(string)
		}
	}

	// Convert mode to int
	modeI, _ := strconv.ParseInt(mode, 8, 64)

	// Create the directory, if it is missing.
	if _, err := os.Stat(str); err != nil {
		if os.IsNotExist(err) {
			// make it
			os.MkdirAll(str, os.FileMode(modeI))
			changed = true
		} else {
			// Error running the stat
			return false, err
		}
	}

	// Get the details of the directory, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(str)
	if err != nil {
		return false, err
	}

	// Are we changing owner?
	if username != "" {
		data, err := user.Lookup(username)
		if err != nil {
			return false, err
		}

		// Existing values
		UID := int(info.Sys().(*syscall.Stat_t).Uid)
		GID := int(info.Sys().(*syscall.Stat_t).Gid)

		// proposed owner
		uid, _ := strconv.Atoi(data.Uid)

		if uid != UID {
			os.Chown(str, uid, GID)
			changed = true
		}
	}

	// Are we changing owner?
	if groupname != "" {
		data, err := user.Lookup(groupname)
		if err != nil {
			return false, err
		}

		// Existing values
		UID := int(info.Sys().(*syscall.Stat_t).Uid)
		GID := int(info.Sys().(*syscall.Stat_t).Gid)

		// proposed owner
		gid, _ := strconv.Atoi(data.Gid)

		if gid != GID {
			os.Chown(str, UID, gid)
			changed = true
		}
	}

	// The current mode.
	if mode_param_present && (info.Mode().Perm() != os.FileMode(modeI)) {
		err := os.Chmod(str, os.FileMode(modeI))
		if err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}

// init is used to dynamically register our module.
func init() {
	Register("directory", func() ModuleAPI {
		return &DirectoryModule{}
	})
}
