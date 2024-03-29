// Package file contains some simple utility functions.
package file

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// Copy copies the contents of the source file into the destination file.
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

// Size returns the named files size.
func Size(name string) (int64, error) {
	fi, err := os.Stat(name)

	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
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

// Identical compares the contents of the two specified files, returning
// true if they're identical.
func Identical(a string, b string) (bool, error) {

	hashA, errA := HashFile(a)
	if errA != nil {
		return false, errA
	}

	hashB, errB := HashFile(b)
	if errB != nil {
		return false, errB
	}

	// Are the hashes are identical?
	// If so then the files are identical.
	if hashA == hashB {
		return true, nil
	}

	return false, nil
}
