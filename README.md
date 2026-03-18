# rulesengine

[![Go Reference](https://pkg.go.dev/badge/github.com/goglue/rulesengine.svg)](https://pkg.go.dev/github.com/goglue/rulesengine)
![Build Status](https://github.com/goglue/rulesengine/actions/workflows/pulls-pipeline.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-74.2%25-brightgreen)

A Go library for declarative, structured rule evaluation against arbitrary data maps — supporting nested logic, array iteration, date arithmetic, regex matching, and extensible custom functions.

---

## Table of Contents

1. [Overview](#overview)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Core Concepts](#core-concepts)
   - [The Rule Struct](#the-rule-struct)
   - [The RuleResult Struct](#the-rulesresult-struct)
   - [Evaluation Model](#evaluation-model)
5. [Field Paths](#field-paths)
6. [Operators Reference](#operators-reference)
   - [Logical](#logical)
   - [Equality](#equality)
   - [Numeric](#numeric)
   - [Membership](#membership)
   - [String](#string)
   - [Length](#length)
   - [Boolean](#boolean)
   - [Date / Time](#date--time)
   - [Array Iteration](#array-iteration-operators)
   - [Existence / Null](#existence--null)
   - [Type Checks](#type-checks)
   - [Custom Functions](#custom-functions-operator)
7. [Relative Time Expressions](#relative-time-expressions)
8. [Duration Strings](#duration-strings)
9. [Array Iteration (ANY / ALL / NONE)](#array-iteration-any--all--none)
10. [Custom Functions](#custom-functions)
11. [Options](#options)
12. [JSON Serialization](#json-serialization)
13. [Error Handling](#error-handling)
14. [Performance](#performance)
15. [License](#license)

---

## Overview

`rulesengine` lets you encode conditional business logic as data structures rather than code. Rules are composed from a fixed set of operators into arbitrarily deep trees, then evaluated against a `map[string]any` at runtime.

Key features:

- **Declarative** — rules are plain Go structs or JSON; no DSL parser, no reflection magic
- **JSON-serializable** — rules round-trip through `encoding/json` with no loss
- **Zero external dependencies** — only the Go standard library
- **Nested evaluation** — logical operators (`AND`, `OR`, `NOT`, `IF_THEN`) compose any tree depth
- **Array iteration** — `ANY`, `ALL`, `NONE` evaluate a predicate rule against each element of a slice field
- **Date arithmetic** — relative time expressions (`now-12mo`, `thisYear`, `today+7d`) as rule values
- **Custom functions** — register arbitrary Go functions and call them from rules
- **Timing and logging** — optional per-evaluation instrumentation via `Options`

---

## Installation

```bash
go get github.com/goglue/rulesengine
```

Requires Go 1.21 or later.

---

## Quick Start

The following example models a simplified loan eligibility decision: the requested amount must be within range, the company must be at least 2 years old, and either a Crefo score must be present or the applicant must be flagged as pre-approved.

```go
package main

import (
	"fmt"
	"time"

	"github.com/goglue/rulesengine"
)

func main() {
	rule := rulesengine.Rule{
		Operator: rulesengine.And,
		Children: []rulesengine.Rule{
			// Loan amount between 5,000 and 250,000
			{
				Operator: rulesengine.Between,
				Field:    "loan.amount",
				Value:    []any{5000, 250000},
			},
			// Company founded at least 2 years ago
			{
				Operator: rulesengine.Before,
				Field:    "company.foundedAt",
				Value:    "now-2y",
			},
			// Either has a Crefo score or is pre-approved
			{
				Operator: rulesengine.Or,
				Children: []rulesengine.Rule{
					{
						Operator: rulesengine.Exists,
						Field:    "applicant.crefoScore",
					},
					{
						Operator: rulesengine.IsTrue,
						Field:    "applicant.preApproved",
					},
				},
			},
		},
	}

	data := map[string]any{
		"loan": map[string]any{
			"amount": 75000,
		},
		"company": map[string]any{
			"foundedAt": time.Now().AddDate(-5, 0, 0),
		},
		"applicant": map[string]any{
			"crefoScore":  650,
			"preApproved": false,
		},
	}

	opts := rulesengine.DefaultOptions().WithTiming()
	result := rulesengine.Evaluate(rule, data, opts)

	fmt.Println("Eligible:", result.Result)           // true
	fmt.Println("Time taken:", result.TimeTaken)
}
```

---

## Core Concepts

### The Rule Struct

```go
type Rule struct {
    Operator Operator `json:"operator"`
    Field    string   `json:"field,omitempty"`
    Value    any      `json:"value,omitempty"`
    Children []Rule   `json:"children,omitempty"`
}
```

| Field      | Purpose                                                                                                  |
|------------|----------------------------------------------------------------------------------------------------------|
| `Operator` | The operation to perform. Always required.                                                               |
| `Field`    | Dot-notation path into the data map. Required for leaf operators; omitted for logical operators.         |
| `Value`    | The expected value to compare against. Type depends on the operator — see the operator reference below.  |
| `Children` | Sub-rules for logical operators (`AND`, `OR`, `NOT`, `IF_THEN`). Also used implicitly by array operators.|

A rule is either a **leaf** (has `Field` and `Value`, no `Children`) or a **composite** (has `Children`, no `Field`/`Value`). Array iteration operators (`ANY`, `ALL`, `NONE`) are a hybrid: `Field` names the slice, and `Value` holds a nested `Rule` as the predicate.

### The RuleResult Struct

```go
type RuleResult struct {
    Rule      Rule          `json:"rule"`
    Result    bool          `json:"result"`
    IsEmpty   bool          `json:"IsEmpty,omitempty"`
    Children  []RuleResult  `json:"children,omitempty"`
    Input     any           `json:"input,omitempty"`
    TimeTaken time.Duration `json:"timeTaken,omitempty"`
    Error     error         `json:"error,omitempty"`
}
```

| Field       | Description                                                                                                               |
|-------------|---------------------------------------------------------------------------------------------------------------------------|
| `Rule`      | The rule that produced this result — useful for debugging tree evaluations.                                               |
| `Result`    | The boolean outcome of the evaluation.                                                                                    |
| `IsEmpty`   | `true` when the field resolved to `nil` (missing key or explicit nil). Operators that require a value return `false` here.|
| `Children`  | Results for each child rule. Mirrors the tree structure of the input `Rule`.                                              |
| `Input`     | The resolved field value at time of evaluation.                                                                           |
| `TimeTaken` | Populated only when `WithTiming()` is active. Duration of this node's evaluation including all descendants.               |
| `Error`     | Non-nil when evaluation failed due to a type mismatch or invalid input. Underlying type is `rulesengine.Error`.           |

Reading the error:

```go
result := rulesengine.Evaluate(rule, data, rulesengine.DefaultOptions())
if result.Error != nil {
    if re, ok := result.Error.(rulesengine.Error); ok {
        fmt.Println("evaluation error:", re.Message, "value:", re.Value)
    }
}
```

### Evaluation Model

`Evaluate` traverses the rule tree depth-first. Composite operators evaluate their children and combine results:

- `AND` — iterates all children; returns `false` as soon as one child is `false` (short-circuits)
- `OR` — returns `true` as soon as one child is `true` (short-circuits)
- `NOT` — evaluates children with OR logic, then negates the result
- `IF_THEN` — requires exactly 2 children; equivalent to `¬A ∨ B`

Leaf operators resolve `Field` via dot-notation against `data`, then compare the resolved value against `Value` using operator-specific logic.

Errors on individual nodes do not halt evaluation of sibling nodes. A node that errors always returns `Result: false` with `Error` set.

---

## Field Paths

Fields use dot-notation to traverse nested maps:

```go
data := map[string]any{
    "user": map[string]any{
        "address": map[string]any{
            "city": "Berlin",
        },
    },
}

rule := rulesengine.Rule{
    Operator: rulesengine.Eq,
    Field:    "user.address.city",
    Value:    "Berlin",
}
// resolves data["user"]["address"]["city"] → "Berlin"
```

If any intermediate key is missing, the field resolves to `nil` and `RuleResult.IsEmpty` is set to `true`. Arrays accessed via iteration operators (`ANY`, `ALL`, `NONE`) use the `Field` path to locate the slice; predicate fields then resolve relative to each element.

---

## Operators Reference

### Logical

Logical operators do not use `Field` or `Value`. They combine `Children`.

#### AND

All children must evaluate to `true`.

```go
rulesengine.Rule{
    Operator: rulesengine.And,
    Children: []rulesengine.Rule{
        {Operator: rulesengine.Gte, Field: "score", Value: 600},
        {Operator: rulesengine.IsTrue, Field: "verified"},
    },
}
```

#### OR

At least one child must evaluate to `true`.

```go
rulesengine.Rule{
    Operator: rulesengine.Or,
    Children: []rulesengine.Rule{
        {Operator: rulesengine.Eq, Field: "tier", Value: "premium"},
        {Operator: rulesengine.Gte, Field: "accountAgeDays", Value: 365},
    },
}
```

#### NOT

Negates the OR-combined result of its children.

```go
rulesengine.Rule{
    Operator: rulesengine.Not,
    Children: []rulesengine.Rule{
        {Operator: rulesengine.In, Field: "country", Value: []any{"US", "CN", "RU"}},
    },
}
```

#### IF_THEN

Material implication: `¬A ∨ B`. If the first child is false, the rule is trivially true. If the first child is true, the second child must also be true. Requires exactly 2 children.

```go
// IF loan.secured THEN collateral.value >= loan.amount  (modeled with separate field checks)
rulesengine.Rule{
    Operator: rulesengine.IfThen,
    Children: []rulesengine.Rule{
        {Operator: rulesengine.IsTrue, Field: "loan.secured"},
        {Operator: rulesengine.Gte, Field: "collateral.value", Value: 50000},
    },
}
```

---

### Equality

**Value type:** any comparable scalar (`string`, `int`, `float64`, `bool`, etc.)

#### EQ

```go
{Operator: rulesengine.Eq, Field: "status", Value: "active"}
```

#### NEQ

```go
{Operator: rulesengine.Neq, Field: "status", Value: "blocked"}
```

---

### Numeric

Accepts all integer and float types, as well as numeric strings. **Value type:** number or numeric string.

#### GT / GTE / LT / LTE

```go
{Operator: rulesengine.Gt,  Field: "revenue", Value: 100000}
{Operator: rulesengine.Gte, Field: "revenue", Value: 100000}
{Operator: rulesengine.Lt,  Field: "riskScore", Value: 0.75}
{Operator: rulesengine.Lte, Field: "riskScore", Value: 0.75}
```

#### BETWEEN

Inclusive range check. **Value type:** `[]any{min, max}`

```go
{
    Operator: rulesengine.Between,
    Field:    "loan.amount",
    Value:    []any{10000, 500000},
}
```

---

### Membership

#### IN

Field value must be one of the listed values. **Value type:** `[]any`

```go
{Operator: rulesengine.In, Field: "legalForm", Value: []any{"GmbH", "AG", "UG"}}
```

#### NOT_IN

```go
{Operator: rulesengine.NotIn, Field: "riskCategory", Value: []any{"high", "critical"}}
```

#### ANY_IN

The field must be a slice. Returns `true` if any element of the slice is present in the value list. **Value type:** `[]any`

```go
// applicant.roles is []string{"analyst", "manager"}
// true if any of those roles is in the allowed list
{
    Operator: rulesengine.AnyIn,
    Field:    "applicant.roles",
    Value:    []any{"admin", "manager", "owner"},
}
```

---

### String

Field and Value must be strings.

#### CONTAINS / NOT_CONTAINS

```go
{Operator: rulesengine.Contains,    Field: "company.name", Value: "GmbH"}
{Operator: rulesengine.NotContains, Field: "email",        Value: "spam"}
```

#### STARTS_WITH / ENDS_WITH

```go
{Operator: rulesengine.StartsWith, Field: "iban", Value: "DE"}
{Operator: rulesengine.EndsWith,   Field: "email", Value: ".de"}
```

#### MATCHES

**Value type:** regex string. Returns `true` if the field value matches the regular expression.

```go
{Operator: rulesengine.Matches, Field: "taxId", Value: `^DE\d{9}$`}
```

---

### Length

Applies to strings (character count) and slices (element count). **Value type:** integer.

#### LENGTH_EQ / LENGTH_GT / LENGTH_LT

```go
{Operator: rulesengine.LengthEq, Field: "iban",       Value: 22}
{Operator: rulesengine.LengthGt, Field: "documents",  Value: 0}
{Operator: rulesengine.LengthLt, Field: "companyName", Value: 100}
```

---

### Boolean

#### IS_TRUE / IS_FALSE

```go
{Operator: rulesengine.IsTrue,  Field: "applicant.kycPassed"}
{Operator: rulesengine.IsFalse, Field: "applicant.sanctionsHit"}
```

---

### Date / Time

**Field value types accepted:** `time.Time`, `*time.Time`, RFC3339 strings (`"2023-01-15T00:00:00Z"`), date-only strings (`"2023-01-15"`).

**Value field types accepted:** `time.Time`, `*time.Time`, relative time expression strings (see [Relative Time Expressions](#relative-time-expressions)), `int` (for `YEAR_EQ` and `MONTH_EQ`).

#### BEFORE / AFTER

```go
// Founded before 2 years ago (company is at least 2 years old)
{Operator: rulesengine.Before, Field: "company.foundedAt", Value: "now-2y"}

// Contract signed after the start of this year
{Operator: rulesengine.After, Field: "contract.signedAt", Value: "thisYear"}
```

#### DATE_BETWEEN

**Value type:** `[]any{start, end}` or `[]time.Time{start, end}`. Both bounds are inclusive.

```go
{
    Operator: rulesengine.DateBetween,
    Field:    "invoice.date",
    Value:    []any{"2024-01-01", "2024-12-31"},
}
```

#### WITHIN_LAST / WITHIN_NEXT

Checks whether the field's time falls within the last or next N units from now. **Value type:** duration string (see [Duration Strings](#duration-strings)).

```go
// Document uploaded within the last 30 days
{Operator: rulesengine.WithinLast, Field: "document.uploadedAt", Value: "30d"}

// Subscription renews within the next 2 weeks
{Operator: rulesengine.WithinNext, Field: "subscription.renewsAt", Value: "2w"}
```

#### YEAR_EQ

**Value type:** `int` (absolute year) or relative time expression string.

```go
{Operator: rulesengine.YearEq, Field: "contract.signedAt", Value: 2024}
{Operator: rulesengine.YearEq, Field: "contract.signedAt", Value: "thisYear"}
```

#### MONTH_EQ

**Value type:** `int` (1–12) or relative time expression string.

```go
{Operator: rulesengine.MonthEq, Field: "payment.dueDate", Value: 12}
{Operator: rulesengine.MonthEq, Field: "payment.dueDate", Value: "thisMonth"}
```

---

### Array Iteration Operators

See also [Array Iteration (ANY / ALL / NONE)](#array-iteration-any--all--none) for detailed examples.

#### ANY / ALL / NONE

**Field:** path to the slice in the data map.
**Value:** a nested `Rule` used as the predicate, evaluated against each element.

```go
// ANY document has DocumentTypeID == 3
{
    Operator: rulesengine.Any,
    Field:    "applicant.documents",
    Value: rulesengine.Rule{
        Operator: rulesengine.Eq,
        Field:    "DocumentTypeID",
        Value:    3,
    },
}
```

---

### Existence / Null

No `Value` required.

#### EXISTS / IS_NOT_NULL

Returns `true` if the field is present in the data map and its value is not `nil`. `Exists` and `IsNotNull` are aliases.

```go
{Operator: rulesengine.Exists,    Field: "applicant.crefoScore"}
{Operator: rulesengine.IsNotNull, Field: "applicant.crefoScore"}
```

#### NOT_EXISTS / IS_NULL

Returns `true` if the field is absent or `nil`. `NotExists` and `IsNull` are aliases.

```go
{Operator: rulesengine.NotExists, Field: "applicant.bankruptcyDate"}
{Operator: rulesengine.IsNull,    Field: "applicant.bankruptcyDate"}
```

---

### Type Checks

No `Value` required.

```go
{Operator: rulesengine.IsNumber, Field: "score"}      // any int or float type
{Operator: rulesengine.IsString, Field: "name"}
{Operator: rulesengine.IsBool,   Field: "active"}
{Operator: rulesengine.IsDate,   Field: "createdAt"}  // time.Time only
{Operator: rulesengine.IsList,   Field: "tags"}       // any slice
{Operator: rulesengine.IsObject, Field: "address"}    // struct or map
```

---

### Custom Functions Operator

Calls a registered custom function by name. **Value type:** `[]any{"funcName", arg1, arg2, ...}`

```go
{
    Operator: rulesengine.Custom,
    Value:    []any{"isEligibleForProduct", "product-42", "applicant-99"},
}
```

See [Custom Functions](#custom-functions) for how to register functions.

---

## Relative Time Expressions

Relative time expressions can be used as the `Value` in any date/time operator. They are evaluated at rule evaluation time (i.e., against the current clock).

### Base Tokens

| Token       | Meaning                            |
|-------------|------------------------------------|
| `now`       | Current timestamp (with time)      |
| `today`     | Start of the current day (00:00:00)|
| `thisday`   | Alias for `today`                  |
| `thisMonth` | First day of the current month     |
| `thisYear`  | First day of the current year      |

### Arithmetic

Append `+` or `-` followed by a quantity and unit to offset the base token:

```
now-12mo        // 12 months ago
now+1y          // 1 year from now
thisYear-2y     // start of the year, 2 years ago
today+7d        // 7 days from today
thisMonth+1mo   // start of next month
```

### Supported Units

| Unit suffix(es)          | Meaning       |
|--------------------------|---------------|
| `y`, `yr`, `years`       | Years         |
| `mo`, `month`, `months`  | Months        |
| `w`, `week`, `weeks`     | Weeks         |
| `d`, `day`, `days`       | Days          |
| `h`, `hr`, `hours`       | Hours         |
| `m`, `min`, `minutes`    | Minutes       |
| `s`, `sec`, `seconds`    | Seconds       |
| `ms`                     | Milliseconds  |
| `us`, `µs`               | Microseconds  |
| `ns`                     | Nanoseconds   |

### Examples

```go
// Company incorporated at least 3 years ago
{Operator: rulesengine.Before, Field: "company.incorporatedAt", Value: "now-3y"}

// Contract expires after today
{Operator: rulesengine.After, Field: "contract.expiresAt", Value: "today"}

// KYC completed this year
{Operator: rulesengine.YearEq, Field: "kyc.completedAt", Value: "thisYear"}

// Invoice dated within the current month
{Operator: rulesengine.After, Field: "invoice.date", Value: "thisMonth"}
```

---

## Duration Strings

`WithinLast` and `WithinNext` use a simpler duration format — a numeric value followed by a unit abbreviation. Decimal values are supported.

| Format   | Meaning         |
|----------|-----------------|
| `"30s"`  | 30 seconds      |
| `"5h"`   | 5 hours         |
| `"2d"`   | 2 days          |
| `"3w"`   | 3 weeks         |
| `"1mo"`  | 1 month         |
| `"1.5y"` | 18 months       |

```go
// Field value must be within the last 90 days
{Operator: rulesengine.WithinLast, Field: "lastLogin", Value: "90d"}

// Appointment is within the next 3 months
{Operator: rulesengine.WithinNext, Field: "appointment.scheduledAt", Value: "3mo"}
```

---

## Array Iteration (ANY / ALL / NONE)

The `ANY`, `ALL`, and `NONE` operators iterate over a slice field and evaluate a predicate rule against each element.

- `ANY` — returns `true` if at least one element satisfies the predicate
- `ALL` — returns `true` if every element satisfies the predicate
- `NONE` — returns `true` if no element satisfies the predicate

### With Object Elements

When the slice contains maps (`[]map[string]any` or `[]any` of maps), the predicate's `Field` paths resolve against each element's own keys:

```go
data := map[string]any{
    "applicant": map[string]any{
        "documents": []any{
            map[string]any{"DocumentTypeID": 1, "verified": true},
            map[string]any{"DocumentTypeID": 3, "verified": true},
            map[string]any{"DocumentTypeID": 5, "verified": false},
        },
    },
}

// True: at least one document has DocumentTypeID == 3
rule := rulesengine.Rule{
    Operator: rulesengine.Any,
    Field:    "applicant.documents",
    Value: rulesengine.Rule{
        Operator: rulesengine.Eq,
        Field:    "DocumentTypeID",
        Value:    3,
    },
}

// True: all documents have a DocumentTypeID present
allHaveID := rulesengine.Rule{
    Operator: rulesengine.All,
    Field:    "applicant.documents",
    Value: rulesengine.Rule{
        Operator: rulesengine.Exists,
        Field:    "DocumentTypeID",
    },
}
```

Compose predicates using logical operators for multi-condition element checks:

```go
// Any document where DocumentTypeID == 3 AND verified == true
rule := rulesengine.Rule{
    Operator: rulesengine.Any,
    Field:    "applicant.documents",
    Value: rulesengine.Rule{
        Operator: rulesengine.And,
        Children: []rulesengine.Rule{
            {Operator: rulesengine.Eq,     Field: "DocumentTypeID", Value: 3},
            {Operator: rulesengine.IsTrue, Field: "verified"},
        },
    },
}
```

### With Primitive Elements

When the slice contains primitive values (strings, numbers), set the predicate's `Field` to an empty string `""`. The element itself is passed as the value to compare against:

```go
data := map[string]any{
    "applicant": map[string]any{
        "tags": []any{"verified", "premium", "de-resident"},
    },
}

// True if any tag equals "premium"
rule := rulesengine.Rule{
    Operator: rulesengine.Any,
    Field:    "applicant.tags",
    Value: rulesengine.Rule{
        Operator: rulesengine.Eq,
        Field:    "",
        Value:    "premium",
    },
}

// True if no tag equals "blocked"
noneBlocked := rulesengine.Rule{
    Operator: rulesengine.None,
    Field:    "applicant.tags",
    Value: rulesengine.Rule{
        Operator: rulesengine.Eq,
        Field:    "",
        Value:    "blocked",
    },
}
```

---

## Custom Functions

Register arbitrary Go functions and invoke them from rules using the `CUSTOM_FUNC` operator.

### Signature

```go
type CustomFunc func(args ...any) (bool, error)
```

### Registering a Function

```go
rulesengine.RegisterFunc("hasSufficientCredit", func(args ...any) (bool, error) {
    if len(args) < 2 {
        return false, fmt.Errorf("hasSufficientCredit requires 2 arguments")
    }
    applicantID, ok1 := args[0].(string)
    threshold, ok2  := args[1].(float64)
    if !ok1 || !ok2 {
        return false, fmt.Errorf("invalid argument types")
    }
    // call external service or perform computation
    score := fetchCreditScore(applicantID)
    return score >= threshold, nil
})
```

### Calling from a Rule

The `Value` field is `[]any` where the first element is the registered function name and subsequent elements are positional arguments:

```go
rule := rulesengine.Rule{
    Operator: rulesengine.Custom,
    Value:    []any{"hasSufficientCredit", "applicant-123", 650.0},
}
```

### Looking Up a Function

```go
fn, ok := rulesengine.GetFunc("hasSufficientCredit")
if ok {
    result, err := fn("applicant-123", 650.0)
}
```

Custom functions are registered globally and are safe to register at startup (e.g., in `init()`). Registration is not concurrency-safe during runtime evaluation — register all functions before starting concurrent evaluation.

---

## Options

`Options` is constructed via a builder pattern. Start with `DefaultOptions()` and chain modifiers.

```go
opts := rulesengine.DefaultOptions().
    WithTiming().
    WithLogger(func(format string, args ...any) {
        log.Printf("[rulesengine] "+format, args...)
    })

result := rulesengine.Evaluate(rule, data, opts)
```

### WithTiming

Enables per-node timing. When active, `RuleResult.TimeTaken` is populated with the wall-clock duration of each node's evaluation, including all its descendants. Use this for performance profiling of complex rule trees.

```go
opts := rulesengine.DefaultOptions().WithTiming()
result := rulesengine.Evaluate(rule, data, opts)
fmt.Println("root evaluation took:", result.TimeTaken)
```

### WithLogger

Accepts a `LoggerFunc` with the signature `func(format string, args ...any)`. Compatible with `log.Printf`, `zap.SugaredLogger.Infof`, or any similar function. The logger receives diagnostic messages during evaluation.

```go
type LoggerFunc func(format string, args ...any)

opts := rulesengine.DefaultOptions().WithLogger(log.Printf)
```

---

## JSON Serialization

`Rule` is fully JSON-serializable using standard `encoding/json`. Rules can be stored in a database, transmitted over a network, or loaded from configuration files and evaluated at runtime.

### Serializing a Rule

```go
import "encoding/json"

rule := rulesengine.Rule{
    Operator: rulesengine.And,
    Children: []rulesengine.Rule{
        {Operator: rulesengine.Gte, Field: "score", Value: 600},
        {Operator: rulesengine.Eq,  Field: "status", Value: "active"},
    },
}

data, err := json.Marshal(rule)
// {"operator":"AND","children":[{"operator":"GTE","field":"score","value":600},{"operator":"EQ","field":"status","value":"active"}]}
```

### Deserializing a Rule

```go
var rule rulesengine.Rule
err := json.Unmarshal(data, &rule)
if err != nil {
    // handle
}
result := rulesengine.Evaluate(rule, inputData, rulesengine.DefaultOptions())
```

### Nested Rule Values (ANY / ALL / NONE)

When `Value` is a nested `Rule` (as used by `ANY`, `ALL`, `NONE`), it serializes and deserializes correctly because `Value` is typed as `any`. The JSON representation encodes the inner rule object inline, and on deserialization it is unmarshaled back into a `map[string]any` which the evaluation engine handles transparently.

```json
{
  "operator": "ANY",
  "field": "applicant.documents",
  "value": {
    "operator": "EQ",
    "field": "DocumentTypeID",
    "value": 3
  }
}
```

---

## Error Handling

### When Errors Occur

Errors are attached to the `RuleResult.Error` field when:

- The resolved field value is the wrong type for the operator (e.g., `GT` applied to a string)
- A date/time string cannot be parsed
- A regex pattern is invalid
- A custom function returns an error
- An operator receives malformed `Value` input (e.g., `Between` with fewer than 2 elements)

### Error Type

The underlying type is `rulesengine.Error`:

```go
type Error struct {
    Message string `json:"message"`
    Value   any    `json:"value"`
}
```

To inspect:

```go
result := rulesengine.Evaluate(rule, data, rulesengine.DefaultOptions())
if result.Error != nil {
    if re, ok := result.Error.(rulesengine.Error); ok {
        fmt.Printf("message: %s, value: %v\n", re.Message, re.Value)
    }
}
```

### Error Isolation

Errors do not propagate up or halt sibling evaluation. A node that errors returns `Result: false` with `Error` set. The parent composite operator receives `false` from that child and continues evaluating other children normally.

To scan the full tree for errors:

```go
func collectErrors(r rulesengine.RuleResult) []error {
    var errs []error
    if r.Error != nil {
        errs = append(errs, r.Error)
    }
    for _, child := range r.Children {
        errs = append(errs, collectErrors(child)...)
    }
    return errs
}
```

### IsEmpty

`RuleResult.IsEmpty` is set to `true` when the field resolved to `nil` — either because the key is absent from the data map or its value was explicitly `nil`. Most leaf operators return `false` when the input is empty. Check `IsEmpty` to distinguish "field was missing" from "field was present but the comparison returned false".

---

## Performance

The library is benchmarked using standard Go benchmarks (`go test -bench=.`). For production use, observe the following:

- **Build rules once, evaluate many times.** Rule struct construction has no internal caching; the tree is fully re-evaluated on each call to `Evaluate`. Parse and assemble rules at startup or when configuration changes, then reuse the same `Rule` value across goroutines (it is read-only during evaluation).

- **Reuse `Options`.** `DefaultOptions()` is lightweight, but if you chain `WithTiming()` or `WithLogger()`, construct the `Options` value once and share it.

- **Avoid unnecessary depth.** Each composite operator adds a recursive call. Flatten rules where semantically equivalent — a single `AND` with N children is more efficient than N nested `AND` nodes each wrapping one child.

- **Use `WithTiming()` only during profiling.** The timing path records `time.Now()` on every node entry and exit. Disable it in production unless you actively consume the timing data.

- **Field path resolution** is O(depth) per field access. For very deep paths with many array iterations, cache frequently accessed intermediate data in shallower keys if the path is hot.

---

## License

This project is licensed under the MIT License. See the [LICENSE](https://github.com/goglue/rulesengine/blob/main/LICENSE) file for details.