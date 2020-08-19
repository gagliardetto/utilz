package utilz

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ryanuber/go-glob"
)

// CreateFolderIfNotExists creates a folder if it does not exists
func CreateFolderIfNotExists(name string, perm os.FileMode) error {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return os.MkdirAll(name, perm)
	}
	return err
}
func MustCreateFolderIfNotExists(path string, perm os.FileMode) {
	err := CreateFolderIfNotExists(path, perm)
	if err != nil {
		panic(Sf("error creating dir %q: %s", path, err))
	}
}

func DirExists(path string) (bool, error) {
	return FileExists(path)
}

func FileExists(filepath string) (bool, error) {
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}
func FileExistsWithStat(filepath string) (os.FileInfo, bool, error) {
	stat, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if err == nil {
		return stat, true, nil
	}
	return nil, false, err
}

// ReadFileLinesAsString iterates on the lines of a text file
func ReadFileLinesAsString(filepath string, iterator func(line string) bool) error {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		doContinue := iterator(scanner.Text())
		if !doContinue {
			return nil
		}
		err = scanner.Err()
		if err != nil {
			return fmt.Errorf("error while iterating over scanner: %s", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error of scanner: %s", err)
	}

	return nil
}

// ReadConfigLinesAsString iterates on the lines of a text file,
// trimming spaces, ignoring empty lines and comments.
func ReadConfigLinesAsString(filepath string, iterator func(line string) bool) error {
	return ReadFileLinesAsString(filepath, func(line string) bool {
		// trim space:
		line = strings.TrimSpace(line)
		// trim newlines:
		line = strings.Trim(line, "\n")

		// ignore empty lines:
		if len(line) == 0 {
			return true
		}

		// ignore comment lines:
		isComment := strings.HasPrefix(line, "#")
		if isComment {
			return true
		}
		return iterator(line)
	})
}

// ReadStringLineByLine iterates over a multiline string
func ReadStringLineByLine(s string, iterator func(line string) bool) error {

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		doContinue := iterator(scanner.Text())
		if !doContinue {
			return nil
		}
		err := scanner.Err()
		if err != nil {
			return fmt.Errorf("error while iterating over scanner: %s", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error of scanner: %s", err)
	}

	return nil
}
func MustAbs(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		panic(Sf("error absolute path of %q: %s", path, err))
	}
	return abs
}

func IsFolder(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func MustFileExists(path string) bool {
	exists, err := FileExists(path)
	if err != nil {
		panic(Sf("%q: %s", path, err))
	}
	return exists
}

func MustCopyFile(src, dst string) {
	_, err := copyFile(src, dst)
	if err != nil {
		panic(Sf("error copying %q to %q: %s", src, dst, err))
	}
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func PanicIfAlreadyExists(path string) {
	if MustFileExists(path) {
		panic(Sf("file %q already exists", path))
	}
}

func MustRemoveFile(path string) {
	if MustFileExists(path) {
		err := os.Remove(path)
		if err != nil {
			panic(Sf("error while removing file %q: %s", path, err))
		}
	}
}

// TrimExt trims the file extension of the file(path).
func TrimExt(p string) string {
	return strings.TrimSuffix(p, path.Ext(p))
}

// AddNumericSuffixIfFileExists: if the specified file exists,
// a numeric suffix is added to the name of the file, right before the
// extension; numeric suffixes are tried incrementally until a non-existing file
// is idedntified; the complete filepath is returned (no file is created at this stage).
func AddNumericSuffixIfFileExists(filePath string) string {

	index := 0
	for {
		if index == 0 {
			if MustFileExists(filePath) {
				index++
			} else {
				return filePath
			}
		} else {
			newPath := TrimExt(filePath) + "." + Itoa(index) + path.Ext(filePath)
			if MustFileExists(newPath) {
				index++
			} else {
				return newPath
			}
		}
	}
}

// Blacklist filters out elements from the `items` slice that are also
// present in the `blacklist` slice; a simple == comparison is done.
func Blacklist(items []string, blacklist []string) []string {
	var filtered []string
	for i := range items {
		item := items[i]
		if !SliceContains(blacklist, item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// HasMatch finds the matching pattern (glob) to which the provided item matches.
func HasMatch(item string, patterns []string) (string, bool) {
	if item == "" {
		return "", false
	}

	// sort the patterns in increasing length order:
	sort.Strings(patterns)

	// first, try to find a precise match:
	for _, pattern := range patterns {
		if pattern == item {
			return pattern, true
		}
	}
	// ... then look for a glob match:
	for _, pattern := range patterns {
		if isMatch := glob.Glob(pattern, item); isMatch {
			return pattern, true
		}
	}
	return "", false
}

// BlacklistGlob filters out elements from the `items` slice that match
// any pattern from the `blacklist` slice; a glob comparison is done.
func BlacklistGlob(items []string, blacklist []string) []string {
	var filtered []string
	for i := range items {
		item := items[i]
		if match, isBlacklisted := HasMatch(item, blacklist); !isBlacklisted && match == "" {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// WhitelistGlob filters out elements from the `items` slice that DON'T match
// any pattern from the `whitelist` slice; a glob comparison is done.
func WhitelistGlob(items []string, whitelist []string) []string {
	var filtered []string
	for i := range items {
		item := items[i]
		if match, isWhitelisted := HasMatch(item, whitelist); isWhitelisted && match != "" {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

var ErrNotDir = errors.New("path is not a directory")

// ListChildDirs lists folders (full path) inside the specified folder (just first level; no grandchildren);
// if the provided filepath is not a folder, and error is returned.
func ListChildDirs(path string) ([]string, error) {
	isDir, err := IsFolder(path)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, ErrNotDir
	}
	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	folders := make([]string, 0)
	for _, content := range contents {
		if content.IsDir() {
			folders = append(folders, filepath.Join(path, content.Name()))
		}
	}
	return folders, nil
}

// ListChildDirsWithBlacklist lists the child folders (full path) of a parent folder, filtering out
// folders that match one of the provided blacklist glob patterns.
func ListChildDirsWithBlacklist(path string, blacklistPatterns []string) ([]string, error) {
	folders, err := ListChildDirs(path)
	if err != nil {
		return nil, err
	}

	filtered := BlacklistGlob(folders, blacklistPatterns)
	return filtered, nil
}

// ListChildDirsWithWhitelist lists the child folders (full path) of a parent folder,
// filtering out folders that DON'T match any of the provided whitelist glob patterns.
func ListChildDirsWithWhitelist(path string, whitelistPatterns []string) ([]string, error) {
	folders, err := ListChildDirs(path)
	if err != nil {
		return nil, err
	}

	filtered := WhitelistGlob(folders, whitelistPatterns)
	return filtered, nil
}

// ListFiles lists files (full path) inside the specified folder (just first level; no grandchildren);
// if the provided filepath is not a folder, and error is returned.
func ListFiles(path string) ([]string, error) {
	isDir, err := IsFolder(path)
	if err != nil {
		return nil, err
	}
	if !isDir {
		return nil, ErrNotDir
	}
	contents, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0)
	for _, content := range contents {
		// TODO: add a better check:
		isFile := !content.IsDir()
		if isFile {
			files = append(files, filepath.Join(path, content.Name()))
		}
	}
	return files, nil
}

func CountLines(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
