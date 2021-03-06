// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autofd_test

import (
	"fmt"
	"strings"
	"testing"

	"gonum.org/v1/tools/autofd"
)

func TestDerivative(t *testing.T) {
	for _, test := range derivativeTests {
		name := fmt.Sprintf("%s.%s", test.name.Path, test.name.Name)
		if test.name.Deriv != "" {
			name += "-" + test.name.Deriv
		}
		switch {
		case test.d2x:
			name += "-d2x"
		default:
			name += "-d1x"
		}
		t.Run(name, func(t *testing.T) {
			buf := new(strings.Builder)
			err := autofd.Derivative(buf, test.name, test.d2x)
			switch {
			case err != nil && test.err != nil:
				if got, want := err.Error(), test.err.Error(); got != want {
					t.Fatalf("invalid error.\ngot= %v\nwant=%v\n", got, want)
				}
			case err != nil && test.err == nil:
				t.Fatalf("could not generate derivative: %+v", err)
			case err == nil && test.err != nil:
				t.Fatalf("got=%v, want=%v", err, test.err)
			case err == nil && test.err == nil:
				if got, want := buf.String(), test.want; got != want {
					t.Fatalf("invalid derivative:\ngot:\n%s\nwant:\n%s\n", got, want)
				}
			}
		})
	}
}

var derivativeTests = []struct {
	name autofd.Func
	d2x  bool
	want string
	err  error
}{
	// first derivatives
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1"},
		want: `func DerivF1(x float64) float64 {
	v := dual.Mul(dual.Number{Real:x, Emag:1}, dual.Number{Real:x, Emag:1})
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1", Deriv: "DxF1"},
		want: `func DxF1(x float64) float64 {
	v := dual.Mul(dual.Number{Real:x, Emag:1}, dual.Number{Real:x, Emag:1})
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T1.F", Deriv: "DxF"},
		want: `func DxF(x float64) float64 {
	v := dual.Add(dual.Add(dual.Mul(dual.Number{Real:2}, dual.Number{Real:x, Emag:1}), dual.Mul(dual.Mul(dual.Number{Real:3}, dual.Number{Real:x, Emag:1}), dual.Number{Real:x, Emag:1})), dual.Mul(dual.Number{Real:4}, dual.Pow(dual.Number{Real:x, Emag:1}, dual.Number{Real:3})))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T2.F", Deriv: "DxF"},
		want: `func DxF(x float64) float64 {
	v := dual.Add(dual.Add(dual.Mul(dual.Number{Real:2}, dual.Number{Real:x, Emag:1}), dual.Mul(dual.Mul(dual.Number{Real:3}, dual.Number{Real:x, Emag:1}), dual.Number{Real:x, Emag:1})), dual.Mul(dual.Number{Real:4}, dual.Pow(dual.Number{Real:x, Emag:1}, dual.Number{Real:3})))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F2"},
		want: `func DerivF2(y float64) float64 {
	v := dual.Mul(dual.Number{Real:y, Emag:1}, dual.Number{Real:y, Emag:1})
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F3"},
		want: `func DerivF3(x float64) float64 {
	v := dual.Mul(dual.Mul(dual.Number{Real:2}, dual.Number{Real:x, Emag:1}), dual.Number{Real:x, Emag:1})
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F4"},
		want: `func DerivF4(x float64) float64 {
	v := dual.Mul(dual.Number{Real:2}, dual.Inv((dual.Mul(dual.Number{Real:x, Emag:1}, dual.Number{Real:x, Emag:1}))))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F5"},
		want: `func DerivF5(x float64) float64 {
	v := dual.Mul(dual.Number{Real:2}, dual.Inv((dual.Mul(dual.Number{Real:x, Emag:1}, dual.Mul(dual.Number{Real:-1}, dual.Number{Real:x, Emag:1})))))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F6"},
		want: `func DerivF6(x float64) float64 {
	v := dual.Sub(dual.Add(dual.Number{Real:2}, dual.Number{Real:x, Emag:1}), dual.Number{Real:x, Emag:1})
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F7"},
		want: `func DerivF7(x float64) float64 {
	v := dual.Cos(dual.Mul(dual.Mul(dual.Number{Real:2}, dual.Number{Real: math.Pi}), dual.Number{Real:x, Emag:1}))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F8"},
		want: `func DerivF8(x float64) float64 {
	v := dual.Mul(dual.Exp(dual.Number{Real:x, Emag:1}), dual.Inv(dual.Sqrt(dual.Add(dual.Pow(dual.Sin(dual.Number{Real:x, Emag:1}), dual.Number{Real:3}), dual.Pow(dual.Cos(dual.Number{Real:x, Emag:1}), dual.Number{Real:3})))))
	return v.Emag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F9"},
		want: `func DerivF9(x float64) float64 {
	v := dual.Mul(dual.Number{Real:pi}, dual.Number{Real:x, Emag:1})
	return v.Emag
}
`,
	},
	// second derivatives
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1"},
		d2x:  true,
		want: `func DerivF1(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1", Deriv: "DxF1"},
		d2x:  true,
		want: `func DxF1(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T1.F", Deriv: "DxF"},
		d2x:  true,
		want: `func DxF(x float64) (d1, d2 float64) {
	v := hyperdual.Add(hyperdual.Add(hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Mul(hyperdual.Mul(hyperdual.Number{Real:3}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:x, E1mag:1, E2mag:1})), hyperdual.Mul(hyperdual.Number{Real:4}, hyperdual.Pow(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:3})))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T2.F", Deriv: "DxF"},
		d2x:  true,
		want: `func DxF(x float64) (d1, d2 float64) {
	v := hyperdual.Add(hyperdual.Add(hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Mul(hyperdual.Mul(hyperdual.Number{Real:3}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:x, E1mag:1, E2mag:1})), hyperdual.Mul(hyperdual.Number{Real:4}, hyperdual.Pow(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:3})))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F2"},
		d2x:  true,
		want: `func DerivF2(y float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:y, E1mag:1, E2mag:1}, hyperdual.Number{Real:y, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F3"},
		d2x:  true,
		want: `func DerivF3(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:x, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F4"},
		d2x:  true,
		want: `func DerivF4(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Inv((hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}))))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F5"},
		d2x:  true,
		want: `func DerivF5(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Inv((hyperdual.Mul(hyperdual.Number{Real:x, E1mag:1, E2mag:1}, hyperdual.Mul(hyperdual.Number{Real:-1}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})))))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F6"},
		d2x:  true,
		want: `func DerivF6(x float64) (d1, d2 float64) {
	v := hyperdual.Sub(hyperdual.Add(hyperdual.Number{Real:2}, hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:x, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F7"},
		d2x:  true,
		want: `func DerivF7(x float64) (d1, d2 float64) {
	v := hyperdual.Cos(hyperdual.Mul(hyperdual.Mul(hyperdual.Number{Real:2}, hyperdual.Number{Real: math.Pi}), hyperdual.Number{Real:x, E1mag:1, E2mag:1}))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F8"},
		d2x:  true,
		want: `func DerivF8(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Exp(hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Inv(hyperdual.Sqrt(hyperdual.Add(hyperdual.Pow(hyperdual.Sin(hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:3}), hyperdual.Pow(hyperdual.Cos(hyperdual.Number{Real:x, E1mag:1, E2mag:1}), hyperdual.Number{Real:3})))))
	return v.E1mag, v.E1E2mag
}
`,
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F9"},
		d2x:  true,
		want: `func DerivF9(x float64) (d1, d2 float64) {
	v := hyperdual.Mul(hyperdual.Number{Real:pi}, hyperdual.Number{Real:x, E1mag:1, E2mag:1})
	return v.E1mag, v.E1E2mag
}
`,
	},
	// errors
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF1"},
		err:  fmt.Errorf("could not create derivative generator: invalid function signature for ErrF1"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF2"},
		err:  fmt.Errorf("could not create derivative generator: invalid function signature for ErrF2"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF3"},
		err:  fmt.Errorf("could not create derivative generator: invalid function signature for ErrF3"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF4"},
		err:  fmt.Errorf("could not create derivative generator: invalid function signature for ErrF4"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF5"},
		err:  fmt.Errorf("could not generate derivative: naked returns not supported"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF6"},
		err:  fmt.Errorf("could not generate derivative: can not handle functions with multiple return statements"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF7"},
		err:  fmt.Errorf("could not generate derivative: can not handle functions with multiple return statements"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrF8"},
		err:  fmt.Errorf("could not generate derivative: can not handle functions with multiple return statements"),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfuncXXX", Name: "F1"},
		err:  fmt.Errorf(`could not create derivative generator: could not find package "gonum.org/v1/tools/autofd/internal/testfuncXXX"`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1xxx"},
		err:  fmt.Errorf(`could not create derivative generator: could not find F1xxx in package "gonum.org/v1/tools/autofd/internal/testfunc"`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "Fxxx.F"},
		err:  fmt.Errorf(`could not create derivative generator: could not find Fxxx in package "gonum.org/v1/tools/autofd/internal/testfunc"`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "F1.F"},
		err:  fmt.Errorf(`could not create derivative generator: object F1 in package "gonum.org/v1/tools/autofd/internal/testfunc" is not a named type (*types.Func)`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T1"},
		err:  fmt.Errorf(`could not create derivative generator: object T1 in package "gonum.org/v1/tools/autofd/internal/testfunc" is not a func (*types.TypeName)`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "T1.Fxxx"},
		err:  fmt.Errorf(`could not create derivative generator: could not find T1.Fxxx in package "gonum.org/v1/tools/autofd/internal/testfunc"`),
	},
	{
		name: autofd.Func{Path: "gonum.org/v1/tools/autofd/internal/testfunc", Name: "ErrT1.F"},
		err:  fmt.Errorf(`could not create derivative generator: could not find ErrT1.F in package "gonum.org/v1/tools/autofd/internal/testfunc"`),
	},
}
