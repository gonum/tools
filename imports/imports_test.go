// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package imports

import (
	"fmt"
	"go/token"
	"reflect"
	"regexp"
	"testing"
)

var blacklist = []string{
	"github.com/gonum/.*", // prefer gonum.org/v1/gonum
	"math/rand",           // prefer golang.org/x/exp/rand
}

var checkTests = []struct {
	whitelist, blacklist []string

	pkg string
	err error
}{
	{
		pkg: "math/rand",
		err: nil,
	},
	{
		pkg: "math/rands",
		err: nil,
	},
	{
		pkg: "math",
		err: nil,
	},
	{
		blacklist: blacklist,
		pkg:       "math/rand",
		err: Error{
			File:    "file.go",
			Imports: []string{"math/rand"},
		},
	},
	{
		blacklist: blacklist,
		pkg:       "math/rands",
		err:       nil,
	},
	{
		blacklist: blacklist,
		pkg:       "math",
		err:       nil,
	},
	{
		blacklist: blacklist,
		pkg:       "github.com/gonum/",
		err: Error{
			File:    "file.go",
			Imports: []string{"github.com/gonum/"},
		},
	},
	{
		blacklist: blacklist,
		pkg:       "github.com/gonum/floats",
		err: Error{
			File:    "file.go",
			Imports: []string{"github.com/gonum/floats"},
		},
	},
	{
		blacklist: blacklist,
		pkg:       "github.com/gonum/plot",
		err: Error{
			File:    "file.go",
			Imports: []string{"github.com/gonum/plot"},
		},
	},
	{
		blacklist: blacklist,
		pkg:       "gonum.org/v1/gonum/floats",
		err:       nil,
	},
	{
		blacklist: blacklist,
		pkg:       "gonum.org/v1/plot",
		err:       nil,
	},
	{
		blacklist: blacklist,
		pkg:       "github.com/gonumnum/floats",
		err:       nil,
	},
	{
		whitelist: []string{"-std", "golang.org/x/exp/rand"}, // exclude std for testing.
		pkg:       "math/rand",
		err: Error{
			File:    "file.go",
			Imports: []string{"math/rand"},
		},
	},
	{
		whitelist: []string{"pkg"},
		blacklist: []string{"math/rand"},
		pkg:       "math/rand",
		err: Error{
			File:    "file.go",
			Imports: []string{"math/rand"},
		},
	},
	{
		whitelist: []string{"pkg"},
		pkg:       "os",
		err:       nil,
	},
	{
		whitelist: []string{"-std", "pkg"}, // exclude std for testing.
		pkg:       "os",
		err: Error{
			File:    "file.go",
			Imports: []string{"os"},
		},
	},
}

func TestCheck(t *testing.T) {
	fset := token.NewFileSet()
	for _, tc := range checkTests {
		whitelist, err := includeStd(tc.whitelist)
		if err != nil {
			t.Fatal(err)
		}
		var whitepat, blackpat []*regexp.Regexp
		if len(tc.whitelist) != 0 {
			whitepat, err = str2RE(whitelist)
			if err != nil {
				t.Fatal(err)
			}
		}
		if len(tc.blacklist) != 0 {
			blackpat, err = str2RE(tc.blacklist)
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Run("", func(t *testing.T) {
			src := fmt.Sprintf("package foo\nimport _ %q\n", tc.pkg)
			err := checkImports(fset, []byte(src), "file.go", whitepat, blackpat)
			if !reflect.DeepEqual(err, tc.err) {
				t.Fatalf("error\ngot= %v\nwant=%v", err, tc.err)
			}
		})
	}
}
