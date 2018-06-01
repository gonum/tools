// Copyright ©2018 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command check-imports inspects a source tree for blacklisted imports.
//
// Example:
//
//  $> check-imports -b="github.com/gonum/.*,math/rand"
//  $> check-imports -b="github.com/gonum/.*,math/rand" .
//  $> check-imports -b="github.com/gonum/.*,math/rand" /some/dir /other/dir
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

	blist := flag.String("b", "", "comma-separated list of blacklisted imports")

	flag.Parse()

	if *blist == "" {
		flag.Usage()
		log.Fatalf("missing blacklist of imports")
	}

	switch flag.NArg() {
	case 0:
		dir, err := os.Getwd()
		if err != nil {
			log.Fatalf("could not retrieve current working directory: %v", err)
		}
		log.Printf("analyzing imports under %q...", dir)
		err = imports.CheckBlacklisted(dir, strings.Split(*blist, ","))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("analyzing imports under %q... [OK]", dir)
	default:
		for _, dir := range flag.Args() {
			log.Printf("analyzing imports under %q...", dir)
			err := imports.CheckBlacklisted(dir, strings.Split(*blist, ","))
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("analyzing imports under %q... [OK]", dir)
		}
	}
}
