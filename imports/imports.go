// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package imports provides an API to check whether code imports packages
// according to a whitelist/blacklist scheme.
package imports // import "gonum.org/v1/tools/imports"

import (
	"bufio"
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// CheckAllowed analyzes all Go files under dir for imports based on a
// whitelist/blacklist scheme.
// If CheckAllowed encounters multiple files importing non-allowed imports, the
// first error is returned to the user.
func CheckAllowed(dir string, whitelist, blacklist []string) error {
	// TODO: handle multiple errors.
	// TODO: add a limit of the number of errors to handle before bailing out.

	if len(whitelist) == 0 && len(blacklist) == 0 {
		return nil
	}

	whitelist, err := includeStd(whitelist)
	if err != nil {
		return err
	}
	whitepat, err := str2RE(whitelist)
	if err != nil {
		return err
	}
	blackpat, err := str2RE(blacklist)
	if err != nil {
		return err
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	fset := token.NewFileSet()
	for _, fname := range files {
		e := process(fname, fset, whitepat, blackpat)
		if e != nil {
			if err == nil {
				err = e
			}
		}
	}
	return err
}

func process(fname string, fset *token.FileSet, whitelist, blacklist []*regexp.Regexp) error {
	src, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	return checkImports(fset, src, fname, whitelist, blacklist)
}

func checkImports(fset *token.FileSet, src []byte, fname string, whitelist, blacklist []*regexp.Regexp) error {
	f, err := parser.ParseFile(fset, fname, src, parser.ImportsOnly)
	if err != nil {
		return err
	}

	imp := Error{File: fname}
	for _, s := range f.Imports {
		path := strings.Trim(s.Path.Value, `"`)
		if len(whitelist) != 0 && !listed(path, whitelist) {
			imp.Imports = append(imp.Imports, path)
		}
		if listed(path, blacklist) {
			imp.Imports = append(imp.Imports, path)
		}
	}
	if len(imp.Imports) > 0 {
		return imp
	}
	return nil
}

func listed(path string, list []*regexp.Regexp) bool {
	for _, v := range list {
		if v.MatchString(path) {
			return true
		}
	}
	return false
}

func str2RE(vs []string) ([]*regexp.Regexp, error) {
	if len(vs) == 0 {
		return nil, nil
	}
	var (
		err error
		o   = make([]*regexp.Regexp, len(vs))
	)
	for i, v := range vs {
		if !strings.HasPrefix(v, "^") {
			v = "^" + v
		}
		if !strings.HasSuffix(v, "$") {
			v += "$"
		}
		o[i], err = regexp.Compile(v)
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

func includeStd(list []string) ([]string, error) {
	if len(list) == 0 {
		return list, nil
	}

	wanted := true
	for i := 0; i < len(list); {
		if list[i] != "-std" {
			i++
			continue
		}
		wanted = false
		list[i] = list[len(list)-1]
		list = list[:len(list)-1]
	}
	if !wanted {
		return list, nil
	}

	pkgs, err := std()
	if err != nil {
		return nil, err
	}
	return append(list, pkgs...), nil
}

// std returns a slice of patterns matching the standard library.
func std() ([]string, error) {
	gocmd, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(gocmd, "list", "std")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	var pkgs []string
	sc := bufio.NewScanner(&buf)
	for sc.Scan() {
		pkg := sc.Text()
		switch strings.Split(pkg, "/")[0] {
		case "builtin", "cmd", "vendor", "internal", "testdata":
			continue
		}
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

// Error stores information about a disallowed import in a Go file.
type Error struct {
	File    string
	Imports []string
}

func (e Error) Error() string {
	return fmt.Sprintf(
		"%s: disallowed imports: %v",
		e.File,
		strings.Join(e.Imports, ", "),
	)
}
