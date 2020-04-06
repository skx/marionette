package modules

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

	// Get the target
	t := args["target"]
	target, ok := t.(string)
	if !ok {
		return false, fmt.Errorf("failed to convert target to string")
	}

	// If we have a source file, copy
	s, ok := args["source"]
	if ok {
		ret, err := f.CopyFile(s.(string), target)
		return ret, err
	}

	// If we have a source URL, fetch.
	src_url, ok := args["source_url"]
	if ok {
		ret, err := f.FetchURL(src_url.(string), target)
		return ret, err
	}

	return false, nil
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
