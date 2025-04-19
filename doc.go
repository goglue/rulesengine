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

A [RuleResult] describes the result of a rule evaluation, the structure contains
the final result in boolean type, the mismatching value (if any), the time taken
to evaluate the rule.
*/

package rulesengine
