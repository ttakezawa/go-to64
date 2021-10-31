package main

import (
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(Analyzer)
}

const Doc = `Convert int/uint to int64/uint64.`

var Analyzer = &analysis.Analyzer{
	Name:             "convertTo64bit",
	Doc:              Doc,
	Run:              run,
	RunDespiteErrors: true,
	FactTypes:        []analysis.Fact{new(noReturn)},
}

var skipLiteral = false

func run(pass *analysis.Pass) (interface{}, error) {
	var fn func(n ast.Node) bool
	fn = func(n ast.Node) bool {
		if decl, ok := n.(*ast.GenDecl); ok {
			switch decl.Tok {
			case token.VAR, token.CONST:
				for _, spec := range decl.Specs {
					if valueSpec, ok := spec.(*ast.ValueSpec); ok {
						if valueSpec.Type != nil {
							ast.Inspect(valueSpec.Type, fn)
							// If a type is specified in lhs, don't fix constants in rhs
							skipLiteral = true
							for _, value := range valueSpec.Values {
								ast.Inspect(value, fn)
							}
							skipLiteral = false
						} else {
							// If it's CONST && a type is not specified, don't fix constants in rhs
							if decl.Tok == token.CONST {
								skipLiteral = true
								for _, value := range valueSpec.Values {
									ast.Inspect(value, fn)
								}
								skipLiteral = false
							} else {
								// Else, fix normally in rhs
								for _, value := range valueSpec.Values {
									ast.Inspect(value, fn)
								}
							}
						}
					}
				}
				return false
			default:
				return true
			}
		}

		if callFn, ok := n.(*ast.CallExpr); ok {
			if ident, ok := callFn.Fun.(*ast.Ident); ok {
				switch ident.Name {
				case "int64":
					return false

				case "len": // Convert return value of builtin len
					msg := fmt.Sprintf("Fix len(...) -> int64(len(...))")
					pass.Report(analysis.Diagnostic{
						Pos:     callFn.Pos(),
						End:     callFn.End(),
						Message: msg,
						SuggestedFixes: []analysis.SuggestedFix{{
							Message: msg,
							TextEdits: []analysis.TextEdit{
								{
									Pos:     callFn.Pos(),
									End:     callFn.Pos(),
									NewText: []byte("int64("),
								},
								{
									Pos:     callFn.End(),
									End:     callFn.End(),
									NewText: []byte(")"),
								},
							},
						}},
					})
					return false
				}
			}

			if fun, ok := callFn.Fun.(*ast.SelectorExpr); ok {
				if pkgIdent, ok := fun.X.(*ast.Ident); ok {
					switch {
					case pkgIdent.Name == "bits" && fun.Sel.Name == "OnesCount":
						msg := fmt.Sprintf("Fix bits.OnesCount(...) -> int64(bits.OnesCount64(...))")
						pass.Report(analysis.Diagnostic{
							Pos:     fun.Pos(),
							End:     fun.End(),
							Message: msg,
							SuggestedFixes: []analysis.SuggestedFix{
								{
									Message: msg,
									TextEdits: []analysis.TextEdit{
										{
											Pos:     callFn.Pos(),
											End:     callFn.Pos(),
											NewText: []byte("int64("),
										},
										{
											Pos:     callFn.End(),
											End:     callFn.End(),
											NewText: []byte(")"),
										},
										{
											Pos:     fun.Sel.Pos(),
											End:     fun.Sel.End(),
											NewText: []byte("OnesCount64"),
										},
									},
								},
							},
						})
						return true
					}
				}
			}
		}

		if !skipLiteral {
			if literal, ok := n.(*ast.BasicLit); ok && literal.Kind == token.INT {
				msg := fmt.Sprintf("Fix literal %s -> int64(%s)", literal.Value, literal.Value)
				pass.Report(analysis.Diagnostic{
					Pos:     literal.Pos(),
					End:     literal.End(),
					Message: msg,
					SuggestedFixes: []analysis.SuggestedFix{{
						Message: msg,
						TextEdits: []analysis.TextEdit{{
							Pos:     literal.Pos(),
							End:     literal.End(),
							NewText: []byte(fmt.Sprintf("int64(%s)", literal.Value)),
						}},
					}},
				})
				return true
			}
		}

		// replace int/uint with int64/uint64
		if ident, ok := n.(*ast.Ident); ok {
			switch ident.Name {
			case "int", "uint":
				msg := fmt.Sprintf("Fix %s -> %s64", ident.Name, ident.Name)
				pass.Report(analysis.Diagnostic{
					Pos:     ident.Pos(),
					End:     ident.End(),
					Message: msg,
					SuggestedFixes: []analysis.SuggestedFix{{
						Message: msg,
						TextEdits: []analysis.TextEdit{{
							Pos:     ident.Pos(),
							End:     ident.End(),
							NewText: []byte(fmt.Sprintf("%s64", ident.Name)),
						}},
					}},
				})
				return true
			}
		}

		return true
	}

	for _, f := range pass.Files {
		if f.Name.Name != "main" {
			continue
		}
		ast.Inspect(f, fn)
	}
	return nil, nil
}

// noReturn is a fact indicating that a function does not return.
type noReturn struct{}

func (*noReturn) AFact() {}

func (*noReturn) String() string { return "noReturn" }
