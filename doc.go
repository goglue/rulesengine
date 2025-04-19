// Copyright 2025 Moath Almallahi. All rights reserved.
// Use of this source code is governed by an MIT
// license that can be found in the LICENSE file.

/*

Package rulesengine defines the different types needed for the rule definition
and the evaluation results, as well as the execution method which triggers the
evaluation.

# Rule

The primary type in the API is [Rule]. A Rule describes a single or a
group of rules:

	package testpackage

	import (
		"fmt"

		"github.com/goglue/rulesengine"
		"github.com/goglue/rulesengine/rules"
	)

	func run() {
		rule := rulesengine.Rule{
			Operator: rulesengine.And,
			Children: []rulesengine.Rule{
				{
					Operator: rulesengine.And,
					Children: []rulesengine.Rule{
						{
							Field: "user.name",
							Operator: rulesengine.LengthGt,
							Value: 2,
						},
						{
							Field: "user.name",
							Operator: rulesengine.LengthLt,
							Value: 25,
						},
					},
				},
				{
					Field:    "user.age",
					Operator: rulesengine.Gte,
					Value:    21,
				},
				{
					Field:    "user.country",
					Operator: rulesengine.Eq,
					Value:    "DE",
				},
			},
		}

		data := map[string]interface{}{
			"user": map[string]interface{}{
				"name":    "Test",
				"age":     25,
				"country": "DE",
			},
		}

		result := rulesengine.Evaluate(rule, data, rulesengine.DefaultOptions())
		fmt.Println("Result:", result.Result) // true
	}

# RuleResult

A [RuleResult] describes a single unit of work: the application of a particular
Analyzer to a particular package of Go code.
The Pass provides information to the Analyzer's Run function about the
package being analyzed, and provides operations to the Run function for
reporting diagnostics and other information back to the driver.

	type Pass struct {
		Fset         *token.FileSet
		Files        []*ast.File
		OtherFiles   []string
		IgnoredFiles []string
		Pkg          *types.Package
		TypesInfo    *types.Info
		ResultOf     map[*Analyzer]interface{}
		Report       func(Diagnostic)
		...
	}

The Fset, Files, Pkg, and TypesInfo fields provide the syntax trees,
type information, and source positions for a single package of Go code.

The OtherFiles field provides the names of non-Go
files such as assembly that are part of this package.
Similarly, the IgnoredFiles field provides the names of Go and non-Go
source files that are not part of this package with the current build
configuration but may be part of other build configurations.
The contents of these files may be read using Pass.ReadFile;
see the "asmdecl" or "buildtags" analyzers for examples of loading
non-Go files and reporting diagnostics against them.

The ResultOf field provides the results computed by the analyzers
required by this one, as expressed in its Analyzer.Requires field. The
driver runs the required analyzers first and makes their results
available in this map. Each Analyzer must return a value of the type
described in its Analyzer.ResultType field.
For example, the "ctrlflow" analyzer returns a *ctrlflow.CFGs, which
provides a control-flow graph for each function in the package (see
golang.org/x/tools/go/cfg); the "inspect" analyzer returns a value that
enables other Analyzers to traverse the syntax trees of the package more
efficiently; and the "buildssa" analyzer constructs an SSA-form
intermediate representation.
Each of these Analyzers extends the capabilities of later Analyzers
without adding a dependency to the core API, so an analysis tool pays
only for the extensions it needs.

The Report function emits a diagnostic, a message associated with a
source position. For most analyses, diagnostics are their primary
result.
For convenience, Pass provides a helper method, Reportf, to report a new
diagnostic by formatting a string.
Diagnostic is defined as:

	type Diagnostic struct {
		Pos      token.Pos
		Category string // optional
		Message  string
	}

The optional Category field is a short identifier that classifies the
kind of message when an analysis produces several kinds of diagnostic.

The [Diagnostic] struct does not have a field to indicate its severity
because opinions about the relative importance of Analyzers and their
diagnostics vary widely among users. The design of this framework does
not hold each Analyzer responsible for identifying the severity of its
diagnostics. Instead, we expect that drivers will allow the user to
customize the filtering and prioritization of diagnostics based on the
producing Analyzer and optional Category, according to the user's
preferences.

Most Analyzers inspect typed Go syntax trees, but a few, such as asmdecl
and buildtag, inspect the raw text of Go source files or even non-Go
files such as assembly. To report a diagnostic against a line of a
raw text file, use the following sequence:

	content, err := pass.ReadFile(filename)
	if err != nil { ... }
	tf := fset.AddFile(filename, -1, len(content))
	tf.SetLinesForContent(content)
	...
	pass.Reportf(tf.LineStart(line), "oops")

# Modular analysis with Facts

To improve efficiency and scalability, large programs are routinely
built using separate compilation: units of the program are compiled
separately, and recompiled only when one of their dependencies changes;
independent modules may be compiled in parallel. The same technique may
be applied to static analyses, for the same benefits. Such analyses are
described as "modular".

A compiler’s type checker is an example of a modular static analysis.
Many other checkers we would like to apply to Go programs can be
understood as alternative or non-standard type systems. For example,
vet's printf checker infers whether a function has the "printf wrapper"
type, and it applies stricter checks to calls of such functions. In
addition, it records which functions are printf wrappers for use by
later analysis passes to identify other printf wrappers by induction.
A result such as “f is a printf wrapper” that is not interesting by
itself but serves as a stepping stone to an interesting result (such as
a diagnostic) is called a [Fact].

The analysis API allows an analysis to define new types of facts, to
associate facts of these types with objects (named entities) declared
within the current package, or with the package as a whole, and to query
for an existing fact of a given type associated with an object or
package.

An Analyzer that uses facts must declare their types:

	var Analyzer = &analysis.Analyzer{
		Name:      "printf",
		FactTypes: []analysis.Fact{new(isWrapper)},
		...
	}

	type isWrapper struct{} // => *types.Func f “is a printf wrapper”

The driver program ensures that facts for a pass’s dependencies are
generated before analyzing the package and is responsible for propagating
facts from one package to another, possibly across address spaces.
Consequently, Facts must be serializable. The API requires that drivers
use the gob encoding, an efficient, robust, self-describing binary
protocol. A fact type may implement the GobEncoder/GobDecoder interfaces
if the default encoding is unsuitable. Facts should be stateless.
Because serialized facts may appear within build outputs, the gob encoding
of a fact must be deterministic, to avoid spurious cache misses in
build systems that use content-addressable caches.
The driver makes a single call to the gob encoder for all facts
exported by a given analysis pass, so that the topology of
shared data structures referenced by multiple facts is preserved.

The Pass type has functions to import and export facts,
associated either with an object or with a package:

	type Pass struct {
		...
		ExportObjectFact func(types.Object, Fact)
		ImportObjectFact func(types.Object, Fact) bool

		ExportPackageFact func(fact Fact)
		ImportPackageFact func(*types.Package, Fact) bool
	}

An Analyzer may only export facts associated with the current package or
its objects, though it may import facts from any package or object that
is an import dependency of the current package.

Conceptually, ExportObjectFact(obj, fact) inserts fact into a hidden map keyed by
the pair (obj, TypeOf(fact)), and the ImportObjectFact function
retrieves the entry from this map and copies its value into the variable
pointed to by fact. This scheme assumes that the concrete type of fact
is a pointer; this assumption is checked by the Validate function.
See the "printf" analyzer for an example of object facts in action.

Some driver implementations (such as those based on Bazel and Blaze) do
not currently apply analyzers to packages of the standard library.
Therefore, for best results, analyzer authors should not rely on
analysis facts being available for standard packages.
For example, although the printf checker is capable of deducing during
analysis of the log package that log.Printf is a printf wrapper,
this fact is built in to the analyzer so that it correctly checks
calls to log.Printf even when run in a driver that does not apply
it to standard packages. We would like to remove this limitation in future.

# Testing an Analyzer

The analysistest subpackage provides utilities for testing an Analyzer.
In a few lines of code, it is possible to run an analyzer on a package
of testdata files and check that it reported all the expected
diagnostics and facts (and no more). Expectations are expressed using
"// want ..." comments in the input code.

# Standalone commands

Analyzers are provided in the form of packages that a driver program is
expected to import. The vet command imports a set of several analyzers,
but users may wish to define their own analysis commands that perform
additional checks. To simplify the task of creating an analysis command,
either for a single analyzer or for a whole suite, we provide the
singlechecker and multichecker subpackages.

The singlechecker package provides the main function for a command that
runs one analyzer. By convention, each analyzer such as
go/analysis/passes/findcall should be accompanied by a singlechecker-based
command such as go/analysis/passes/findcall/cmd/findcall, defined in its
entirety as:

	package main

	import (
		"golang.org/x/tools/go/analysis/passes/findcall"
		"golang.org/x/tools/go/analysis/singlechecker"
	)

	func main() { singlechecker.Main(findcall.Analyzer) }

A tool that provides multiple analyzers can use multichecker in a
similar way, giving it the list of Analyzers.
*/

package rulesengine
