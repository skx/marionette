// Package file contains some simple utility functions.
package file

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// Copy copies the source file into the destination file.
func Copy(src string, dst string) error {
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

// Exists reports whether the named file or directory exists.
func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// HashFile returns the SHA1-hash of the contents of the specified file.
func HashFile(filePath string) (string, error) {
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

// Changed compares the contents of two files, and the return
// code will indicate if they are identical.
func Identical(a string, b string) (bool, error) {

	hashA, errA := HashFile(a)
	if errA != nil {
		return false, errA
	}
	hashB, errB := HashFile(b)
	if errB != nil {
		return false, errB
	}

	// hashes are identical?  No change
	if hashA == hashB {
		return true, nil
	}

	return false, nil
}
