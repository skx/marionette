package modules

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"syscall"
)

// FileModule stores our state
type FileModule struct {
}

// Check is part of the module-api, and checks arguments.
func (f *FileModule) Check(args map[string]interface{}) error {

	// Ensure we have a target (i.e. name to operate upon).
	_, ok := args["target"]
	if !ok {
		return fmt.Errorf("missing 'target' parameter")
	}

	return nil
}

// Execute is part of the module-api, and is invoked to run a rule.
func (f *FileModule) Execute(args map[string]interface{}) (bool, error) {

	var ret bool
	var err error

	// Get the target
	t := args["target"]
	target, ok := t.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// If we have a source file, copy
	s, ok := args["source"]
	if ok {
		ret, err = f.CopyFile(s.(string), target)
	}

	// If we have a source URL, fetch.
	src_url, ok := args["source_url"]
	if ok {
		ret, err = f.FetchURL(src_url.(string), target)
	}

	//
	// Now we can change the owner/group
	//

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

	// Get the details of the file, so we can see if we need
	// to change owner, group, and mode.
	info, err := os.Stat(target)
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
			os.Chown(target, uid, GID)
			ret = true
		}
	}

	// Are we changing group?
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
			os.Chown(target, UID, gid)
			ret = true
		}
	}

	// The current mode.
	mode_i, _ := strconv.ParseInt(mode, 8, 64)

	if mode_param_present && (info.Mode().Perm() != os.FileMode(mode_i)) {
		err := os.Chmod(target, os.FileMode(mode_i))
		if err != nil {
			return false, err
		}
		ret = true
	}

	return ret, err
}

// FileExists reports whether the named file or directory exists.
func (f *FileModule) FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// HashFile returns the SHA1-hash of the contents of the specified file.
func (f *FileModule) HashFile(filePath string) (string, error) {
	var returnSHA1String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	defer file.Close()

	hash := sha1.New()

	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}

	hashInBytes := hash.Sum(nil)[:20]
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
}

func (f *FileModule) copy(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// We changed
	return out.Close()
}

// Copy the source file to the destination, returning if we changed
// the contents.
func (f *FileModule) CopyFile(src string, dst string) (bool, error) {

	// File doesn't exist - copy it
	if !f.FileExists(dst) {
		err := f.copy(src, dst)
		return true, err
	}

	// OK file does exist.  Compare contents
	a, err_a := f.HashFile(src)
	if err_a != nil {
		return false, err_a
	}
	b, err_b := f.HashFile(dst)
	if err_b != nil {
		return false, err_b
	}

	// hashes are identical?  No change
	if a == b {
		return false, nil
	}

	// otherwise change
	err := f.copy(src, dst)
	return true, err
}

// Fetch the remote URL and save to the given file.
func (f *FileModule) FetchURL(url string, dst string) (bool, error) {

	// Download to temporary file
	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		return false, nil
	}
	defer os.Remove(tmpfile.Name())

	// Get the remote URL
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(tmpfile, resp.Body)

	// File doesn't exist - copy it
	if !f.FileExists(dst) {
		err := f.copy(tmpfile.Name(), dst)
		return true, err
	}

	// OK file does exist.  Compare contents
	a, err_a := f.HashFile(tmpfile.Name())
	if err_a != nil {
		return false, err_a
	}
	b, err_b := f.HashFile(dst)
	if err_b != nil {
		return false, err_b
	}

	// hashes are identical?  No change
	if a == b {
		return false, nil
	}

	// otherwise change
	err = f.copy(tmpfile.Name(), dst)
	return true, err
}

// init is used to dynamically register our file-module
func init() {
	Register("file", func() ModuleAPI {
		return &FileModule{}
	})
}
