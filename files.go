package main

import (
	"crypto/sha1"
	"fmt"
	"github.com/Jeffail/gabs/v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

// We will index files by their sha1 string repr
type hfiles map[string]*hfile

type hfile struct {
	Path string
	Blob gabs.Container
}

func HashFolder(dir string) hfiles {
	folderHfiles := make(hfiles)
	//files := make(map[string]string)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			singleFile, err := ioutil.ReadFile(path)
			if err != nil {
				fmt.Println(err)
			}
			h := sha1.New()
			h.Write(singleFile)
			digest := fmt.Sprintf("%x", h.Sum(nil))
			folderHfiles[digest] = &hfile{path, gabs.Container{}}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return folderHfiles
}
