package rulesengine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	assertion "github.com/stretchr/testify/assert"
)

var (
	parsedTimeData, _ = time.Parse("2006-01-02", "2014-12-01")
	parsedTimeExp, _  = time.Parse("2006-01-02", "2015-01-01")
	last10Sec         = "10s"
	relativeNow       = time.Now()
	relativeTime      = time.Date(relativeNow.Year(), time.June, 1, 12, 0, 0, 0, relativeNow.Location())
	testData          = []struct {
		ruleNodes Rule
		inputData map[string]any
		expResult RuleResult
	}{
		{
			ruleNodes: Rule{
				Operator: And,
				Children: []Rule{
					{
						Operator: Eq,
						Field:    "user.firstName",
						Value:    "John",
					},
					{
						Operator: Eq,
						Field:    "user.lastName",
						Value:    "Doe",
					},
				},
			},
			inputData: map[string]any{
				"user": map[string]any{
					"firstName": "John",
					"lastName":  "Doe",
				},
			},
			expResult: RuleResult{
				Result: true,
			},
		},
		{
			ruleNodes: Rule{
				Operator: Eq, Field: "name", Value: "John",
			},
			inputData: map[string]any{"name": "John"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Neq, Field: "name", Value: "Johny",
			},
			inputData: map[string]any{"name": "John"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Gt, Field: "age", Value: 10,
			},
			inputData: map[string]any{"age": 11},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Lt, Field: "age", Value: 10,
			},
			inputData: map[string]any{"age": 9},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Gte, Field: "age", Value: 10,
			},
			inputData: map[string]any{"age": 10},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Lte, Field: "age", Value: 10,
			},
			inputData: map[string]any{"age": 10},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Between, Field: "age", Value: []any{10, 20},
			},
			inputData: map[string]any{"age": 15},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: In, Field: "role", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"role": "admin"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: In, Field: "role", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"role": "editor"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: NotIn, Field: "role", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"role": "editor"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: NotIn, Field: "role", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"role": "manager"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "roles", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"roles": []any{"editor", "admin"}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "roles", Value: []any{"admin", "manager"},
			},
			inputData: map[string]any{"roles": []any{"editor", "dev"}},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []int{1, 2},
			},
			inputData: map[string]any{"nums": []int{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []any{1, 2},
			},
			inputData: map[string]any{"nums": []int{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []int{1, 2},
			},
			inputData: map[string]any{"nums": []any{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]any{"secret": "some_password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]any{"secret": "some_password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: StartsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: StartsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: EndsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: EndsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]any{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]any{"secret": "peach"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]any{"secret": "pencil"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "abc"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "abcd"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "abcd"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "abc"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "ab"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]any{"secret": "abc"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsTrue, Field: "secret",
			},
			inputData: map[string]any{"secret": true},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsTrue, Field: "secret",
			},
			inputData: map[string]any{"secret": false},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsFalse, Field: "secret",
			},
			inputData: map[string]any{"secret": false},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsFalse, Field: "secret",
			},
			inputData: map[string]any{"secret": true},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Before, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]any{"startDate": parsedTimeData},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: After, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]any{"startDate": parsedTimeData},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]any{"startDate": parsedTimeExp.Add(3 * time.Hour)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]any{"startDate": parsedTimeExp.Add(-3 * time.Hour)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Before, Field: "startDate", Value: "thisYear+1y",
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: After, Field: "startDate", Value: "thisYear-1",
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: DateBetween, Field: "startDate", Value: []any{"thisYear", "thisYear+1y"},
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: YearEq, Field: "startDate", Value: relativeNow.Year(),
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: YearEq, Field: "startDate", Value: 2020,
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: MonthEq, Field: "startDate", Value: int(time.June),
			},
			inputData: map[string]any{"startDate": relativeTime},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: YearEq, Field: "startDate", Value: "thisYear-1",
			},
			inputData: map[string]any{"startDate": relativeTime.AddDate(-1, 0, 0)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: MonthEq, Field: "startDate", Value: "thisMonth",
			},
			inputData: map[string]any{"startDate": time.Date(relativeNow.Year(), relativeNow.Month(), 1, 12, 0, 0, 0, relativeNow.Location())},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]any{"startDate": time.Now().Add(-5 * time.Second)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]any{"startDate": time.Now().Add(-15 * time.Second)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]any{"startDate": time.Now().Add(5 * time.Second)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]any{"startDate": time.Now().Add(11 * time.Second)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsNull, Field: "someField",
			},
			inputData: map[string]any{"anotherField": "something"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNull, Field: "someField",
			},
			inputData: map[string]any{"someField": "something"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsNotNull, Field: "someField",
			},
			inputData: map[string]any{"someField": "something"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNotNull, Field: "someField",
			},
			inputData: map[string]any{"anotherField": "something"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsString, Field: "zipCode",
			},
			inputData: map[string]any{"zipCode": "1300"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNumber, Field: "age",
			},
			inputData: map[string]any{"age": uint32(20)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsBool, Field: "agreed",
			},
			inputData: map[string]any{"agreed": true},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsList, Field: "roles",
			},
			inputData: map[string]any{"roles": []int{1, 2, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsObject, Field: "address",
			},
			inputData: map[string]any{"address": map[string]any{}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsObject, Field: "address",
			},
			inputData: map[string]any{"address": struct{}{}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsDate, Field: "created_at",
			},
			inputData: map[string]any{"created_at": time.Now()},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Or,
				Children: []Rule{
					{Operator: Eq, Field: "firstName", Value: "John"},
					{Operator: Eq, Field: "lastName", Value: "Doe"},
				},
			},
			inputData: map[string]any{
				"firstName": "Mustermann",
				"lastName":  "Doe",
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Not,
				Children: []Rule{
					{Operator: Eq, Field: "firstName", Value: "John"},
				},
			},
			inputData: map[string]any{
				"firstName": "Mustermann",
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: And,
				Children: []Rule{
					{
						Operator: And,
						Field:    "user",
						Children: []Rule{
							{
								Field:    "user.name",
								Operator: LengthGt,
								Value:    2,
							},
							{
								Field:    "user.name",
								Operator: LengthLt,
								Value:    25,
							},
						},
					},
					{
						Field:    "user.age",
						Operator: Gte,
						Value:    21,
					},
					{
						Field:    "user.country",
						Operator: Eq,
						Value:    "DE",
					},
				},
			},
			inputData: map[string]any{
				"user": map[string]any{
					"name":    "Sam",
					"age":     25,
					"country": "DE",
				},
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Gt,
				Field:    "firstName",
				Value:    100,
			},
			inputData: map[string]any{
				"firstName": "Ali",
			},
			expResult: RuleResult{Result: false, Error: newError("invalid numerical value", "Ali")},
		},
		{
			ruleNodes: Rule{
				Operator: All,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 3},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: All,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 5},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Any,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 5},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Any,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 10},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: None,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 10},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: None,
				Field:    "roles",
				Value:    Rule{Operator: LengthGt, Value: 5},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: All,
				Field:    "roles",
				Value: Rule{
					Operator: And,
					Children: []Rule{
						{
							Operator: LengthGt,
							Value:    1,
						},
						{
							Operator: LengthLt,
							Value:    10,
						},
						{
							Operator: Neq,
							Value:    "admin",
						},
						{
							Operator: In,
							Value:    []string{"test", "roles", "longer", "than", "three"},
						},
					},
				},
			},
			inputData: map[string]any{
				"roles": []string{"test", "roles", "longer", "than", "three"},
			},
			expResult: RuleResult{Result: true},
		},
	}
)

func TestEvaluate(t *testing.T) {
	for _, testCase := range testData {
		t.Run(fmt.Sprintf(
			"%s_%v",
			testCase.ruleNodes.Operator,
			testCase.expResult.Result,
		), func(t *testing.T) {
			assert := assertion.New(t)
			result := Evaluate(
				testCase.ruleNodes,
				testCase.inputData,
				DefaultOptions().
					WithTiming().
					WithLogger(
						func(
							fieldName string,
							operator Operator,
							actual, expected any,
						) {
						}),
			)
			assert.Equal(testCase.expResult.Result, result.Result)
			assert.Equal(testCase.expResult.Error, result.Error)
			if testCase.expResult.Error != nil {
				assert.Equal(testCase.expResult.Error.Error(), result.Error.Error())
			}
		})
	}
}

func TestEvaluateWithCustomFunc(t *testing.T) {
	assert := assertion.New(t)

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

	ruleNode := Rule{
		Operator: Custom,
		Field:    "email",
		Value:    []any{"isEmail", "floating.tester@domain.ext"},
	}

	data := map[string]any{
		"email": "some.email@domain.ext",
	}

	result := Evaluate(ruleNode, data, DefaultOptions())
	assert.Equal(true, result.Result)
}

func BenchmarkEvaluate(b *testing.B) {
	ruleNodes := Rule{
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
	for i := 0; i < b.N; i++ {
		_ = Evaluate(ruleNodes, data, DefaultOptions())
	}
}
