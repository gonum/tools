// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Command autofd generates a function derivative from a given function or method
// location.
package main // import "gonum.org/v1/tools/cmd/autofd"

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gonum.org/v1/tools/autofd"
)

func main() {
	log.SetPrefix("autofd: ")
	log.SetFlags(0)

	pkg := flag.String("pkg", "", "import path of the package holding the function or method definition")
	fct := flag.String("fct", "", "name of the function or method definition")
	d2 := flag.Bool("d2", false, "whether to generate both first and second derivatives")
	der := flag.String("der", "", "name of the derivative to generate")

	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: autofd [options]

ex:
 $> autofd -pkg gonum.org/v1/tools/autofd/internal/testfunc -fct F1
 func DerivF1(x float64) float64 {
 	v := dual.Mul(dual.Number{Real:x, Emag:1}, dual.Number{Real:x, Emag:1})
 	return v.Emag
 }

 $> autofd -pkg gonum.org/v1/tools/autofd/internal/testfunc -fct F1 -d2
 func DerivF1(x float64) (d1, d2 float64) {
 	v := hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})
 	return v.E1mag, v.E1E2mag
 }

 $> autofd -pkg gonum.org/v1/tools/autofd/internal/testfunc -fct F1 -der DxF1 -d2
 func DxF1(x float64) (d1, d2 float64) {
 	v := hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})
 	return v.E1mag, v.E1E2mag
 }

 $> autofd -pkg gonum.org/v1/tools/autofd/internal/testfunc -fct T1.F

Options:
`,
		)
		flag.PrintDefaults()
	}

	flag.Parse()

	switch {
	case *pkg == "":
		flag.Usage()
		log.Fatalf("missing import path")
	case *fct == "":
		flag.Usage()
		log.Fatalf("missing function or method name")
	}

	err := autofd.Derivative(os.Stdout, autofd.Func{
		Path:  *pkg,
		Name:  *fct,
		Deriv: *der,
	}, *d2)
	if err != nil {
		log.Fatalf("could not generate derivative of %s.%s: %+v",
			*pkg, *fct, err,
		)
	}
}
