package rulesengine

import (
	"github.com/goglue/rulesengine/rules"
	assertion "github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

var (
	parsedTimeData, _ = time.Parse("2006-01-02", "2014-12-01")
	parsedTimeExp, _  = time.Parse("2006-01-02", "2015-01-01")
	last10Sec, _      = time.ParseDuration("10s")
	testData          = []struct {
		ruleNodes rules.Rule
		inputData map[string]interface{}
		expResult rules.RuleResult
	}{
		{
			ruleNodes: rules.Rule{
				Operator: rules.And,
				Children: []rules.Rule{
					{
						Operator: rules.Eq,
						Field:    "user.firstName",
						Value:    "John",
					},
					{
						Operator: rules.Eq,
						Field:    "user.lastName",
						Value:    "Doe",
					},
				},
			},
			inputData: map[string]interface{}{
				"user": map[string]interface{}{
					"firstName": "John",
					"lastName":  "Doe",
				},
			},
			expResult: rules.RuleResult{
				Result: true,
			},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Eq, Field: "name", Value: "John",
			},
			inputData: map[string]interface{}{"name": "John"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Neq, Field: "name", Value: "Johny",
			},
			inputData: map[string]interface{}{"name": "John"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Gt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 11},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Lt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 9},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Gte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Lte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Between, Field: "age", Value: []interface{}{10, 20},
			},
			inputData: map[string]interface{}{"age": 15},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "admin"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "manager"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.StartsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.StartsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.EndsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.EndsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "peach"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "pencil"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "ab"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Before, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.After, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(3 * time.Hour)},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(-3 * time.Hour)},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-5 * time.Second)},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-15 * time.Second)},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(5 * time.Second)},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(11 * time.Second)},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: rules.RuleResult{Result: false},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsString, Field: "zipCode",
			},
			inputData: map[string]interface{}{"zipCode": "1300"},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsNumber, Field: "age",
			},
			inputData: map[string]interface{}{"age": uint32(20)},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsBool, Field: "agreed",
			},
			inputData: map[string]interface{}{"agreed": true},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsList, Field: "roles",
			},
			inputData: map[string]interface{}{"roles": []int{1, 2, 3}},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": map[string]interface{}{}},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": struct{}{}},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.IsDate, Field: "created_at",
			},
			inputData: map[string]interface{}{"created_at": time.Now()},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Or,
				Children: []rules.Rule{
					{Operator: rules.Eq, Field: "firstName", Value: "John"},
					{Operator: rules.Eq, Field: "lastName", Value: "Doe"},
				},
			},
			inputData: map[string]interface{}{
				"firstName": "Mustermann",
				"lastName":  "Doe",
			},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.Not,
				Children: []rules.Rule{
					{Operator: rules.Eq, Field: "firstName", Value: "John"},
				},
			},
			inputData: map[string]interface{}{
				"firstName": "Mustermann",
			},
			expResult: rules.RuleResult{Result: true},
		},
		{
			ruleNodes: rules.Rule{
				Operator: rules.And,
				Children: []rules.Rule{
					{
						Operator: rules.And,
						Field:    "user",
						Children: []rules.Rule{
							{
								Field:    "user.name",
								Operator: rules.LengthGt,
								Value:    2,
							},
							{
								Field:    "user.name",
								Operator: rules.LengthLt,
								Value:    25,
							},
						},
					},
					{
						Field:    "user.age",
						Operator: rules.Gte,
						Value:    21,
					},
					{
						Field:    "user.country",
						Operator: rules.Eq,
						Value:    "DE",
					},
				},
			},
			inputData: map[string]interface{}{
				"user": map[string]interface{}{
					"name":    "Sam",
					"age":     25,
					"country": "DE",
				},
			},
			expResult: rules.RuleResult{Result: true},
		},
	}
)

func TestEvaluate(t *testing.T) {
	for _, testCase := range testData {
		t.Run("", func(t *testing.T) {
			assert := assertion.New(t)
			result := Evaluate(
				testCase.ruleNodes,
				testCase.inputData,
				DefaultOptions().
					WithTiming().
					WithLogger(
						func(
							fieldName string,
							operator rules.Operator,
							actual, expected interface{},
						) {

						}),
			)
			assert.Equal(testCase.expResult.Result, result.Result)
		})
	}
}

func TestEvaluateWithCustomFunc(t *testing.T) {
	assert := assertion.New(t)

	rules.RegisterFunc("isEmail", func(args ...interface{}) (bool, any) {
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

	ruleNode := rules.Rule{
		Operator: rules.Custom,
		Field:    "email",
		Value:    []interface{}{"isEmail", "floating.tester@domain.ext"},
	}

	data := map[string]interface{}{
		"email": "some.email@domain.ext",
	}

	result := Evaluate(ruleNode, data, DefaultOptions())
	assert.Equal(true, result.Result)
}

func BenchmarkEvaluate(b *testing.B) {
	ruleNodes := rules.Rule{
		Operator: rules.And,
		Children: []rules.Rule{
			{Operator: rules.IsNumber, Field: "user.age"},
			{Operator: rules.Gte, Field: "user.age", Value: 25},
			{Operator: rules.Matches, Field: "jobTitle", Value: "s([a-z]+)re"},
			{Operator: rules.IsObject, Field: "user.address"},
			{Operator: rules.Eq, Field: "user.address.zipCode", Value: 5},
			{Operator: rules.IsNotNull, Field: "user.address.streetName"},
			{Operator: rules.LengthGt, Field: "user.address.streetName", Value: 5},
			{Operator: rules.LengthGt, Field: "user.firstName", Value: 2},
			{Operator: rules.LengthGt, Field: "user.lastName", Value: 2},
		},
	}
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
			"age":       25,
			"jobTitle":  "software",
			"address": map[string]interface{}{
				"streetName": "Johannisstraße",
				"zipCode":    "13088",
			},
		},
	}
	for i := 0; i < b.N; i++ {
		_ = Evaluate(ruleNodes, data, DefaultOptions())
	}
}
