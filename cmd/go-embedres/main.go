package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/cosiner/flag"
	embedres "github.com/cosiner/go-embedres"
)

type Flags struct {
	Pkg     string   `names:"--pkg" usage:"result file package name, directory name is used by default"`
	Output  string   `names:"-o, --output" usage:"output file"`
	Prefix  string   `names:"--prefix" usage:"remove path prefix, default current directory"`
	Ignores []string `names:"--ignore" usage:"ignore file pattern"`
	Paths   []string `names:"--path" usage:"embed file path"`
}

type EmbedFile struct {
	Path        string
	Size        int64
	Mode        uint32
	ModTimeUnix int64
	IsDir       bool
	Content     []byte
}
type RenderData struct {
	Package string
	Imports []string
	Files   []EmbedFile
}

func formatBytes(bs []byte) string {
	var buf bytes.Buffer
	buf.WriteString(`"`)
	for _, b := range bs {
		fmt.Fprintf(&buf, "\\x%02x", b)
	}
	buf.WriteString(`"`)
	return buf.String()
}

const outputFileTemplate = `
package {{.Package}}
import (
	{{ range .Imports }}
		"{{ . }}"
	{{- end }}
)
var embedFs = embedres.NewEmbedFs()
var Fs embedres.Fs = embedFs

func init() {
	{{- range $f := .Files }}
	embedFs.Add(
		"{{$f.Path}}", 
		{{$f.Size}}, 
		{{$f.Mode}}, 
		time.Unix({{$f.ModTimeUnix}}, 0).In(time.Local), 
		{{$f.IsDir}}, 
		{{- if $f.IsDir }} 
		nil,
		{{- else }} 
		func() (io.ReadCloser, error) {return gzip.NewReader(strings.NewReader({{$f.Content | formatBytes }}))},
		{{- end }}
	)
	{{- end }}
}
`

func main() {
	var flags Flags
	flag.ParseStruct(&flags)

	localFs, err := embedres.NewLocalFs(flags.Prefix, flags.Paths, flags.Ignores)
	if err != nil {
		log.Fatalln("create local fs failed:", err.Error())
		return
	}

	var renderData RenderData
	err = embedres.Walk(localFs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		f := EmbedFile{
			Path:        path,
			Size:        info.Size(),
			Mode:        uint32(info.Mode()),
			ModTimeUnix: info.ModTime().Unix(),
			IsDir:       info.IsDir(),
		}
		if !info.IsDir() {
			fd, err := localFs.Open(path)
			var buf bytes.Buffer
			if err != nil {
				return fmt.Errorf("open file failed: %s, %s", path, err.Error())
			}
			defer fd.Close()

			w := gzip.NewWriter(&buf)
			_, err = io.Copy(w, fd)
			if err != nil {
				return fmt.Errorf("read file content failed: %s, %s", path, err)
			}
			err = w.Close()
			if err != nil {
				return fmt.Errorf("read file content failed: %s, %s", path, err)
			}

			f.Content = buf.Bytes()
		}
		renderData.Files = append(renderData.Files, f)
		return nil
	})
	if err != nil {
		log.Fatalln("walk failed:", err.Error())
	}

	abspath, err := filepath.Abs(flags.Output)
	if err != nil {
		log.Fatalln("get absolute output path failed:", err.Error())
	}
	absdir := filepath.Dir(abspath)
	err = os.MkdirAll(absdir, 0755)
	if err != nil {
		log.Fatalln("create output path parent dirs failed:", err.Error())
	}

	renderData.Package = flags.Pkg
	if renderData.Package == "" {
		renderData.Package = filepath.Base(absdir)
	}
	for _, v := range []interface{}{
		strings.Reader{},
		gzip.Reader{},
		embedres.LocalFs{},
		io.LimitedReader{},
		time.Time{},
	} {
		renderData.Imports = append(renderData.Imports, reflect.TypeOf(v).PkgPath())
	}

	t := template.New("").Funcs(template.FuncMap{"formatBytes": formatBytes})
	_, err = t.Parse(outputFileTemplate)
	if err != nil {
		log.Fatalln("parse outout template failed:", err.Error())
	}

	fd, err := os.OpenFile(abspath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln("open output file failed:", err.Error())
	}
	defer fd.Close()

	err = t.Execute(fd, renderData)
	if err != nil {
		log.Fatalln("generate output failed:", err.Error())
	}
}
