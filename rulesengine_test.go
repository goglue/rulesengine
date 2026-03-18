package rulesengine

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// eval is a convenience wrapper used throughout the tests.
func eval(rule Rule, data map[string]any) RuleResult {
	return Evaluate(rule, data, DefaultOptions())
}

// evalWithTiming calls Evaluate with timing and a no-op logger enabled.
func evalWithTiming(rule Rule, data map[string]any) RuleResult {
	return Evaluate(rule, data, DefaultOptions().WithTiming().WithLogger(
		func(_ string, _ Operator, _, _ any) {},
	))
}

// ────────────────────────────────────────────────────────────────────────────
// Logical operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Logical(t *testing.T) {
	t.Run("AND/all children true returns true", func(t *testing.T) {
		rule := Rule{
			Operator: And,
			Children: []Rule{
				{Operator: Eq, Field: "a", Value: 1},
				{Operator: Eq, Field: "b", Value: 2},
			},
		}
		res := eval(rule, map[string]any{"a": 1, "b": 2})
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
		assert.Len(t, res.Children, 2)
	})

	t.Run("AND/one child false returns false", func(t *testing.T) {
		rule := Rule{
			Operator: And,
			Children: []Rule{
				{Operator: Eq, Field: "a", Value: 1},
				{Operator: Eq, Field: "b", Value: 99},
			},
		}
		res := eval(rule, map[string]any{"a": 1, "b": 2})
		assert.False(t, res.Result)
	})

	t.Run("AND/nested dot-notation fields", func(t *testing.T) {
		rule := Rule{
			Operator: And,
			Children: []Rule{
				{Operator: Eq, Field: "user.firstName", Value: "John"},
				{Operator: Eq, Field: "user.lastName", Value: "Doe"},
			},
		}
		res := eval(rule, map[string]any{
			"user": map[string]any{"firstName": "John", "lastName": "Doe"},
		})
		assert.True(t, res.Result)
	})

	t.Run("OR/at least one child true returns true", func(t *testing.T) {
		rule := Rule{
			Operator: Or,
			Children: []Rule{
				{Operator: Eq, Field: "x", Value: "wrong"},
				{Operator: Eq, Field: "x", Value: "right"},
			},
		}
		res := eval(rule, map[string]any{"x": "right"})
		assert.True(t, res.Result)
	})

	t.Run("OR/all children false returns false", func(t *testing.T) {
		rule := Rule{
			Operator: Or,
			Children: []Rule{
				{Operator: Eq, Field: "x", Value: "no"},
				{Operator: Eq, Field: "x", Value: "nope"},
			},
		}
		res := eval(rule, map[string]any{"x": "right"})
		assert.False(t, res.Result)
	})

	t.Run("NOT/negates false child to true", func(t *testing.T) {
		rule := Rule{
			Operator: Not,
			Children: []Rule{
				{Operator: Eq, Field: "firstName", Value: "John"},
			},
		}
		res := eval(rule, map[string]any{"firstName": "Mustermann"})
		assert.True(t, res.Result)
	})

	t.Run("NOT/negates true child to false", func(t *testing.T) {
		rule := Rule{
			Operator: Not,
			Children: []Rule{
				{Operator: Eq, Field: "firstName", Value: "John"},
			},
		}
		res := eval(rule, map[string]any{"firstName": "John"})
		assert.False(t, res.Result)
	})

	t.Run("IfThen/true implies true is true", func(t *testing.T) {
		rule := Rule{
			Operator: IfThen,
			Children: []Rule{
				{Operator: IsTrue, Field: "a"},
				{Operator: IsTrue, Field: "b"},
			},
		}
		res := eval(rule, map[string]any{"a": true, "b": true})
		assert.True(t, res.Result)
		assert.Len(t, res.Children, 2)
	})

	t.Run("IfThen/true implies false is false", func(t *testing.T) {
		rule := Rule{
			Operator: IfThen,
			Children: []Rule{
				{Operator: IsTrue, Field: "a"},
				{Operator: IsTrue, Field: "b"},
			},
		}
		res := eval(rule, map[string]any{"a": true, "b": false})
		assert.False(t, res.Result)
	})

	t.Run("IfThen/false implies true is vacuously true", func(t *testing.T) {
		rule := Rule{
			Operator: IfThen,
			Children: []Rule{
				{Operator: IsTrue, Field: "a"},
				{Operator: IsTrue, Field: "b"},
			},
		}
		res := eval(rule, map[string]any{"a": false, "b": true})
		assert.True(t, res.Result)
	})

	t.Run("IfThen/false implies false is vacuously true", func(t *testing.T) {
		rule := Rule{
			Operator: IfThen,
			Children: []Rule{
				{Operator: IsTrue, Field: "a"},
				{Operator: IsTrue, Field: "b"},
			},
		}
		res := eval(rule, map[string]any{"a": false, "b": false})
		assert.True(t, res.Result)
	})

	t.Run("IfThen/wrong number of children returns false with error", func(t *testing.T) {
		rule := Rule{
			Operator: IfThen,
			Children: []Rule{
				{Operator: IsTrue, Field: "a"},
			},
		}
		res := eval(rule, map[string]any{"a": true})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
		assert.Contains(t, res.Error.Error(), "IF_THEN requires exactly two child rules")
	})

	t.Run("IfThen/no children returns false with error", func(t *testing.T) {
		rule := Rule{Operator: IfThen, Children: []Rule{}}
		res := eval(rule, map[string]any{})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	t.Run("AND/timing records non-zero duration", func(t *testing.T) {
		rule := Rule{
			Operator: And,
			Children: []Rule{
				{Operator: Eq, Field: "v", Value: 1},
			},
		}
		res := evalWithTiming(rule, map[string]any{"v": 1})
		assert.True(t, res.Result)
		assert.Positive(t, res.TimeTaken)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Equality operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Equality(t *testing.T) {
	t.Run("EQ/string match", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "name", Value: "John"}, map[string]any{"name": "John"})
		assert.True(t, res.Result)
	})

	t.Run("EQ/string mismatch", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "name", Value: "John"}, map[string]any{"name": "Jane"})
		assert.False(t, res.Result)
	})

	t.Run("EQ/integer match", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "age", Value: 30}, map[string]any{"age": 30})
		assert.True(t, res.Result)
	})

	t.Run("EQ/bool match", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "active", Value: true}, map[string]any{"active": true})
		assert.True(t, res.Result)
	})

	t.Run("NEQ/values differ returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Neq, Field: "name", Value: "Johny"}, map[string]any{"name": "John"})
		assert.True(t, res.Result)
	})

	t.Run("NEQ/values equal returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Neq, Field: "name", Value: "John"}, map[string]any{"name": "John"})
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Numeric operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Numeric(t *testing.T) {
	t.Run("GT/greater returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Gt, Field: "age", Value: 10}, map[string]any{"age": 11})
		assert.True(t, res.Result)
	})

	t.Run("GT/equal returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Gt, Field: "age", Value: 10}, map[string]any{"age": 10})
		assert.False(t, res.Result)
	})

	t.Run("GT/non-numeric field returns error", func(t *testing.T) {
		res := eval(Rule{Operator: Gt, Field: "name", Value: 100}, map[string]any{"name": "Ali"})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
		assert.Contains(t, res.Error.Error(), "invalid numerical value")
	})

	t.Run("GTE/equal returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Gte, Field: "age", Value: 10}, map[string]any{"age": 10})
		assert.True(t, res.Result)
	})

	t.Run("GTE/less returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Gte, Field: "age", Value: 10}, map[string]any{"age": 9})
		assert.False(t, res.Result)
	})

	t.Run("LT/less returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Lt, Field: "age", Value: 10}, map[string]any{"age": 9})
		assert.True(t, res.Result)
	})

	t.Run("LT/equal returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Lt, Field: "age", Value: 10}, map[string]any{"age": 10})
		assert.False(t, res.Result)
	})

	t.Run("LTE/equal returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Lte, Field: "age", Value: 10}, map[string]any{"age": 10})
		assert.True(t, res.Result)
	})

	t.Run("LTE/greater returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Lte, Field: "age", Value: 10}, map[string]any{"age": 11})
		assert.False(t, res.Result)
	})

	t.Run("BETWEEN/value inside range returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Between, Field: "age", Value: []any{10, 20}}, map[string]any{"age": 15})
		assert.True(t, res.Result)
	})

	t.Run("BETWEEN/value on lower boundary returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Between, Field: "age", Value: []any{10, 20}}, map[string]any{"age": 10})
		assert.True(t, res.Result)
	})

	t.Run("BETWEEN/value on upper boundary returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Between, Field: "age", Value: []any{10, 20}}, map[string]any{"age": 20})
		assert.True(t, res.Result)
	})

	t.Run("BETWEEN/value outside range returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Between, Field: "age", Value: []any{10, 20}}, map[string]any{"age": 21})
		assert.False(t, res.Result)
	})

	t.Run("GT/uint32 value passes numeric check", func(t *testing.T) {
		res := eval(Rule{Operator: Gt, Field: "score", Value: uint32(5)}, map[string]any{"score": uint32(10)})
		assert.True(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Membership operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Membership(t *testing.T) {
	t.Run("IN/value present returns true", func(t *testing.T) {
		res := eval(Rule{Operator: In, Field: "role", Value: []any{"admin", "manager"}}, map[string]any{"role": "admin"})
		assert.True(t, res.Result)
	})

	t.Run("IN/value absent returns false", func(t *testing.T) {
		res := eval(Rule{Operator: In, Field: "role", Value: []any{"admin", "manager"}}, map[string]any{"role": "editor"})
		assert.False(t, res.Result)
	})

	t.Run("NOT_IN/value absent returns true", func(t *testing.T) {
		res := eval(Rule{Operator: NotIn, Field: "role", Value: []any{"admin", "manager"}}, map[string]any{"role": "editor"})
		assert.True(t, res.Result)
	})

	t.Run("NOT_IN/value present returns false", func(t *testing.T) {
		res := eval(Rule{Operator: NotIn, Field: "role", Value: []any{"admin", "manager"}}, map[string]any{"role": "manager"})
		assert.False(t, res.Result)
	})

	t.Run("ANY_IN/overlap returns true ([]any vs []any)", func(t *testing.T) {
		res := eval(
			Rule{Operator: AnyIn, Field: "roles", Value: []any{"admin", "manager"}},
			map[string]any{"roles": []any{"editor", "admin"}},
		)
		assert.True(t, res.Result)
	})

	t.Run("ANY_IN/no overlap returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: AnyIn, Field: "roles", Value: []any{"admin", "manager"}},
			map[string]any{"roles": []any{"editor", "dev"}},
		)
		assert.False(t, res.Result)
	})

	t.Run("ANY_IN/typed slices []int vs []int", func(t *testing.T) {
		res := eval(
			Rule{Operator: AnyIn, Field: "nums", Value: []int{1, 2}},
			map[string]any{"nums": []int{1, 3}},
		)
		assert.True(t, res.Result)
	})

	t.Run("ANY_IN/typed slices []any vs []int", func(t *testing.T) {
		res := eval(
			Rule{Operator: AnyIn, Field: "nums", Value: []any{1, 2}},
			map[string]any{"nums": []int{1, 3}},
		)
		assert.True(t, res.Result)
	})

	t.Run("ANY_IN/typed slices []int vs []any", func(t *testing.T) {
		res := eval(
			Rule{Operator: AnyIn, Field: "nums", Value: []int{1, 2}},
			map[string]any{"nums": []any{1, 3}},
		)
		assert.True(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// String operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_String(t *testing.T) {
	t.Run("CONTAINS/substring present returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Contains, Field: "secret", Value: "%"}, map[string]any{"secret": "some%password"})
		assert.True(t, res.Result)
	})

	t.Run("CONTAINS/substring absent returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Contains, Field: "secret", Value: "%"}, map[string]any{"secret": "some_password"})
		assert.False(t, res.Result)
	})

	t.Run("NOT_CONTAINS/substring absent returns true", func(t *testing.T) {
		res := eval(Rule{Operator: NotContains, Field: "secret", Value: "%"}, map[string]any{"secret": "some_password"})
		assert.True(t, res.Result)
	})

	t.Run("NOT_CONTAINS/substring present returns false", func(t *testing.T) {
		res := eval(Rule{Operator: NotContains, Field: "secret", Value: "%"}, map[string]any{"secret": "some%password"})
		assert.False(t, res.Result)
	})

	t.Run("STARTS_WITH/prefix matches returns true", func(t *testing.T) {
		res := eval(Rule{Operator: StartsWith, Field: "s", Value: "some"}, map[string]any{"s": "some%password"})
		assert.True(t, res.Result)
	})

	t.Run("STARTS_WITH/prefix mismatch returns false", func(t *testing.T) {
		res := eval(Rule{Operator: StartsWith, Field: "s", Value: "password"}, map[string]any{"s": "some%password"})
		assert.False(t, res.Result)
	})

	t.Run("ENDS_WITH/suffix matches returns true", func(t *testing.T) {
		res := eval(Rule{Operator: EndsWith, Field: "s", Value: "password"}, map[string]any{"s": "some%password"})
		assert.True(t, res.Result)
	})

	t.Run("ENDS_WITH/suffix mismatch returns false", func(t *testing.T) {
		res := eval(Rule{Operator: EndsWith, Field: "s", Value: "some"}, map[string]any{"s": "some%password"})
		assert.False(t, res.Result)
	})

	t.Run("MATCHES/pattern matches returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Matches, Field: "word", Value: "p([a-z]+)ch"}, map[string]any{"word": "peach"})
		assert.True(t, res.Result)
	})

	t.Run("MATCHES/pattern no match returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Matches, Field: "word", Value: "p([a-z]+)ch"}, map[string]any{"word": "pencil"})
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Length operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Length(t *testing.T) {
	t.Run("LENGTH_EQ/exact length returns true", func(t *testing.T) {
		res := eval(Rule{Operator: LengthEq, Field: "s", Value: "3"}, map[string]any{"s": "abc"})
		assert.True(t, res.Result)
	})

	t.Run("LENGTH_EQ/wrong length returns false", func(t *testing.T) {
		res := eval(Rule{Operator: LengthEq, Field: "s", Value: "3"}, map[string]any{"s": "abcd"})
		assert.False(t, res.Result)
	})

	t.Run("LENGTH_GT/longer returns true", func(t *testing.T) {
		res := eval(Rule{Operator: LengthGt, Field: "s", Value: "3"}, map[string]any{"s": "abcd"})
		assert.True(t, res.Result)
	})

	t.Run("LENGTH_GT/equal not greater returns false", func(t *testing.T) {
		res := eval(Rule{Operator: LengthGt, Field: "s", Value: "3"}, map[string]any{"s": "abc"})
		assert.False(t, res.Result)
	})

	t.Run("LENGTH_LT/shorter returns true", func(t *testing.T) {
		res := eval(Rule{Operator: LengthLt, Field: "s", Value: "3"}, map[string]any{"s": "ab"})
		assert.True(t, res.Result)
	})

	t.Run("LENGTH_LT/equal not less returns false", func(t *testing.T) {
		res := eval(Rule{Operator: LengthLt, Field: "s", Value: "3"}, map[string]any{"s": "abc"})
		assert.False(t, res.Result)
	})

	t.Run("LENGTH_GT/integer target value", func(t *testing.T) {
		res := eval(Rule{Operator: LengthGt, Field: "s", Value: 2}, map[string]any{"s": "abc"})
		assert.True(t, res.Result)
	})

	t.Run("LENGTH_EQ/slice field counts elements", func(t *testing.T) {
		res := eval(Rule{Operator: LengthEq, Field: "items", Value: "3"}, map[string]any{"items": []int{1, 2, 3}})
		assert.True(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Boolean operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Boolean(t *testing.T) {
	t.Run("IS_TRUE/true value returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsTrue, Field: "flag"}, map[string]any{"flag": true})
		assert.True(t, res.Result)
	})

	t.Run("IS_TRUE/false value returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsTrue, Field: "flag"}, map[string]any{"flag": false})
		assert.False(t, res.Result)
	})

	t.Run("IS_FALSE/false value returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsFalse, Field: "flag"}, map[string]any{"flag": false})
		assert.True(t, res.Result)
	})

	t.Run("IS_FALSE/true value returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsFalse, Field: "flag"}, map[string]any{"flag": true})
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Date operators (time.Time values)
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Date(t *testing.T) {
	past, _ := time.Parse("2006-01-02", "2014-12-01")
	future, _ := time.Parse("2006-01-02", "2015-01-01")
	now := time.Now()

	t.Run("BEFORE/past date before future date returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Before, Field: "d", Value: future}, map[string]any{"d": past})
		assert.True(t, res.Result)
	})

	t.Run("BEFORE/future date before past date returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Before, Field: "d", Value: past}, map[string]any{"d": future})
		assert.False(t, res.Result)
	})

	t.Run("AFTER/future date after past date returns true", func(t *testing.T) {
		res := eval(Rule{Operator: After, Field: "d", Value: past}, map[string]any{"d": future})
		assert.True(t, res.Result)
	})

	t.Run("AFTER/past date after future date returns false", func(t *testing.T) {
		res := eval(Rule{Operator: After, Field: "d", Value: future}, map[string]any{"d": past})
		assert.False(t, res.Result)
	})

	t.Run("DATE_BETWEEN/value inside window returns true", func(t *testing.T) {
		res := eval(
			Rule{Operator: DateBetween, Field: "d", Value: []time.Time{future, future.Add(24 * time.Hour)}},
			map[string]any{"d": future.Add(3 * time.Hour)},
		)
		assert.True(t, res.Result)
	})

	t.Run("DATE_BETWEEN/value outside window returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: DateBetween, Field: "d", Value: []time.Time{future, future.Add(24 * time.Hour)}},
			map[string]any{"d": future.Add(-3 * time.Hour)},
		)
		assert.False(t, res.Result)
	})

	t.Run("DATE_BETWEEN/relative string range covers current year", func(t *testing.T) {
		inYear := time.Date(now.Year(), time.June, 1, 12, 0, 0, 0, now.Location())
		res := eval(
			Rule{Operator: DateBetween, Field: "d", Value: []any{"thisYear", "thisYear+1y"}},
			map[string]any{"d": inYear},
		)
		assert.True(t, res.Result)
	})

	t.Run("BEFORE/relative future string returns true for current year date", func(t *testing.T) {
		inYear := time.Date(now.Year(), time.June, 1, 12, 0, 0, 0, now.Location())
		res := eval(
			Rule{Operator: Before, Field: "d", Value: "thisYear+1y"},
			map[string]any{"d": inYear},
		)
		assert.True(t, res.Result)
	})

	t.Run("AFTER/relative past string returns true for current year date", func(t *testing.T) {
		inYear := time.Date(now.Year(), time.June, 1, 12, 0, 0, 0, now.Location())
		res := eval(
			Rule{Operator: After, Field: "d", Value: "thisYear-1"},
			map[string]any{"d": inYear},
		)
		assert.True(t, res.Result)
	})

	t.Run("YEAR_EQ/current year matches", func(t *testing.T) {
		res := eval(
			Rule{Operator: YearEq, Field: "d", Value: now.Year()},
			map[string]any{"d": time.Date(now.Year(), time.June, 1, 0, 0, 0, 0, time.UTC)},
		)
		assert.True(t, res.Result)
	})

	t.Run("YEAR_EQ/wrong year returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: YearEq, Field: "d", Value: 2020},
			map[string]any{"d": time.Date(now.Year(), time.June, 1, 0, 0, 0, 0, time.UTC)},
		)
		assert.False(t, res.Result)
	})

	t.Run("YEAR_EQ/relative string thisYear-1 matches date one year ago", func(t *testing.T) {
		res := eval(
			Rule{Operator: YearEq, Field: "d", Value: "thisYear-1"},
			map[string]any{"d": time.Date(now.Year(), time.June, 1, 0, 0, 0, 0, time.UTC).AddDate(-1, 0, 0)},
		)
		assert.True(t, res.Result)
	})

	t.Run("MONTH_EQ/June matches int(6)", func(t *testing.T) {
		res := eval(
			Rule{Operator: MonthEq, Field: "d", Value: int(time.June)},
			map[string]any{"d": time.Date(now.Year(), time.June, 15, 0, 0, 0, 0, time.UTC)},
		)
		assert.True(t, res.Result)
	})

	t.Run("MONTH_EQ/thisMonth relative string matches", func(t *testing.T) {
		res := eval(
			Rule{Operator: MonthEq, Field: "d", Value: "thisMonth"},
			map[string]any{"d": time.Date(now.Year(), now.Month(), 1, 12, 0, 0, 0, now.Location())},
		)
		assert.True(t, res.Result)
	})

	t.Run("WITHIN_LAST/5s ago within 10s window returns true", func(t *testing.T) {
		res := eval(
			Rule{Operator: WithinLast, Field: "d", Value: "10s"},
			map[string]any{"d": now.Add(-5 * time.Second)},
		)
		assert.True(t, res.Result)
	})

	t.Run("WITHIN_LAST/15s ago outside 10s window returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: WithinLast, Field: "d", Value: "10s"},
			map[string]any{"d": now.Add(-15 * time.Second)},
		)
		assert.False(t, res.Result)
	})

	t.Run("WITHIN_NEXT/5s ahead within 10s window returns true", func(t *testing.T) {
		res := eval(
			Rule{Operator: WithinNext, Field: "d", Value: "10s"},
			map[string]any{"d": now.Add(5 * time.Second)},
		)
		assert.True(t, res.Result)
	})

	t.Run("WITHIN_NEXT/11s ahead outside 10s window returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: WithinNext, Field: "d", Value: "10s"},
			map[string]any{"d": now.Add(11 * time.Second)},
		)
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Date operators with RFC3339 string inputs (Fix 2)
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_DateStrings(t *testing.T) {
	// RFC3339 strings simulate what happens after a JSON round-trip of time.Time.

	t.Run("BEFORE/RFC3339 string data before time.Time value", func(t *testing.T) {
		res := eval(
			Rule{Operator: Before, Field: "d", Value: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("BEFORE/RFC3339 string data after cutoff returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: Before, Field: "d", Value: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.False(t, res.Result)
	})

	t.Run("AFTER/RFC3339 string data after cutoff returns true", func(t *testing.T) {
		res := eval(
			Rule{Operator: After, Field: "d", Value: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("DATE_BETWEEN/RFC3339 string data within window", func(t *testing.T) {
		lo := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		hi := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		res := eval(
			Rule{Operator: DateBetween, Field: "d", Value: []time.Time{lo, hi}},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("DATE_BETWEEN/RFC3339 string data outside window", func(t *testing.T) {
		lo := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		hi := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		res := eval(
			Rule{Operator: DateBetween, Field: "d", Value: []time.Time{lo, hi}},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.False(t, res.Result)
	})

	t.Run("WITHIN_LAST/RFC3339 string 2s ago within 10s window", func(t *testing.T) {
		twoSecondsAgo := time.Now().Add(-2 * time.Second).UTC().Format(time.RFC3339)
		res := eval(
			Rule{Operator: WithinLast, Field: "d", Value: "10s"},
			map[string]any{"d": twoSecondsAgo},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("WITHIN_NEXT/RFC3339 string 2s ahead within 10s window", func(t *testing.T) {
		twoSecondsAhead := time.Now().Add(2 * time.Second).UTC().Format(time.RFC3339)
		res := eval(
			Rule{Operator: WithinNext, Field: "d", Value: "10s"},
			map[string]any{"d": twoSecondsAhead},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("YEAR_EQ/RFC3339 string year matches", func(t *testing.T) {
		res := eval(
			Rule{Operator: YearEq, Field: "d", Value: 2020},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("YEAR_EQ/RFC3339 string year mismatch returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: YearEq, Field: "d", Value: 2019},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.False(t, res.Result)
	})

	t.Run("MONTH_EQ/RFC3339 string month matches", func(t *testing.T) {
		// June is month 6
		res := eval(
			Rule{Operator: MonthEq, Field: "d", Value: int(time.June)},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("MONTH_EQ/RFC3339 string month mismatch returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: MonthEq, Field: "d", Value: int(time.January)},
			map[string]any{"d": "2020-06-15T00:00:00Z"},
		)
		assert.False(t, res.Result)
	})

	t.Run("BEFORE/date-only string (YYYY-MM-DD) is accepted", func(t *testing.T) {
		res := eval(
			Rule{Operator: Before, Field: "d", Value: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
			map[string]any{"d": "2020-06-15"},
		)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Array operators (ANY, ALL, NONE)
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Array(t *testing.T) {
	// ── Primitive elements ──────────────────────────────────────────────────

	t.Run("ALL/all strings longer than 3 returns true", func(t *testing.T) {
		rule := Rule{Operator: All, Field: "roles", Value: Rule{Operator: LengthGt, Value: 3}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.True(t, res.Result)
	})

	t.Run("ALL/not all strings longer than 5 returns false", func(t *testing.T) {
		rule := Rule{Operator: All, Field: "roles", Value: Rule{Operator: LengthGt, Value: 5}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.False(t, res.Result)
	})

	t.Run("ANY/at least one string longer than 5 returns true", func(t *testing.T) {
		rule := Rule{Operator: Any, Field: "roles", Value: Rule{Operator: LengthGt, Value: 5}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.True(t, res.Result)
	})

	t.Run("ANY/none longer than 10 returns false", func(t *testing.T) {
		rule := Rule{Operator: Any, Field: "roles", Value: Rule{Operator: LengthGt, Value: 10}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.False(t, res.Result)
	})

	t.Run("NONE/none longer than 10 returns true", func(t *testing.T) {
		rule := Rule{Operator: None, Field: "roles", Value: Rule{Operator: LengthGt, Value: 10}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.True(t, res.Result)
	})

	t.Run("NONE/some longer than 5 returns false", func(t *testing.T) {
		rule := Rule{Operator: None, Field: "roles", Value: Rule{Operator: LengthGt, Value: 5}}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.False(t, res.Result)
	})

	t.Run("ALL/compound AND predicate on each element", func(t *testing.T) {
		rule := Rule{
			Operator: All,
			Field:    "roles",
			Value: Rule{
				Operator: And,
				Children: []Rule{
					{Operator: LengthGt, Value: 1},
					{Operator: LengthLt, Value: 10},
					{Operator: Neq, Value: "admin"},
					{Operator: In, Value: []string{"test", "roles", "longer", "than", "three"}},
				},
			},
		}
		res := eval(rule, map[string]any{"roles": []string{"test", "roles", "longer", "than", "three"}})
		assert.True(t, res.Result)
	})

	t.Run("ANY/non-slice field returns false with error", func(t *testing.T) {
		rule := Rule{Operator: Any, Field: "roles", Value: Rule{Operator: LengthGt, Value: 1}}
		res := eval(rule, map[string]any{"roles": "not-a-slice"})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	// ── Object elements (Fix 1) ─────────────────────────────────────────────
	// Prior to the fix, object elements were wrapped under key "" so
	// field-based predicates like {Field: "DocumentTypeID", Operator: Eq}
	// would always fail. After the fix, map[string]any elements are passed
	// directly as the data map.

	t.Run("ANY/object elements - at least one document has DocumentTypeID==64", func(t *testing.T) {
		rule := Rule{
			Operator: Any,
			Field:    "Documents",
			Value:    Rule{Operator: Eq, Field: "DocumentTypeID", Value: float64(64)},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64), "IsValidated": true},
				map[string]any{"DocumentTypeID": float64(77), "IsValidated": true},
			},
		}
		res := eval(rule, data)
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("ANY/object elements - no document has DocumentTypeID==99 returns false", func(t *testing.T) {
		rule := Rule{
			Operator: Any,
			Field:    "Documents",
			Value:    Rule{Operator: Eq, Field: "DocumentTypeID", Value: float64(99)},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64), "IsValidated": true},
				map[string]any{"DocumentTypeID": float64(77), "IsValidated": true},
			},
		}
		res := eval(rule, data)
		assert.False(t, res.Result)
	})

	t.Run("ALL/object elements - all documents validated", func(t *testing.T) {
		rule := Rule{
			Operator: All,
			Field:    "Documents",
			Value:    Rule{Operator: IsTrue, Field: "IsValidated"},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64), "IsValidated": true},
				map[string]any{"DocumentTypeID": float64(77), "IsValidated": true},
			},
		}
		res := eval(rule, data)
		assert.True(t, res.Result)
	})

	t.Run("ALL/object elements - not all documents validated returns false", func(t *testing.T) {
		rule := Rule{
			Operator: All,
			Field:    "Documents",
			Value:    Rule{Operator: IsTrue, Field: "IsValidated"},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64), "IsValidated": true},
				map[string]any{"DocumentTypeID": float64(77), "IsValidated": false},
			},
		}
		res := eval(rule, data)
		assert.False(t, res.Result)
	})

	t.Run("NONE/object elements - no document has DocumentTypeID==99", func(t *testing.T) {
		rule := Rule{
			Operator: None,
			Field:    "Documents",
			Value:    Rule{Operator: Eq, Field: "DocumentTypeID", Value: float64(99)},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64)},
				map[string]any{"DocumentTypeID": float64(77)},
			},
		}
		res := eval(rule, data)
		assert.True(t, res.Result)
	})

	t.Run("NONE/object elements - one document has DocumentTypeID==64 returns false", func(t *testing.T) {
		rule := Rule{
			Operator: None,
			Field:    "Documents",
			Value:    Rule{Operator: Eq, Field: "DocumentTypeID", Value: float64(64)},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64)},
				map[string]any{"DocumentTypeID": float64(77)},
			},
		}
		res := eval(rule, data)
		assert.False(t, res.Result)
	})

	t.Run("ANY/object elements with nested AND predicate", func(t *testing.T) {
		rule := Rule{
			Operator: Any,
			Field:    "Documents",
			Value: Rule{
				Operator: And,
				Children: []Rule{
					{Operator: Eq, Field: "DocumentTypeID", Value: float64(64)},
					{Operator: IsTrue, Field: "IsValidated"},
				},
			},
		}
		data := map[string]any{
			"Documents": []any{
				map[string]any{"DocumentTypeID": float64(64), "IsValidated": true},
				map[string]any{"DocumentTypeID": float64(77), "IsValidated": false},
			},
		}
		res := eval(rule, data)
		assert.True(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Existence operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Existence(t *testing.T) {
	t.Run("IS_NULL/missing field returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsNull, Field: "someField"}, map[string]any{"anotherField": "x"})
		assert.True(t, res.Result)
	})

	t.Run("IS_NULL/present field returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsNull, Field: "someField"}, map[string]any{"someField": "x"})
		assert.False(t, res.Result)
	})

	t.Run("IS_NOT_NULL/present field returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsNotNull, Field: "someField"}, map[string]any{"someField": "x"})
		assert.True(t, res.Result)
	})

	t.Run("IS_NOT_NULL/missing field returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsNotNull, Field: "someField"}, map[string]any{"anotherField": "x"})
		assert.False(t, res.Result)
	})

	t.Run("EXISTS/present field returns true", func(t *testing.T) {
		res := eval(Rule{Operator: Exists, Field: "name"}, map[string]any{"name": "John"})
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("EXISTS/missing field returns false", func(t *testing.T) {
		res := eval(Rule{Operator: Exists, Field: "name"}, map[string]any{"age": 30})
		assert.False(t, res.Result)
	})

	t.Run("NOT_EXISTS/missing field returns true", func(t *testing.T) {
		res := eval(Rule{Operator: NotExists, Field: "name"}, map[string]any{"age": 30})
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("NOT_EXISTS/present field returns false", func(t *testing.T) {
		res := eval(Rule{Operator: NotExists, Field: "name"}, map[string]any{"name": "John"})
		assert.False(t, res.Result)
	})

	t.Run("EXISTS/nested field present via dot notation", func(t *testing.T) {
		res := eval(
			Rule{Operator: Exists, Field: "user.age"},
			map[string]any{"user": map[string]any{"age": 25}},
		)
		assert.True(t, res.Result)
	})

	t.Run("EXISTS/nested field missing via dot notation returns false", func(t *testing.T) {
		res := eval(
			Rule{Operator: Exists, Field: "user.email"},
			map[string]any{"user": map[string]any{"age": 25}},
		)
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Type check operators
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_TypeChecks(t *testing.T) {
	t.Run("IS_NUMBER/uint32 returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsNumber, Field: "v"}, map[string]any{"v": uint32(20)})
		assert.True(t, res.Result)
	})

	t.Run("IS_NUMBER/float64 returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsNumber, Field: "v"}, map[string]any{"v": float64(3.14)})
		assert.True(t, res.Result)
	})

	t.Run("IS_NUMBER/string returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsNumber, Field: "v"}, map[string]any{"v": "hello"})
		assert.False(t, res.Result)
	})

	t.Run("IS_STRING/string value returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsString, Field: "v"}, map[string]any{"v": "1300"})
		assert.True(t, res.Result)
	})

	t.Run("IS_STRING/integer returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsString, Field: "v"}, map[string]any{"v": 42})
		assert.False(t, res.Result)
	})

	t.Run("IS_BOOL/bool returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsBool, Field: "v"}, map[string]any{"v": true})
		assert.True(t, res.Result)
	})

	t.Run("IS_BOOL/string returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsBool, Field: "v"}, map[string]any{"v": "true"})
		assert.False(t, res.Result)
	})

	t.Run("IS_DATE/time.Time returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsDate, Field: "v"}, map[string]any{"v": time.Now()})
		assert.True(t, res.Result)
	})

	t.Run("IS_DATE/string returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsDate, Field: "v"}, map[string]any{"v": "2020-01-01"})
		assert.False(t, res.Result)
	})

	t.Run("IS_LIST/slice returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsList, Field: "v"}, map[string]any{"v": []int{1, 2, 3}})
		assert.True(t, res.Result)
	})

	t.Run("IS_LIST/map returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsList, Field: "v"}, map[string]any{"v": map[string]any{}})
		assert.False(t, res.Result)
	})

	t.Run("IS_OBJECT/map[string]any returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsObject, Field: "v"}, map[string]any{"v": map[string]any{}})
		assert.True(t, res.Result)
	})

	t.Run("IS_OBJECT/struct returns true", func(t *testing.T) {
		res := eval(Rule{Operator: IsObject, Field: "v"}, map[string]any{"v": struct{}{}})
		assert.True(t, res.Result)
	})

	t.Run("IS_OBJECT/slice returns false", func(t *testing.T) {
		res := eval(Rule{Operator: IsObject, Field: "v"}, map[string]any{"v": []int{1}})
		assert.False(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Field path resolution
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_FieldPaths(t *testing.T) {
	t.Run("top-level field resolved correctly", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "name", Value: "Alice"}, map[string]any{"name": "Alice"})
		assert.True(t, res.Result)
	})

	t.Run("one-level nested field resolved correctly", func(t *testing.T) {
		res := eval(
			Rule{Operator: Eq, Field: "user.name", Value: "Alice"},
			map[string]any{"user": map[string]any{"name": "Alice"}},
		)
		assert.True(t, res.Result)
	})

	t.Run("deeply nested field resolved correctly", func(t *testing.T) {
		res := eval(
			Rule{Operator: Eq, Field: "a.b.c.d", Value: "deep"},
			map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{"d": "deep"},
					},
				},
			},
		)
		assert.True(t, res.Result)
	})

	t.Run("missing top-level field returns IsEmpty", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "missing", Value: "x"}, map[string]any{})
		assert.False(t, res.Result)
		assert.True(t, res.IsEmpty)
	})

	t.Run("missing nested field returns IsEmpty", func(t *testing.T) {
		res := eval(
			Rule{Operator: Eq, Field: "user.email", Value: "x"},
			map[string]any{"user": map[string]any{"name": "Alice"}},
		)
		assert.False(t, res.Result)
		assert.True(t, res.IsEmpty)
	})

	t.Run("intermediate path not a map returns IsEmpty", func(t *testing.T) {
		res := eval(
			Rule{Operator: Eq, Field: "user.name.first", Value: "x"},
			map[string]any{"user": map[string]any{"name": "Alice"}},
		)
		assert.False(t, res.Result)
		assert.True(t, res.IsEmpty)
	})

	t.Run("complex nested AND rule with multiple fields", func(t *testing.T) {
		rule := Rule{
			Operator: And,
			Children: []Rule{
				{
					Operator: And,
					Field:    "user",
					Children: []Rule{
						{Field: "user.name", Operator: LengthGt, Value: 2},
						{Field: "user.name", Operator: LengthLt, Value: 25},
					},
				},
				{Field: "user.age", Operator: Gte, Value: 21},
				{Field: "user.country", Operator: Eq, Value: "DE"},
			},
		}
		res := eval(rule, map[string]any{
			"user": map[string]any{"name": "Sam", "age": 25, "country": "DE"},
		})
		assert.True(t, res.Result)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Custom function operator
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Custom(t *testing.T) {
	t.Run("registered custom function evaluates correctly", func(t *testing.T) {
		RegisterFunc("isEmail", func(args ...any) (bool, error) {
			dataEmail, ok := args[0].(string)
			if !ok {
				return false, nil
			}
			passedEmail, ok := args[1].(string)
			if !ok {
				return false, nil
			}
			return strings.Contains(dataEmail, "@") &&
				strings.Contains(dataEmail, ".") &&
				passedEmail == "floating.tester@domain.ext" &&
				dataEmail == "some.email@domain.ext", nil
		})

		rule := Rule{
			Operator: Custom,
			Field:    "email",
			Value:    []any{"isEmail", "floating.tester@domain.ext"},
		}
		res := eval(rule, map[string]any{"email": "some.email@domain.ext"})
		assert.True(t, res.Result)
		assert.Nil(t, res.Error)
	})

	t.Run("unregistered function name returns false with error", func(t *testing.T) {
		rule := Rule{
			Operator: Custom,
			Field:    "email",
			Value:    []any{"nonExistentFn", "arg1"},
		}
		res := eval(rule, map[string]any{"email": "test@example.com"})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
		assert.Contains(t, res.Error.Error(), "function not registered")
	})

	t.Run("custom function with non-slice value returns false with error", func(t *testing.T) {
		rule := Rule{
			Operator: Custom,
			Field:    "email",
			Value:    "not-a-slice",
		}
		res := eval(rule, map[string]any{"email": "test@example.com"})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	t.Run("GetFunc returns registered function", func(t *testing.T) {
		RegisterFunc("dummy", func(args ...any) (bool, error) { return true, nil })
		fn, ok := GetFunc("dummy")
		require.True(t, ok)
		result, err := fn()
		require.NoError(t, err)
		assert.True(t, result)
	})

	t.Run("GetFunc returns false for unknown name", func(t *testing.T) {
		_, ok := GetFunc("unknownFunctionXYZ")
		assert.False(t, ok)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Error cases
// ────────────────────────────────────────────────────────────────────────────

func TestEvaluate_Errors(t *testing.T) {
	t.Run("GT on string field returns error", func(t *testing.T) {
		res := eval(Rule{Operator: Gt, Field: "v", Value: 100}, map[string]any{"v": "not-a-number"})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
		assert.Contains(t, res.Error.Error(), "invalid numerical value")
	})

	t.Run("missing field with EQ sets IsEmpty, not an operator error", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "missing", Value: "x"}, map[string]any{})
		assert.False(t, res.Result)
		assert.True(t, res.IsEmpty)
		// The emptyValErr is returned, not nil, but IsEmpty is the semantic signal.
		require.NotNil(t, res.Error)
	})

	t.Run("unknown operator returns error", func(t *testing.T) {
		res := eval(Rule{Operator: Operator("UNKNOWN_OP"), Field: "v", Value: 1}, map[string]any{"v": 1})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
		assert.Contains(t, res.Error.Error(), "invalid operator")
	})

	t.Run("BETWEEN with non-slice value returns error", func(t *testing.T) {
		res := eval(Rule{Operator: Between, Field: "age", Value: "not-a-range"}, map[string]any{"age": 5})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	t.Run("BEFORE with non-time value returns error", func(t *testing.T) {
		res := eval(Rule{Operator: Before, Field: "d", Value: "not-a-date"}, map[string]any{"d": time.Now()})
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	t.Run("BEFORE with non-time field value returns error", func(t *testing.T) {
		res := eval(
			Rule{Operator: Before, Field: "d", Value: time.Now()},
			map[string]any{"d": "definitely-not-a-date"},
		)
		assert.False(t, res.Result)
		require.NotNil(t, res.Error)
	})

	t.Run("RuleResult Input is populated on leaf rule", func(t *testing.T) {
		res := eval(Rule{Operator: Eq, Field: "name", Value: "Alice"}, map[string]any{"name": "Alice"})
		assert.Equal(t, "Alice", res.Input)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Benchmark
// ────────────────────────────────────────────────────────────────────────────

func BenchmarkEvaluate(b *testing.B) {
	rule := Rule{
		Operator: And,
		Children: []Rule{
			{Operator: IsNumber, Field: "user.age"},
			{Operator: Gte, Field: "user.age", Value: 25},
			{Operator: Matches, Field: "user.jobTitle", Value: "s([a-z]+)re"},
			{Operator: IsObject, Field: "user.address"},
			{Operator: Eq, Field: "user.address.zipCode", Value: 5},
			{Operator: IsNotNull, Field: "user.address.streetName"},
			{Operator: LengthGt, Field: "user.address.streetName", Value: 5},
			{Operator: LengthGt, Field: "user.firstName", Value: 2},
			{Operator: LengthGt, Field: "user.lastName", Value: 2},
		},
	}
	data := map[string]any{
		"user": map[string]any{
			"firstName": "John",
			"lastName":  "Doe",
			"age":       25,
			"jobTitle":  "software",
			"address": map[string]any{
				"streetName": "Johannisstraße",
				"zipCode":    "13088",
			},
		},
	}
	opts := DefaultOptions()
	for i := 0; i < b.N; i++ {
		_ = Evaluate(rule, data, opts)
	}
}