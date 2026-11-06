package core

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// TestEveryCoreRouteIsInTheOpenAPIContract reads the route checks in http.go
// and fails when one of them has no matching entry in contracts/openapi/local.json.
// An exact path must be documented as is; a prefix must start a documented path;
// a prefix and suffix pair must frame one templated segment, as in /v1/goals/{id}/plan.
// When the same condition fixes the method, the documented path must carry it.
func TestEveryCoreRouteIsInTheOpenAPIContract(t *testing.T) {
	raw, err := os.ReadFile("../../contracts/openapi/local.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &contract); err != nil {
		t.Fatal(err)
	}
	file, err := parser.ParseFile(token.NewFileSet(), "http.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	documented := func(method string, match func(path string) bool) bool {
		for path, operations := range contract.Paths {
			if !match(path) {
				continue
			}
			if _, ok := operations[strings.ToLower(method)]; method == "" || ok {
				return true
			}
		}
		return false
	}

	checked := 0
	ast.Inspect(file, func(node ast.Node) bool {
		stmt, ok := node.(*ast.IfStmt)
		if !ok {
			return true
		}
		route := routeCheck{}
		for _, term := range andTerms(stmt.Cond) {
			route.add(term)
		}
		if !strings.HasPrefix(route.exact+route.prefix, "/v1/") {
			return true
		}
		checked++
		var want string
		var match func(string) bool
		switch {
		case route.exact != "":
			want = route.exact
			match = func(path string) bool { return path == route.exact }
		case route.suffix != "":
			want = route.prefix + "{id}" + route.suffix
			pattern := regexp.MustCompile("^" + regexp.QuoteMeta(route.prefix) + `\{[^/}]+\}` + regexp.QuoteMeta(route.suffix) + "$")
			match = pattern.MatchString
		default:
			want = route.prefix + "..."
			match = func(path string) bool { return strings.HasPrefix(path, route.prefix) }
		}
		if !documented(route.method, match) {
			t.Errorf("http.go route %s is not in contracts/openapi/local.json", strings.TrimSpace(strings.ToUpper(route.method)+" "+want))
		}
		return true
	})
	if checked == 0 {
		t.Fatal("found no /v1/ route checks in http.go")
	}
}

// routeCheck gathers what one && condition in ServeHTTP says about a request.
type routeCheck struct {
	method, exact, prefix, suffix string
}

func (c *routeCheck) add(term ast.Expr) {
	switch expr := term.(type) {
	case *ast.BinaryExpr:
		if expr.Op != token.EQL {
			return
		}
		for _, sides := range [][2]ast.Expr{{expr.X, expr.Y}, {expr.Y, expr.X}} {
			switch exprString(sides[0]) {
			case "r.Method":
				c.method = strings.TrimPrefix(exprString(sides[1]), "http.Method")
			case "r.URL.Path":
				c.exact = stringLiteral(sides[1])
			}
		}
	case *ast.CallExpr:
		if len(expr.Args) != 2 || exprString(expr.Args[0]) != "r.URL.Path" {
			return
		}
		switch exprString(expr.Fun) {
		case "strings.HasPrefix":
			c.prefix = stringLiteral(expr.Args[1])
		case "strings.HasSuffix":
			c.suffix = stringLiteral(expr.Args[1])
		}
	}
}

func andTerms(expr ast.Expr) []ast.Expr {
	if binary, ok := expr.(*ast.BinaryExpr); ok && binary.Op == token.LAND {
		return append(andTerms(binary.X), andTerms(binary.Y)...)
	}
	return []ast.Expr{expr}
}

func exprString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return exprString(e.X) + "." + e.Sel.Name
	}
	return ""
}

func stringLiteral(expr ast.Expr) string {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return ""
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil {
		return ""
	}
	return value
}
