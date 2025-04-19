package rulesengine

import (
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	reMatchMap = map[string]*regexp.Regexp{}
)

// Evaluate method executes the evaluation of the passed rule and all its
// children, it returns [RuleResult] containing the rule evaluation results.
func Evaluate(
	node Rule, data map[string]any, opts Options,
) RuleResult {
	var now time.Time
	if opts.Timing {
		now = time.Now()
	}
	evaluation := RuleResult{
		Node: Rule{
			Operator: node.Operator,
		},
	}

	switch node.Operator {
	case And:
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

	case Or:
		for _, child := range node.Children {
			childEvaluation := Evaluate(child, data, opts)
			evaluation.Children = append(evaluation.Children, childEvaluation)
			evaluation.Result = childEvaluation.Result || evaluation.Result
		}

		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	case Not:
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

func evaluateRule(operator Operator, actual, expected any) (bool, any) {
	switch operator {
	// ---------- Equality ----------
	case Eq:
		if !compareEqual(actual, expected) {
			return false, actual
		}
		return true, nil

	case Neq:
		if compareEqual(actual, expected) {
			return false, actual
		}
		return true, nil

	// ---------- Numeric ----------
	case Gt, Gte, Lt, Lte:
		if !compareNumeric(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	case Between:
		if !isBetween(actual, expected) {
			return false, actual
		}
		return true, nil

	case In:
		if !inList(actual, expected) {
			return false, actual
		}
		return true, nil

	case NotIn:
		if inList(actual, expected) {
			return false, actual
		}
		return true, nil

	// ---------- String ----------
	case Contains:
		if !strings.Contains(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case NotContains:
		if strings.Contains(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case StartsWith:
		if !strings.HasPrefix(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case EndsWith:
		if !strings.HasSuffix(toString(actual), toString(expected)) {
			return false, actual
		}
		return true, nil

	case Matches:
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

	case LengthEq, LengthGt, LengthLt:
		if !compareLength(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	// ---------- Boolean ----------
	case IsTrue:
		if actual != true {
			return false, actual
		}
		return true, nil

	case IsFalse:
		if actual != false {
			return false, actual
		}
		return true, nil

	// ---------- Date ----------
	case Before, After:
		if !compareTime(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	case DateBetween:
		if !isTimeBetween(actual, expected) {
			return false, actual
		}
		return true, nil

	case WithinLast, WithinNext:
		if !isWithinTime(actual, expected, operator) {
			return false, actual
		}
		return true, nil

	// ---------- Null / Existence ----------
	case IsNull, NotExists:
		if actual != nil {
			return false, actual
		}
		return true, nil

	case IsNotNull, Exists:
		if actual == nil {
			return false, actual
		}
		return true, nil

	// ---------- Type Checks ----------
	case IsString:
		_, ok := actual.(string)
		if !ok {
			return false, actual
		}
		return true, nil

	case IsNumber:
		if !isNumeric(actual) {
			return false, actual
		}
		return true, nil

	case IsBool:
		if _, ok := actual.(bool); !ok {
			return false, actual
		}
		return true, nil

	case IsList:
		if reflect.TypeOf(actual).Kind() != reflect.Slice {
			return false, actual
		}
		return true, nil

	case IsObject:
		if reflect.TypeOf(actual).Kind() != reflect.Map &&
			reflect.TypeOf(actual).Kind() != reflect.Struct {
			return false, actual
		}
		return true, nil

	case IsDate:
		if _, ok := actual.(time.Time); !ok {
			return false, actual
		}
		return true, nil

	// ---------- Custom Checks ----------
	case Custom:
		argsList, ok := expected.([]any)
		if !ok || len(argsList) == 0 {
			return false, nil
		}
		fnName, ok := argsList[0].(string)
		if !ok {
			return false, nil
		}
		fn, found := GetFunc(fnName)
		if !found {
			return false, nil
		}

		return fn(append([]any{actual}, argsList[1:]...)...)

	default:
		return false, nil
	}
}
