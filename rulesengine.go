package rulesengine

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var (
	reMatchMap  = map[string]*regexp.Regexp{}
	emptyValErr = newError("empty value", "")
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

	case IfThen:
		if len(node.Children) != 2 {
			evaluation.Result = false
			evaluation.Error = newError(errOperator, "IF_THEN requires exactly two child rules")
			return evaluation
		}
		ifEvaluation := Evaluate(node.Children[0], data, opts)
		thenEvaluation := Evaluate(node.Children[1], data, opts)
		evaluation.Children = append(evaluation.Children, ifEvaluation, thenEvaluation)
		// Material implication: A -> B is equivalent to !A or B
		evaluation.Result = !ifEvaluation.Result || thenEvaluation.Result

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
		evaluation.Input = resolveField(node.Field, data)
		evaluation.Result, evaluation.Error = evaluateRule(
			node.Operator, resolveField(node.Field, data), node.Value,
		)
		evaluation.IsEmpty = errors.Is(evaluation.Error, emptyValErr)
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

func evaluateRule(operator Operator, actual, expected any) (bool, error) {
	if actual == nil && operator != IsNull && operator != NotExists &&
		operator != IsNotNull && operator != Exists {
		return false, emptyValErr
	}

	switch operator {
	// ---------- Equality ----------
	case Eq:
		return compareEqual(actual, expected), nil

	case Neq:
		return !compareEqual(actual, expected), nil

	// ---------- Numeric ----------
	case Gt, Gte, Lt, Lte:
		if res, err := compareNumeric(actual, expected, operator); !res || err != nil {
			return res, err
		}
		return true, nil

	case Between:
		if res, err := isBetween(actual, expected); !res || err != nil {
			return res, err
		}
		return true, nil

	case In:
		if res, err := inList(actual, expected); !res || err != nil {
			return res, err
		}
		return true, nil

	case NotIn:
		if res, err := inList(actual, expected); err != nil {
			return false, err
		} else if res {
			return false, nil
		}
		return true, nil

	case AnyIn:
		if res, err := anyInList(actual, expected); !res || err != nil {
			return res, err
		}
		return true, nil

	// ---------- String ----------
	case Contains:
		if !strings.Contains(toString(actual), toString(expected)) {
			return false, nil
		}
		return true, nil

	case NotContains:
		if strings.Contains(toString(actual), toString(expected)) {
			return false, nil
		}
		return true, nil

	case StartsWith:
		if !strings.HasPrefix(toString(actual), toString(expected)) {
			return false, nil
		}
		return true, nil

	case EndsWith:
		if !strings.HasSuffix(toString(actual), toString(expected)) {
			return false, nil
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
			return false, nil
		}
		return true, nil

	case LengthEq, LengthGt, LengthLt:
		if res, err := compareLength(actual, expected, operator); !res || err != nil {
			return res, err
		}
		return true, nil

	// ---------- Boolean ----------
	case IsTrue:
		return actual == true, nil

	case IsFalse:
		return actual == false, nil

	// ---------- Date ----------
	case Before, After:
		if res, err := compareTime(actual, expected, operator); !res || err != nil {
			return res, err
		}
		return true, nil

	case DateBetween:
		if res, err := isTimeBetween(actual, expected); !res || err != nil {
			return res, err
		}
		return true, nil

	case WithinLast, WithinNext:
		if res, err := isWithinTime(actual, expected, operator); !res || err != nil {
			return res, err
		}
		return true, nil

	case YearEq, MonthEq:
		if res, err := compareTimePart(actual, expected, operator); !res || err != nil {
			return res, err
		}
		return true, nil

	// ---------- Null / Existence ----------
	case IsNull, NotExists:
		if actual != nil {
			return false, nil
		}
		return true, nil

	case IsNotNull, Exists:
		if actual == nil {
			return false, nil
		}
		return true, nil

	// ---------- Type Checks ----------
	case IsString:
		_, ok := actual.(string)
		if !ok {
			return false, newError(errType, actual)
		}
		return true, nil

	case IsNumber:
		if !isNumeric(actual) {
			return false, nil
		}
		return true, nil

	case IsBool:
		if _, ok := actual.(bool); !ok {
			return false, nil
		}
		return true, nil

	case IsList:
		if reflect.TypeOf(actual).Kind() != reflect.Slice {
			return false, nil
		}
		return true, nil

	case IsObject:
		if reflect.TypeOf(actual).Kind() != reflect.Map &&
			reflect.TypeOf(actual).Kind() != reflect.Struct {
			return false, nil
		}
		return true, nil

	case IsDate:
		if _, ok := actual.(time.Time); !ok {
			return false, nil
		}
		return true, nil

	// ---------- Custom Checks ----------
	case Custom:
		argsList, ok := expected.([]any)
		if !ok || len(argsList) == 0 {
			return false, newError(errType, expected)
		}
		fnName, ok := argsList[0].(string)
		if !ok {
			return false, newError(errType, expected)
		}
		fn, found := GetFunc(fnName)
		if !found {
			return false, newError(errType, "function not registered")
		}

		return fn(append([]any{actual}, argsList[1:]...)...)

	default:
		return false, newError(errOperator, operator)
	}
}
