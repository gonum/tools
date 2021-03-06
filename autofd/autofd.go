// Copyright ©2020 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autofd

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Func describes which function will be derived.
type Func struct {
	Path  string // Import path of the package holding the function.
	Name  string // Function or method name.
	Deriv string // Name of the output derivative function.
}

// Derivative generates code for derivatives from the given function declaration.
// If d2 is true, the generated function returns both the first and second
// derivatives. Otherwise, only the first derivative function is generated.
func Derivative(w io.Writer, f Func, d2 bool) error {
	gen, err := newGenerator(w, f, d2)
	if err != nil {
		return fmt.Errorf("could not create derivative generator: %w", err)
	}
	err = gen.generate()
	if err != nil {
		return fmt.Errorf("could not generate derivative: %w", err)
	}
	return nil
}

type kind bool

const (
	d1xKind kind = false
	d2xKind kind = true
)

func (k kind) dpkg() string {
	switch k {
	case d1xKind:
		return "dual"
	case d2xKind:
		return "hyperdual"
	}
	panic("impossible")
}

type generator struct {
	w    io.Writer
	pkg  *packages.Package
	fct  *types.Func
	kind kind
	xvar string
	der  string
	err  error
}

func newGenerator(w io.Writer, f Func, d2 bool) (*generator, error) {
	path := f.Path
	name := f.Name

	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, fmt.Errorf("could not load package of %q %s: %w", f.Path, f.Name, err)
	}

	var pkg *packages.Package
	for _, p := range pkgs {
		if p.PkgPath == path {
			pkg = p
			break
		}
	}

	if pkg == nil || len(pkg.Errors) > 0 {
		return nil, fmt.Errorf("could not find package %q", path)
	}

	var fct *types.Func
	scope := pkg.Types.Scope()
	switch {
	case strings.Contains(name, "."):
		idx := strings.Index(name, ".")
		obj := scope.Lookup(name[:idx])
		if obj == nil {
			return nil, fmt.Errorf("could not find %s in package %q", name[:idx], path)
		}
		typ, ok := obj.Type().(*types.Named)
		if !ok {
			return nil, fmt.Errorf(
				"object %s in package %q is not a named type (%T)",
				name[:idx], path, obj,
			)
		}
		for i := 0; i < typ.NumMethods(); i++ {
			m := typ.Method(i)
			if m.Name() == name[idx+1:] {
				fct = m
				break
			}
		}

		if fct == nil {
			return nil, fmt.Errorf("could not find %s in package %q", name, path)
		}

	default:
		obj := scope.Lookup(name)
		if obj == nil {
			return nil, fmt.Errorf("could not find %s in package %q", name, path)
		}
		var ok bool
		fct, ok = obj.(*types.Func)
		if !ok {
			return nil, fmt.Errorf("object %s in package %q is not a func (%T)", name, path, obj)
		}
	}

	if !types.Identical(fct.Type(), f1x.Type()) {
		return nil, fmt.Errorf("invalid function signature for %s", name)
	}

	der := f.Deriv
	if der == "" {
		der = "Deriv" + strings.Replace(f.Name, ".", "_", -1)
	}

	return &generator{w: w, pkg: pkg, fct: fct, kind: kind(d2), der: der}, nil
}

func (g *generator) generate() error {
	var fct *ast.FuncDecl
	for _, f := range g.pkg.Syntax {
		for i := range f.Decls {
			decl, ok := f.Decls[i].(*ast.FuncDecl)
			if !ok {
				continue
			}

			if decl.Name.Name == g.fct.Name() {
				fct = decl
				break
			}
		}
	}

	var (
		ret     *ast.ReturnStmt
		returns int
	)
	ast.Inspect(fct.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.ReturnStmt:
			returns++
			ret = stmt
		}
		return true
	})

	switch returns {
	case 0:
		return fmt.Errorf("could not find a return statement")
	case 1:
		// ok
	default:
		return fmt.Errorf("can not handle functions with multiple return statements")
	}

	switch len(ret.Results) {
	case 0:
		return fmt.Errorf("naked returns not supported")
	case 1:
		// ok
	default:
		return fmt.Errorf("too many return values")
	}

	args := g.fct.Type().Underlying().(*types.Signature).Params()
	g.xvar = args.At(0).Name()
	switch g.kind {
	case d1xKind:
		g.printf("func %s(%s float64) float64 {\n", g.der, g.xvar)
		g.printf("\tv := ")
		g.expr(ret.Results[0])
		g.printf("\n\treturn v.Emag\n")
	case d2xKind:
		g.printf("func %s(%s float64) (d1, d2 float64) {\n", g.der, g.xvar)
		g.printf("\tv := ")
		g.expr(ret.Results[0])
		g.printf("\n\treturn v.E1mag, v.E1E2mag\n")
	}
	g.printf("}\n")

	return g.err
}

func (g *generator) expr(expr ast.Expr) {
	if g.err != nil {
		return
	}

	switch expr := expr.(type) {
	default:
		g.err = fmt.Errorf("invalid expr type: %#v (%T)", expr, expr)
	case *ast.BasicLit:
		g.printf("%s.Number{Real:%s}", g.dpkg(), expr.Value)
	case *ast.Ident:
		switch expr.Name {
		case g.xvar:
			switch g.kind {
			case d1xKind:
				g.printf("%s.Number{Real:%s, Emag:1}", g.dpkg(), expr.Name)
			case d2xKind:
				g.printf("%s.Number{Real:%s, E1mag:1, E2mag:1}", g.dpkg(), expr.Name)
			}
		default:
			g.printf("%s.Number{Real:%s}", g.dpkg(), expr.Name)
		}
	case *ast.ParenExpr:
		g.printf("(")
		g.expr(expr.X)
		g.printf(")")
	case *ast.UnaryExpr:
		switch expr.Op {
		default:
			g.err = fmt.Errorf("invalid binary expression token %v", expr.Op)
		case token.ADD:
			// no op
		case token.SUB:
			g.printf("%[1]s.Mul(%[1]s.Number{Real:-1}, ", g.dpkg())
			g.expr(expr.X)
			g.printf(")")
		}
	case *ast.BinaryExpr:
		switch expr.Op {
		default:
			g.err = fmt.Errorf("invalid binary expression token %v", expr.Op)
		case token.ADD:
			g.printf("%s.Add(", g.dpkg())
			g.expr(expr.X)
			g.printf(", ")
			g.expr(expr.Y)
			g.printf(")")
		case token.SUB:
			g.printf("%s.Sub(", g.dpkg())
			g.expr(expr.X)
			g.printf(", ")
			g.expr(expr.Y)
			g.printf(")")
		case token.MUL:
			g.printf("%s.Mul(", g.dpkg())
			g.expr(expr.X)
			g.printf(", ")
			g.expr(expr.Y)
			g.printf(")")
		case token.QUO:
			g.printf("%s.Mul(", g.dpkg())
			g.expr(expr.X)
			g.printf(", ")
			g.printf("%s.Inv(", g.dpkg())
			g.expr(expr.Y)
			g.printf("))")
		}

	case *ast.CallExpr:
		g.expr(expr.Fun)
		g.printf("(")
		for i, arg := range expr.Args {
			if i > 0 {
				g.printf(", ")
			}
			g.expr(arg)
		}
		g.printf(")")

	case *ast.SelectorExpr:
		x, ok := expr.X.(*ast.Ident)
		if !ok || x.Name != "math" {
			g.err = fmt.Errorf("invalid selector expression %#v", expr)
		}
		switch expr.Sel.Name {
		case "Abs",
			"Acos", "Acosh",
			"Asin", "Asinh",
			"Atan", "Atanh",
			"Cos", "Cosh",
			"Exp", "Log",
			"Pow",
			"Sin", "Sinh",
			"Sqrt",
			"Tan", "Tanh":
			g.printf("%s.%s", g.dpkg(), expr.Sel.Name)
		case "E", "Pi", "Phi",
			"Sqrt2", "SqrtE", "SqrtPi", "SqrtPhi",
			"Ln2", "Log2E", "Ln10", "Log10E":
			g.printf("%s.Number{Real: math.%s}", g.dpkg(), expr.Sel.Name)
		default:
			g.err = fmt.Errorf("invalid selector expression %#v", expr)
		}
	}
}

func (g *generator) printf(format string, args ...interface{}) {
	fmt.Fprintf(g.w, format, args...)
}

func (g *generator) dpkg() string {
	return g.kind.dpkg()
}

// f1x is the pre-computed signature of 'func(float64) float64'.
// This will be checked against to make sure Derivative is called on valid functions.
var f1x *types.Func

func init() {
	const variadic = false
	f64 := types.NewParam(0, nil, "x", types.Typ[types.Float64])

	sig := types.NewSignature(nil, types.NewTuple(f64), types.NewTuple(f64), variadic)
	f1x = types.NewFunc(0, nil, "f1x", sig)
}
