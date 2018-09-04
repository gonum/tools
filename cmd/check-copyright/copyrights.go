// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// copyrights checks Go files for copyright headers based on a regexp.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	notice := flag.String("notice", "Copyright", "header notice to look for above package clause")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [-notice <regexp>] [<package path>...]\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	copyright, err := regexp.Compile(*notice)
	if err != nil {
		log.Fatalf("could not compile notice regexp: %v", err)
	}

	var missing bool
	dirs := flag.Args()
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	for _, dir := range dirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			switch {
			case info.IsDir():
				switch info.Name() {
				case "testdata":
					return filepath.SkipDir
				}
			default:
				if filepath.Ext(info.Name()) != ".go" {
					return nil
				}
				fset := token.NewFileSet()
				ok, err := hasCopyrightHeader(path, fset, copyright)
				if err != nil {
					log.Fatalf("could not check %q: %v", path, err)
				}
				if !ok {
					missing = true
					fmt.Println(path)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("error during walk: %v", err)
		}
	}
	if missing {
		os.Exit(3)
	}
}

func hasCopyrightHeader(fname string, fset *token.FileSet, copyright *regexp.Regexp) (ok bool, err error) {
	src, err := ioutil.ReadFile(fname)
	if err != nil {
		return false, err
	}
	f, err := parser.ParseFile(fset, fname, src, parser.ParseComments|parser.PackageClauseOnly)
	if err != nil {
		return false, err
	}

	for _, cg := range f.Comments {
		var comment bytes.Buffer
		var inComment bool
		for _, c := range cg.List {
			text := c.Text
			if !inComment {
				switch {
				case strings.HasPrefix(text, "/*"):
					inComment = true
					fallthrough
				case strings.HasPrefix(text, "//"):
					text = text[2:]
					if strings.HasPrefix(text, " ") {
						text = text[1:]
					}
				}
			}
			if inComment {
				if strings.HasSuffix(text, "*/") {
					inComment = false
					text = text[:len(text)-2]
					if strings.HasSuffix(text, " ") {
						text = text[:len(text)-1]
					}
				}
			}
			fmt.Fprintln(&comment, text)
		}
		if copyright.Match(comment.Bytes()) {
			return true, nil
		}
	}
	return false, nil
}
