package rulesengine

import (
	"github.com/goglue/rulesengine/rules"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func Evaluate(
	node rules.Node, data map[string]interface{}, opts Options,
) rules.NodeEvaluation {
	var now time.Time
	if opts.Timing {
		now = time.Now()
	}
	evaluation := rules.NodeEvaluation{
		Node: rules.Node{
			Operator: node.Operator, Field: node.Field, Value: data,
		},
	}

	switch node.Operator {
	case rules.And:
		evaluation.Result = true
		for _, child := range node.Children {
			childEvaluation := Evaluate(child, data, opts)
			evaluation.Children = append(evaluation.Children, childEvaluation)
			evaluation.Result = childEvaluation.Result && evaluation.Result
		}

		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	case rules.Or:
		for _, child := range node.Children {
			childEvaluation := Evaluate(child, data, opts)
			evaluation.Children = append(evaluation.Children, childEvaluation)
			evaluation.Result = childEvaluation.Result || evaluation.Result
		}

		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	case rules.Not:
		evaluation := Evaluate(node.Children[0], data, opts)
		evaluation.Result = !evaluation.Result
		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	default:
		evaluation.Node.Value = resolveField(node.Field, data)
		evaluation.Result = evaluateRule(
			node.Operator, evaluation.Node.Value, node.Value,
		)
		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}

		return evaluation
	}
}

func resolveField(path string, data map[string]interface{}) interface{} {
	keys := strings.Split(path, ".")
	var current interface{} = data
	for _, key := range keys {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[key]
		} else {
			return nil
		}
	}
	return current
}

func evaluateRule(operator rules.Operator, actual, expected interface{}) bool {
	switch operator {
	// ---------- Equality ----------
	case rules.Eq:
		return compareEqual(actual, expected)
	case rules.Neq:
		return !compareEqual(actual, expected)

	// ---------- Numeric ----------
	case rules.Gt, rules.Gte, rules.Lt, rules.Lte:
		return compareNumeric(actual, expected, operator)
	case rules.Between:
		return isBetween(actual, expected)
	case rules.In:
		return inList(actual, expected)
	case rules.NotIn:
		return !inList(actual, expected)

	// ---------- String ----------
	case rules.Contains:
		return strings.Contains(toString(actual), toString(expected))
	case rules.NotContains:
		return !strings.Contains(toString(actual), toString(expected))
	case rules.StartsWith:
		return strings.HasPrefix(toString(actual), toString(expected))
	case rules.EndsWith:
		return strings.HasSuffix(toString(actual), toString(expected))
	case rules.Matches:
		re := regexp.MustCompile(toString(expected))
		return re.MatchString(toString(actual))
	case rules.LengthEq, rules.LengthGt, rules.LengthLt:
		return compareLength(actual, expected, operator)

	// ---------- Boolean ----------
	case rules.IsTrue:
		return actual == true
	case rules.IsFalse:
		return actual == false

	// ---------- Date ----------
	case rules.Before, rules.After:
		return compareTime(actual, expected, operator)
	case rules.DateBetween:
		return isTimeBetween(actual, expected)
	case rules.WithinLast, rules.WithinNext:
		return isWithinTime(actual, expected, operator)

	// ---------- Null / Existence ----------
	case rules.IsNull, rules.NotExists:
		return actual == nil
	case rules.IsNotNull, rules.Exists:
		return actual != nil

	// ---------- Type Checks ----------
	case rules.IsString:
		_, ok := actual.(string)
		return ok
	case rules.IsNumber:
		return isNumeric(actual)
	case rules.IsBool:
		_, ok := actual.(bool)
		return ok
	case rules.IsList:
		return reflect.TypeOf(actual).Kind() == reflect.Slice
	case rules.IsObject:
		return reflect.TypeOf(actual).Kind() == reflect.Map ||
			reflect.TypeOf(actual).Kind() == reflect.Struct
	case rules.IsDate:
		_, ok := actual.(time.Time)
		return ok

	// ---------- Custom Checks ----------
	case rules.Custom:
		argsList, ok := expected.([]interface{})
		if !ok || len(argsList) == 0 {
			return false
		}
		fnName, ok := argsList[0].(string)
		if !ok {
			return false
		}
		fn, found := rules.GetFunc(fnName)
		if !found {
			return false
		}

		return fn(append([]interface{}{actual}, argsList[1:]...)...)

	default:
		return false
	}
}
