package main

import (
	"compress/gzip"
	"io"
	"strings"
	"time"

	"github.com/cosiner/go-embedres"
)

var embedFs = embedres.NewEmbedFs()
var Fs embedres.Fs = embedFs

func init() {
	embedFs.Add(
		"/",
		224,
		2147484141,
		time.Unix(1541923938, 0).In(time.Local),
		true,
		nil,
	)
	embedFs.Add(
		"/assets",
		96,
		2147484141,
		time.Unix(1541922802, 0).In(time.Local),
		true,
		nil,
	)
	embedFs.Add(
		"/assets/index.html",
		10,
		420,
		time.Unix(1541921416, 0).In(time.Local),
		false,
		func() (io.ReadCloser, error) {
			return gzip.NewReader(strings.NewReader("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xca\xcc\x4b\x49\xad\xd0\xcb\x28\xc9\xcd\x01\x04\x00\x00\xff\xff\xc9\xe9\x4b\x77\x0a\x00\x00\x00"))
		},
	)
}
