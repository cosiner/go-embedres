package embedres

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type embedFileInfo struct {
	dir     string
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool

	content func() (io.ReadCloser, error)
}

func (e *embedFileInfo) Name() string       { return e.name }
func (e *embedFileInfo) Size() int64        { return e.size }
func (e *embedFileInfo) Mode() os.FileMode  { return e.mode }
func (e *embedFileInfo) ModTime() time.Time { return e.modTime }
func (e *embedFileInfo) IsDir() bool        { return e.isDir }
func (e *embedFileInfo) Sys() interface{}   { return nil }

type embedFile struct {
	path string
	fs   *EmbedFs
	info *embedFileInfo
	io.ReadCloser
}

func (f embedFile) Read(b []byte) (int, error) {
	if f.ReadCloser != nil {
		return f.ReadCloser.Read(b)
	}
	return 0, io.EOF
}
func (f embedFile) Close() error {
	if f.ReadCloser != nil {
		return f.ReadCloser.Close()
	}
	return nil
}

func (f embedFile) Stat() (os.FileInfo, error) {
	return f.info, nil
}
func (f embedFile) Readdir(n int) ([]os.FileInfo, error) {
	if !f.info.isDir {
		return nil, nil
	}
	items := f.fs.searchDirItems(f.path, n)
	ditems := make([]os.FileInfo, len(items))
	for i := range items {
		ditems[i] = items[i]
	}
	return ditems, nil
}
func (f embedFile) Readdirnames(n int) ([]string, error) {
	if !f.info.isDir {
		return nil, nil
	}
	items := f.fs.searchDirItems(f.path, n)
	ditems := make([]string, len(items))
	for i := range items {
		ditems[i] = items[i].Name()
	}
	return ditems, nil
}

type sortedDirItems []*embedFileInfo

func (s sortedDirItems) Len() int {
	return len(s)
}
func (s sortedDirItems) Less(i, j int) bool {
	return s[i].name < s[j].name
}
func (s sortedDirItems) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type EmbedFs struct {
	files map[string]*embedFileInfo
}

func NewEmbedFs() *EmbedFs {
	return &EmbedFs{
		files: make(map[string]*embedFileInfo),
	}
}

func (e *EmbedFs) searchDirItems(dir string, n int) []*embedFileInfo {
	var items []*embedFileInfo
	for _, item := range e.files {
		if item.dir == dir {
			items = append(items, item)
		}
	}
	sort.Sort(sortedDirItems(items))
	if n > 0 && n < len(items) {
		items = items[:n]
	}
	return items
}

func (e *EmbedFs) Add(path string, size int64, mode os.FileMode, modTime time.Time, isDir bool, content func() (io.ReadCloser, error)) {
	path = filepath.ToSlash(path)
	path = cleanPath(path)
	var (
		dir  string
		name string
	)
	if path == "/" {
		name = path
	} else {
		idx := strings.LastIndex(path, "/")
		if idx >= 0 {
			dir = path[:idx]
			name = path[idx+1:]
		} else {
			name = path
		}
		if dir == "" {
			dir = "/"
		}
	}
	e.files[path] = &embedFileInfo{
		dir:     dir,
		name:    name,
		size:    size,
		mode:    mode,
		modTime: modTime,
		isDir:   isDir,
		content: content,
	}
}

func (e *EmbedFs) Stat(path string) (os.FileInfo, error) {
	path = filepath.ToSlash(path)
	path = cleanPath(path)
	info, has := e.files[path]
	if !has {
		return nil, os.ErrNotExist
	}
	return info, nil
}

func (e *EmbedFs) Open(path string) (File, error) {
	path = filepath.ToSlash(path)
	path = cleanPath(path)
	info, has := e.files[path]
	if !has {
		return nil, os.ErrNotExist
	}
	f := embedFile{
		path: path,
		info: info,
		fs:   e,
	}
	if !info.isDir {
		content, err := info.content()
		if err != nil {
			return nil, err
		}
		f.ReadCloser = content
	}
	return &f, nil
}
