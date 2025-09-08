# 🧠 RulesEngine

[![Go Reference](https://pkg.go.dev/badge/github.com/goglue/rulesengine.svg)](https://pkg.go.dev/github.com/goglue/rulesengine)
![Build Status](https://github.com/goglue/rulesengine/actions/workflows/pulls-pipeline.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-64.9%25-yellow)

A flexible and extensible rules engine written in Go. Use it to define
conditional logic in structured, declarative form and evaluate it dynamically
against data — both flat and deeply nested.

---

## 📦 Installation

```bash
go get github.com/goglue/rulesengine
```

---

## ⚙️ Supported Operators

| Operator    | Data Type               | Description                                                                          |
|-------------|-------------------------|--------------------------------------------------------------------------------------|
| And         | Logical                 | Returns `true` if all child rules are `true`.                                        | 
| Or          | Logical                 | Returns `true` if any or all child rule is `true`, otherwise `false`                 |
| Not         | Logical                 | Returns the negation of the child rule's result.                                     |
| Eq          | Equality (Scalar types) | Checks if a field is equal to a given values.                                        |
| Neq         | Equality (Scalar types) | Checks if a field is not equal to a given values.                                    |
| Gt          | Numeric                 | Checks if a field is greater than the specified.                                     |
| Gte         | Numeric                 | Checks if a field is greater than or equal to the specified.                         |
| Lt          | Numeric                 | Checks if a field is less than the specified.                                        |
| Lte         | Numeric                 | Checks if a field is less than or equal to the specified.                            |
| Between     | Numeric                 | Checks if a field is between the specified range.                                    |
| In          | (Scalar types)          | Checks if field is in a specified list.                                              |
| NotIn       | (Scalar types)          | Checks if field is not in a specified list.                                          |
| AnyIn       | (Scalar types)          | Checks if any of field's values is in a specified list. Field should be array                                          |
| Contains    | String                  | Checks if a field contains a specified `string`.                                     |
| NotContains | String                  | Checks if a field does not contain a specified `string`.                             |
| StartsWith  | String                  | Checks if a field starts with a `string`.                                            |
| EndsWith    | String                  | Checks if a field ends with a `string`.                                              |
| Matches     | String                  | Checks if a field matches with a given regex.                                        |
| LengthEq    | String / Slice          | Checks if a field length is equal to a specified length.                             |
| LengthGt    | String / Slice          | Checks if a field length is greater than a specified length.                         |
| LengthLt    | String / Slice          | Checks if a field length is less than a specified length.                            |
| IsTrue      | Boolean                 | Checks if a field value is `true`.                                                   |
| IsFalse     | Boolean                 | Checks if a field value is `false`.                                                  |
| Before      | Date / Time             | Checks is a time is before a given time.                                             |
| After       | Date / Time             | Checks is a time is after a given time.                                              |
| DateBetween | Date / Time             | Checks is a time is between a given range.                                           |
| WithinLast  | Date / Time             | Checks is a time is within the last X of a given duration.                           |
| WithinNext  | Date / Time             | Checks is a time is within the next X a given duration.                              |
| Any         | Array                   | Returns `true` if any element statisfy the given rule.                               |
| All         | Array                   | Returns `true` if all elements satisfy the given rule.                               |
| None        | Array                   | Returns `true` if no elements satisfy the given rule.                                |
| Exists      | Existence / Null        | Returns `true` if an attribute is set. (Alias for `IsNotNull`)                       |
| NotExists   | Existence / Null        | Returns `true` if an attribute is not set. (Alias for `IsNull`)                      |
| IsNull      | Existence / Null        |                                                                                      |
| IsNotNull   | Existence / Null        |                                                                                      |
| IsNumber    | Type                    | Returns `true` if an attribute type is a number (any of the `int` or `float` types). |
| IsString    | Type                    | Returns `true` if an attribute type is a `string`.                                   |
| IsBool      | Type                    | Returns `true` if an attribute type is a `bool`.                                     |
| IsDate      | Type                    | Returns `true` if an attribute type is a `time.Time`.                                |
| IsList      | Type                    | Returns `true` if an attribute type is a `Slice`.                                    |
| IsObject    | Type                    | Returns `true` if an attribute type is a `struct` or `map`                           |
| Custom      | Custom                  | Returns `true` if the custom function defined returns true                           |

## 🧪 Usage Examples

```go
import (
    "fmt"

    "github.com/goglue/rulesengine"
)

func main() {
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

    data := map[string]any{
        "user": map[string]any{
            "name":    "Test",
            "age":     25,
            "country": "DE",
        },
    }

    result := rulesengine.Evaluate(rule, data, rulesengine.DefaultOptions())
    fmt.Println("Result:", result.Result) // true
}
```

For more examples, check the [test](https://github.com/goglue/rulesengine/blob/main/rulesengine_test.go) file.

## 📄 License

This project is licensed under the MIT License. See
the [LICENSE](https://github.com/goglue/rulesengine/blob/main/LICENSE) file for
details.