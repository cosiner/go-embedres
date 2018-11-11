package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/cosiner/go-embedres"
)

func main() {
	err := embedres.Walk(Fs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		log.Printf("%s %s %d\n", info.Mode(), path, info.Size())

		if !info.IsDir() {
			fd, err := Fs.Open(path)
			if err != nil {
				log.Fatalln("open file failed:", err.Error())
			}
			defer fd.Close()
			content, err := ioutil.ReadAll(fd)
			if err != nil {
				log.Fatalln("read file failed:", err.Error())
			}
			os.Stdout.Write(content)
		}
		return nil
	})
	if err != nil {
		log.Fatalln("walk embed fs failed:", err.Error())
	}
}
