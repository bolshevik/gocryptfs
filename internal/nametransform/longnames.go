package nametransform

import (
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rfjakob/gocryptfs/internal/toggledlog"
)

// Files with long names are stored in two files:
// gocryptfs.longname.[sha256]       <--- File content
// gocryptfs.longname.[sha256].name  <--- File name
const longNamePrefix = "gocryptfs.longname."
const longNameSuffix = ".name"

// HashLongName - take the hash of a long string "name" and return
// "gocryptfs.longname.[sha256]"
func HashLongName(name string) string {
	hashBin := sha256.Sum256([]byte(name))
	hashBase64 := base64.URLEncoding.EncodeToString(hashBin[:])
	return longNamePrefix + hashBase64
}

// IsLongName - detect if cName is
// gocryptfs.longname.* ........ 1
// gocryptfs.longname.*.name ... 2
// else ........................ 0
func IsLongName(cName string) int {
	if !strings.HasPrefix(cName, longNamePrefix) {
		return 0
	}
	if strings.HasSuffix(cName, longNameSuffix) {
		return 2
	}
	return 1
}

// ReadLongName - read path.name
func ReadLongName(path string) (string, error) {
	content, err := ioutil.ReadFile(path + longNameSuffix)
	if err != nil {
		toggledlog.Warn.Printf("ReadLongName: %v", err)
	}
	return string(content), err
}

// DeleteLongName - if cPath ends in "gocryptfs.longname.[sha256]",
// delete the "gocryptfs.longname.[sha256].name" file
func DeleteLongName(cPath string) error {
	if IsLongName(filepath.Base(cPath)) == 1 {
		err := syscall.Unlink(cPath + longNameSuffix)
		if err != nil {
			toggledlog.Warn.Printf("DeleteLongName: %v", err)
		}
		return err
	}
	return nil
}

// WriteLongName - if cPath ends in "gocryptfs.longname.[sha256]", write the
// "gocryptfs.longname.[sha256].name" file
func (n *NameTransform) WriteLongName(cPath string, plainPath string) (err error) {
	cHashedName := filepath.Base(cPath)
	if IsLongName(cHashedName) != 1 {
		// This is not a hashed file name, nothing to do
		return nil
	}
	// Encrypt (but do not hash) the plaintext name
	cDir := filepath.Dir(cPath)
	dirIV, err := ReadDirIV(cDir)
	if err != nil {
		toggledlog.Warn.Printf("WriteLongName: %v", err)
		return err
	}
	plainName := filepath.Base(plainPath)
	cName := n.EncryptName(plainName, dirIV)
	// Write the encrypted name into gocryptfs.longname.[sha256].name
	err = ioutil.WriteFile(filepath.Join(cDir, cHashedName+longNameSuffix), []byte(cName), 0600)
	if err != nil {
		toggledlog.Warn.Printf("WriteLongName: %v", err)
	}
	return err
}