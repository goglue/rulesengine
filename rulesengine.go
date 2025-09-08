package rulesengine

import (
	"reflect"
	"regexp"
	"strings"
	"time"
)

var reMatchMap = map[string]*regexp.Regexp{}

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
		Rule: Rule{
			Operator: node.Operator, Field: node.Field,
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

	case Or, Not:
		for _, child := range node.Children {
			childEvaluation := Evaluate(child, data, opts)
			evaluation.Children = append(evaluation.Children, childEvaluation)
			evaluation.Result = childEvaluation.Result || evaluation.Result
		}
		if node.Operator == Not {
			evaluation.Result = !evaluation.Result
		}
		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	case Any, All, None:
		arr, ok := toInterfaceSlice(resolveField(node.Field, data))
		if !ok {
			evaluation.Result = false
			evaluation.Error = newError(errType, node.Field)
			return evaluation
		}
		ruleVal, ok := node.Value.(Rule)
		if !ok {
			evaluation.Result = false
			evaluation.Error = newError(errType, node.Value)
			return evaluation
		}

		dataLen := len(arr)
		var passCount int
		for _, elem := range arr {
			res := Evaluate(ruleVal, map[string]any{"": elem}, opts)
			evaluation.Children = append(evaluation.Children, res)
			if res.Result {
				passCount++
			}
		}

		switch node.Operator {
		case Any:
			evaluation.Result = passCount > 0
		case All:
			evaluation.Result = passCount == dataLen
		case None:
			evaluation.Result = passCount == 0
		}

		if opts.Timing {
			evaluation.TimeTaken = time.Since(now)
		}
		return evaluation

	default:
		evaluation.Rule.Value = node.Value
		evaluation.Result, evaluation.Mismatch, evaluation.Error = evaluateRule(
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

func evaluateRule(operator Operator, actual, expected any) (bool, any, error) {
	switch operator {
	// ---------- Equality ----------
	case Eq:
		if !compareEqual(actual, expected) {
			return false, actual, nil
		}
		return true, nil, nil

	case Neq:
		if compareEqual(actual, expected) {
			return false, actual, nil
		}
		return true, nil, nil

	// ---------- Numeric ----------
	case Gt, Gte, Lt, Lte:
		if res, err := compareNumeric(actual, expected, operator); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	case Between:
		if res, err := isBetween(actual, expected); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	case In:
		if res, err := inList(actual, expected); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	case NotIn:
		if res, err := inList(actual, expected); err != nil {
			return false, nil, err
		} else if res {
			return false, actual, nil
		}
		return true, nil, nil

	case AnyIn:
		if res, err := anyInList(actual, expected); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	// ---------- String ----------
	case Contains:
		if !strings.Contains(toString(actual), toString(expected)) {
			return false, actual, nil
		}
		return true, nil, nil

	case NotContains:
		if strings.Contains(toString(actual), toString(expected)) {
			return false, actual, nil
		}
		return true, nil, nil

	case StartsWith:
		if !strings.HasPrefix(toString(actual), toString(expected)) {
			return false, actual, nil
		}
		return true, nil, nil

	case EndsWith:
		if !strings.HasSuffix(toString(actual), toString(expected)) {
			return false, actual, nil
		}
		return true, nil, nil

	case Matches:
		templateName := toString(expected)
		re, ok := reMatchMap[templateName]
		if !ok {
			re = regexp.MustCompile(templateName)
			reMatchMap[templateName] = re
		}
		if !re.MatchString(toString(actual)) {
			return false, actual, nil
		}
		return true, nil, nil

	case LengthEq, LengthGt, LengthLt:
		if res, err := compareLength(actual, expected, operator); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	// ---------- Boolean ----------
	case IsTrue:
		if actual != true {
			return false, actual, nil
		}
		return true, nil, nil

	case IsFalse:
		if actual != false {
			return false, actual, nil
		}
		return true, nil, nil

	// ---------- Date ----------
	case Before, After:
		if res, err := compareTime(actual, expected, operator); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	case DateBetween:
		if res, err := isTimeBetween(actual, expected); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	case WithinLast, WithinNext:
		if res, err := isWithinTime(actual, expected, operator); !res || err != nil {
			return res, actual, err
		}
		return true, nil, nil

	// ---------- Null / Existence ----------
	case IsNull, NotExists:
		if actual != nil {
			return false, actual, nil
		}
		return true, nil, nil

	case IsNotNull, Exists:
		if actual == nil {
			return false, actual, nil
		}
		return true, nil, nil

	// ---------- Type Checks ----------
	case IsString:
		_, ok := actual.(string)
		if !ok {
			return false, actual, newError(errType, actual)
		}
		return true, nil, nil

	case IsNumber:
		if !isNumeric(actual) {
			return false, actual, nil
		}
		return true, nil, nil

	case IsBool:
		if _, ok := actual.(bool); !ok {
			return false, actual, nil
		}
		return true, nil, nil

	case IsList:
		if reflect.TypeOf(actual).Kind() != reflect.Slice {
			return false, actual, nil
		}
		return true, nil, nil

	case IsObject:
		if reflect.TypeOf(actual).Kind() != reflect.Map &&
			reflect.TypeOf(actual).Kind() != reflect.Struct {
			return false, actual, nil
		}
		return true, nil, nil

	case IsDate:
		if _, ok := actual.(time.Time); !ok {
			return false, actual, nil
		}
		return true, nil, nil

	// ---------- Custom Checks ----------
	case Custom:
		argsList, ok := expected.([]any)
		if !ok || len(argsList) == 0 {
			return false, nil, newError(errType, expected)
		}
		fnName, ok := argsList[0].(string)
		if !ok {
			return false, nil, newError(errType, expected)
		}
		fn, found := GetFunc(fnName)
		if !found {
			return false, nil, newError(errType, "function not registered")
		}

		return fn(append([]any{actual}, argsList[1:]...)...)

	default:
		return false, nil, newError(errOperator, operator)
	}
}
