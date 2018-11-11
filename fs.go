package embedres

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

type ignorePatterns []*regexp.Regexp

func parseIgnores(patterns ...string) (ignorePatterns, error) {
	igs := make(ignorePatterns, 0, len(patterns))
	for _, p := range patterns {
		r, err := regexp.Compile(p)
		if err != nil {
			return nil, fmt.Errorf("invalid regexp: %s, %s", p, err.Error())
		}
		igs = append(igs, r)
	}
	return igs, nil
}

func (i ignorePatterns) Match(s string) bool {
	for _, p := range i {
		if p.MatchString(s) {
			return true
		}
	}
	return false
}
func appendOsPathSeparator(path string) string {
	sep := string(os.PathSeparator)
	if strings.HasSuffix(path, sep) {
		return path
	}
	return path + sep
}
func join(dir, name string) string {
	if dir == "" {
		return name
	}
	if dir == "/" {
		return "/" + name
	}
	return dir + "/" + name
}
func cleanPath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

type File interface {
	Stat() (os.FileInfo, error)
	Readdir(n int) ([]os.FileInfo, error)
	Readdirnames(n int) ([]string, error)
	io.ReadCloser
}

type Fs interface {
	Stat(path string) (os.FileInfo, error)
	Open(path string) (File, error)
}
