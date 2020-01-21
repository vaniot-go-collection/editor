package main

// Needs to run when there are changes on the ../debug pkg.
//go:generate go run debugpack.go

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

// Pack debug package into a single file to be included by godebug.
// This file is called from the godebug pkg with "go generate"
func main() {
	if err := main2(); err != nil {
		fmt.Println(err)
	}
}
func main2() error {
	dir := "../debug"
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	filenames := []string{}
	data := []string{}
	for _, fi := range fis {
		// don't add test files to pack
		if strings.HasSuffix(fi.Name(), "_test.go") {
			continue
		}

		filename := filepath.Join(dir, fi.Name())
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		filenames = append(filenames, fi.Name())
		data = append(data, strconv.QuoteToASCII(string(b)))
	}
	if err := buildDataFile(filenames, data, "../zdebugpack.go"); err != nil {
		return err
	}
	return nil
}

func buildDataFile(filenames []string, data []string, filename string) error {
	z := []string{}
	for i := range filenames {
		e := fmt.Sprintf("{%q, %s}", filenames[i], data[i])
		z = append(z, e)
	}

	w := `package godebug

// DO NOT EDIT: code generated by debugpack/debugpack.go

type FilePack struct {
	Name string
	Data string
}

func DebugFilePacks() []*FilePack {
	return []*FilePack{%s}
}
`
	u := fmt.Sprintf(w, strings.Join(z, ",\n\t\t"))
	return ioutil.WriteFile(filename, []byte(u), 0644)
}
