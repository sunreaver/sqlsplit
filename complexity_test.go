package sqlsplit

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func calcComplexity(fn *ast.FuncDecl) int {
	complexity := 1
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch node := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
			complexity++
		case *ast.CaseClause:
			if len(node.List) > 0 {
				complexity++
			}
		case *ast.CommClause:
			if node.Comm != nil {
				complexity++
			}
		case *ast.BinaryExpr:
			if node.Op == token.LAND || node.Op == token.LOR {
				complexity++
			}
		}
		return true
	})
	return complexity
}

func TestComplexity(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}

	fset := token.NewFileSet()
	hasViolation := false
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(fset, file, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				comp := calcComplexity(fn)
				name := fn.Name.Name
				if fn.Recv != nil && len(fn.Recv.List) > 0 {
					recvType := fn.Recv.List[0].Type
					if star, ok := recvType.(*ast.StarExpr); ok {
						if id, ok := star.X.(*ast.Ident); ok {
							name = id.Name + "." + name
						}
					} else if id, ok := recvType.(*ast.Ident); ok {
						name = id.Name + "." + name
					}
				}
				if comp > 10 {
					t.Logf("VIOLATION: %s in %s has cyclomatic complexity %d (> 10)", name, file, comp)
					hasViolation = true
				} else {
					t.Logf("PASS: %s in %s has cyclomatic complexity %d", name, file, comp)
				}
			}
		}
	}
	if hasViolation {
		t.Errorf("One or more methods exceed cyclomatic complexity 10")
	}
}
