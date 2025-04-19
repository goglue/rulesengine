package rulesengine

import (
	"github.com/goglue/rulesengine/rules"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	reMatchMap = map[string]*regexp.Regexp{}
)

func Evaluate(
	node rules.Rule, data map[string]any, opts Options,
) rules.RuleResult {
	var now time.Time
	if opts.Timing {
		now = time.Now()
	}
	evaluation := rules.RuleResult{
		Node: rules.Rule{
			Operator: node.Operator,
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
		evaluation.Node.Value = node.Value
		evaluation.Result, evaluation.Mismatch = evaluateRule(
			node.Operator, resolveField(node.Field, data), node.Value,
		)
		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}

		return evaluation
	}
}

func resolveField(path string, data map[string]any) any {
	keys := strings.Split(path, ".")
	var current any = data
	for _, key := range keys {
		if m, ok := current.(map[string]any); ok {
			current = m[key]
		} else {
			return nil
		}
	}
	return current
}

func evaluateRule(operator rules.Operator, actual, expected any) (bool, any) {
	switch operator {
	// ---------- Equality ----------
	case rules.Eq:
		if !compareEqual(actual, expected) {
			return false, actual
		}
		return true, nil

	case rules.Neq:
		if compareEqual(actual, expected) {
			return false, actual
		}
		return true, nil

	// ---------- Numeric ----------
	case rules.Gt, rules.Gte, rules.Lt, rules.Lte:
		if !compareNumeric(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	case rules.Between:
		if !isBetween(actual, expected) {
			return false, actual
		}
		return true, nil

	case rules.In:
		if !inList(actual, expected) {
			return false, actual
		}
		return true, nil

	case rules.NotIn:
		if inList(actual, expected) {
			return false, actual
		}
		return true, nil

	// ---------- String ----------
	case rules.Contains:
		if !strings.Contains(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case rules.NotContains:
		if strings.Contains(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case rules.StartsWith:
		if !strings.HasPrefix(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case rules.EndsWith:
		if !strings.HasSuffix(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case rules.Matches:
		templateName := toString(expected)
		re, ok := reMatchMap[templateName]
		if !ok {
			re = regexp.MustCompile(templateName)
			reMatchMap[templateName] = re
		}
		if !re.MatchString(toString(actual)) {
			return false, actual
		}
		return true, nil

	case rules.LengthEq, rules.LengthGt, rules.LengthLt:
		if !compareLength(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	// ---------- Boolean ----------
	case rules.IsTrue:
		if actual != true {
			return false, actual
		}
		return true, nil

	case rules.IsFalse:
		if actual != false {
			return false, actual
		}
		return true, nil

	// ---------- Date ----------
	case rules.Before, rules.After:
		if !compareTime(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	case rules.DateBetween:
		if !isTimeBetween(actual, expected) {
			return false, actual
		}
		return true, nil

	case rules.WithinLast, rules.WithinNext:
		if !isWithinTime(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	// ---------- Null / Existence ----------
	case rules.IsNull, rules.NotExists:
		if actual != nil {
			return false, actual
		}
		return true, nil

	case rules.IsNotNull, rules.Exists:
		if actual == nil {
			return false, actual
		}
		return true, nil

	// ---------- Type Checks ----------
	case rules.IsString:
		_, ok := actual.(string)
		if !ok {
			return false, actual
		}
		return true, nil

	case rules.IsNumber:
		if !isNumeric(actual) {
			return false, actual
		}
		return true, nil

	case rules.IsBool:
		if _, ok := actual.(bool); !ok {
			return false, actual
		}
		return true, nil

	case rules.IsList:
		if reflect.TypeOf(actual).Kind() != reflect.Slice {
			return false, actual
		}
		return true, nil

	case rules.IsObject:
		if reflect.TypeOf(actual).Kind() != reflect.Map &&
			reflect.TypeOf(actual).Kind() != reflect.Struct {
			return false, actual
		}
		return true, nil

	case rules.IsDate:
		if _, ok := actual.(time.Time); !ok {
			return false, actual
		}
		return true, nil

	// ---------- Custom Checks ----------
	case rules.Custom:
		argsList, ok := expected.([]any)
		if !ok || len(argsList) == 0 {
			return false, nil
		}
		fnName, ok := argsList[0].(string)
		if !ok {
			return false, nil
		}
		fn, found := rules.GetFunc(fnName)
		if !found {
			return false, nil
		}

		return fn(append([]any{actual}, argsList[1:]...)...)

	default:
		return false, nil
	}
}
