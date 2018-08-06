// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command check-imports inspects a source tree for imports satisfying a
// whitelist/blacklist scheme. If a whitelist is included, packages from
// the standard library are included. The order of application is whitelist
// then blacklist.
//
// If the standard library is not wanted in the whitelist, the pseudo-package
// "-std" can be specified in the whitelist to exclude it.
//
// Example:
//
//  $> check-imports -b="github.com/gonum/.*,math/rand"
//  $> check-imports -b="github.com/gonum/.*,math/rand" .
//  $> check-imports -b="github.com/gonum/.*,math/rand" /some/dir /other/dir
//  $> check-imports -w="github.com/.*" -b="github.com/gonum/.*" /some/dir /other/dir
package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"gonum.org/v1/tools/imports"
)

func main() {
	log.SetPrefix("check-imports: ")
	log.SetFlags(0)

	wlist := flag.String("w", "", "comma-separated list of whitelisted imports")
	blist := flag.String("b", "", "comma-separated list of blacklisted imports")

	flag.Parse()

	if *wlist == "" && *blist == "" {
		flag.Usage()
		log.Fatalf("missing white/blacklist of imports")
	}

	switch flag.NArg() {
	case 0:
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("could not retrieve current working directory: %v", err)
		}
		log.Printf("analyzing imports under %q...", dir)
		err = imports.CheckAllowed(dir, split(*wlist), split(*blist))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("analyzing imports under %q... [OK]", dir)
	default:
		for _, dir := range flag.Args() {
			log.Printf("analyzing imports under %q...", dir)
			err := imports.CheckAllowed(dir, split(*wlist), split(*blist))
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("analyzing imports under %q... [OK]", dir)
		}
	}
}

func split(list string) []string {
	if len(list) == 0 {
		return nil
	}
	return strings.Split(list, ",")
}
