// Package main
//
// # Purpose
//
// Multichecker combines multiple static analyzers to check code for
// quality standards, potential errors, and security issues.
//
// # Running
//
// To run the analyzer, execute:
//
//	go run cmd/staticlint/main.go ./...
//
// Or build a binary file:
//
//	go build -o staticlint cmd/staticlint/main.go
//	./staticlint ./...
//
// # Analyzer Composition
//
// Multichecker includes the following analyzer groups:
//
// ## 1. Standard analyzers from golang.org/x/tools/go/analysis/passes
//
// Selected key analyzers:
//
//   - printf: checks correctness of formatting in Printf functions
//   - shadow: detects variable shadowing
//   - structtag: validates struct field tags
//   - unusedresult: finds unused results of function calls
//
// ## 2. SA-class analyzers from staticcheck.io
//
// Key SA analyzers:
//
//   - SA1000: invalid regular expression
//   - SA1019: using deprecated APIs
//   - SA2000: sync.WaitGroup misuse
//   - SA4000: meaningless comparisons
//   - SA5000: assignment to nil map
//   - SA6000: regexp optimization
//
// ## 3. Additional analyzers from staticcheck.io
//
//   - ST1003: poor naming choice
//   - QF1001: apply De Morgan's law for simplification
//
// ## 4. Custom analyzer
//
//   - noosexit: prohibits direct os.Exit calls in main function of main package
//
// # Custom analyzer noosexit
//
// The analyzer prohibits using os.Exit directly in the main function
// of the main package. This improves code testability and enables
// graceful shutdown.
//
// Example of correct structure:
//
//	package main
//
//	import "log"
//
//	func main() {
//	    if err := run(); err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
//	func run() error {
//	    if someError {
//	        return fmt.Errorf("error occurred")
//	    }
//	    return nil
//	}
//
// # Usage Examples
//
// Check entire project:
//
//	./staticlint ./...
//
// Check specific package:
//
//	./staticlint ./internal/handler
//
// With verbose output:
//
//	./staticlint -v ./...
package main

import (
	"go/ast"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unusedresult"

	"honnef.co/go/tools/staticcheck"
)

var noOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "prohibits using os.Exit in main function of main package",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		filePath := pass.Fset.File(file.Pos()).Name()
		if strings.Contains(filePath, "/go-build") ||
			strings.Contains(filePath, os.TempDir()) {
			continue
		}
		if strings.HasSuffix(filePath, ".test") ||
			strings.Contains(filePath, "_test") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			if fn.Name.Name == "main" {
				checkFunctionBody(fn, pass)
			}
			return true
		})
	}

	return nil, nil
}

func checkFunctionBody(fn *ast.FuncDecl, pass *analysis.Pass) {
	if fn.Body == nil {
		return
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		checkIfOsExit(callExpr, pass)
		return true
	})
}

func checkIfOsExit(call *ast.CallExpr, pass *analysis.Pass) {
	selExpr, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	xIdent, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return
	}

	if xIdent.Name == "os" && selExpr.Sel.Name == "Exit" {
		pass.Reportf(call.Pos(),
			"direct os.Exit call in main function is prohibited. "+
				"Use log.Fatal or return error from separate function.")
	}
}

func main() {
	checks := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		unusedresult.Analyzer,
		noOsExitAnalyzer,
	}

	saChecks := []string{
		"SA1000",
		"SA1019",
		"SA2000",
		"SA4000",
		"SA5000",
		"SA6000",
	}

	for _, analyzer := range staticcheck.Analyzers {
		for _, saCheck := range saChecks {
			if analyzer.Analyzer.Name == saCheck {
				checks = append(checks, analyzer.Analyzer)
			}
		}

		if analyzer.Analyzer.Name == "ST1003" ||
			analyzer.Analyzer.Name == "QF1001" {
			checks = append(checks, analyzer.Analyzer)
		}
	}

	println("Running multichecker with", len(checks), "analyzers")
	multichecker.Main(checks...)
}
