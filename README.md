# 🧠 RulesEngine

![Build Status](https://github.com/cubeox/lighthouse/actions/workflows/pulls-pipeline.yml/badge.svg)
![Coverage](https://img.shields.io/badge/Coverage-81.7%25-brightgreen)

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

| Operator    | Description      |
|-------------|------------------|
| And         | Logical          |
| Or          | Logical          |
| Not         | Logical          |
| Eq          | Equality         |
| Neq         | Equality         |
| Gt          | Numeric          |
| Gte         | Numeric          |
| Lt          | Numeric          |
| Lte         | Numeric          |
| Between     | Numeric          |
| In          | Numeric          |
| NotIn       | Numeric          |
| Contains    | String           |
| NotContains | String           |
| StartsWith  | String           |
| EndsWith    | String           |
| Matches     | String           |
| LengthEq    | String           |    
| LengthGt    | String           |    
| LengthLt    | String           |
| IsTrue      | Boolean          |
| IsFalse     | Boolean          |
| Before      | Date / Time      |
| After       | Date / Time      |
| DateBetween | Date / Time      |
| WithinLast  | Date / Time      |
| WithinNext  | Date / Time      |
| Any         | Array            |
| All         | Array            |
| None        | Array            |
| Exists      | Existence / Null |
| NotExists   | Existence / Null |
| IsNull      | Existence / Null |
| IsNotNull   | Existence / Null |
| IsNumber    | Type             |
| IsString    | Type             |
| IsBool      | Type             |
| IsDate      | Type             |
| IsList      | Type             |
| IsObject    | Type             |
| Custom      | Custom           |

## 🧪 Usage Examples

```go
import (
    "fmt"

    "github.com/goglue/rulesengine"
    "github.com/goglue/rulesengine/rules"
)

func main() {
    rule := rules.Node{
        Operator: rules.And,
        Children: []rules.Node{
            {
                Field:    "age",
                Operator: rules.Gte,
                Value:    21,
            },
            {
                Field:    "country",
                Operator: rules.Eq,
                Value:    "DE",
            },
        },
    }

    data := map[string]interface{}{
        "age":     25,
        "country": "US",
    }

    result := rulesengine.Evaluate(rule, data, rulesengine.Options{})
    fmt.Println("Result:", result.Result) // true
}
```
## 📄 License
This project is licensed under the MIT License. See the [LICENSE](https://github.com/goglue/rulesengine/blob/main/LICENSE) file for details.