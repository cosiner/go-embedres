package embedres

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type localFile struct {
	path string
	*os.File
	fs *LocalFs
}

func (l *localFile) Readdirnames(n int) ([]string, error) {
	names, err := l.File.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	var nnames []string
	for _, name := range names {
		if n > 0 && n >= len(nnames) {
			break
		}
		path := join(l.path, name)
		_, err := l.fs.realPath(path)
		if err != nil {
			continue
		}
		nnames = append(nnames, name)
	}
	return nnames, nil
}

func (l *localFile) Readdir(n int) ([]os.FileInfo, error) {
	items, err := l.File.Readdir(-1)
	if err != nil {
		return nil, err
	}
	var nitems []os.FileInfo
	for _, item := range items {
		if n > 0 && n >= len(items) {
			break
		}
		path := join(l.path, item.Name())
		_, err := l.fs.realPath(path)
		if err != nil {
			continue
		}
		nitems = append(nitems, item)
	}
	return nitems, nil
}

type LocalFs struct {
	ignores ignorePatterns
	paths   []string
	prefix  string
}

func NewLocalFs(prefix string, paths, ignores []string) (*LocalFs, error) {
	if prefix != "" {
		absprefix, err := filepath.Abs(filepath.Clean(prefix))
		if err != nil {
			return nil, fmt.Errorf("get absolute path of prefix failed: %s, %s", prefix, err.Error())
		}
		prefix = absprefix
	}

	for i, path := range paths {
		abspath, err := filepath.Abs(filepath.Clean(path))
		if err != nil {
			return nil, fmt.Errorf("get absolute path failed: %s, %s", path, err.Error())
		}
		if prefix != "" && prefix != abspath && !strings.HasPrefix(abspath, appendOsPathSeparator(prefix)) {
			return nil, fmt.Errorf("invalid path prefix: %s, %s", path, prefix)
		}
		paths[i] = abspath
	}

	ignorePatterns, err := parseIgnores(ignores...)
	if err != nil {
		return nil, fmt.Errorf("invalid ignore pattern: %s", err.Error())
	}

	return &LocalFs{
		ignores: ignorePatterns,
		paths:   paths,
		prefix:  prefix,
	}, nil
}

func (l *LocalFs) realPath(path string) (string, error) {
	path = filepath.FromSlash(path)
	path = filepath.Join(l.prefix, path)
	var match bool
	for _, p := range l.paths {
		match = path == p ||
			strings.HasPrefix(path, appendOsPathSeparator(p)) ||
			strings.HasPrefix(p, appendOsPathSeparator(path))
		if match {
			break
		}
	}
	if !match {
		return "", os.ErrNotExist
	}
	if l.ignores.Match(path) {
		return "", os.ErrNotExist
	}
	return path, nil
}

func (l *LocalFs) Stat(path string) (os.FileInfo, error) {
	path = cleanPath(path)
	path, err := l.realPath(path)
	if err != nil {
		return nil, err
	}
	return os.Lstat(path)
}

func (l *LocalFs) Open(path string) (File, error) {
	path = cleanPath(path)
	realpath, err := l.realPath(path)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(realpath, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	return &localFile{
		path: path,
		File: fd,
		fs:   l,
	}, nil
}
