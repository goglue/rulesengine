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
		ruleNodes rules.Node
		inputData map[string]interface{}
		expResult rules.NodeEvaluation
	}{
		{
			ruleNodes: rules.Node{
				Operator: rules.And,
				Children: []rules.Node{
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
			expResult: rules.NodeEvaluation{
				Result: true,
			},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Eq, Field: "name", Value: "John",
			},
			inputData: map[string]interface{}{"name": "John"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Neq, Field: "name", Value: "Johny",
			},
			inputData: map[string]interface{}{"name": "John"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Gt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 11},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Lt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 9},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Gte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Lte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Between, Field: "age", Value: []interface{}{10, 20},
			},
			inputData: map[string]interface{}{"age": 15},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "admin"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "manager"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.StartsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.StartsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.EndsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.EndsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "peach"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "pencil"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "ab"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Before, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.After, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(3 * time.Hour)},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(-3 * time.Hour)},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-5 * time.Second)},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-15 * time.Second)},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(5 * time.Second)},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(11 * time.Second)},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: rules.NodeEvaluation{Result: false},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsString, Field: "zipCode",
			},
			inputData: map[string]interface{}{"zipCode": "1300"},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsNumber, Field: "age",
			},
			inputData: map[string]interface{}{"age": uint32(20)},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsBool, Field: "agreed",
			},
			inputData: map[string]interface{}{"agreed": true},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsList, Field: "roles",
			},
			inputData: map[string]interface{}{"roles": []int{1, 2, 3}},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": map[string]interface{}{}},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": struct{}{}},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.IsDate, Field: "created_at",
			},
			inputData: map[string]interface{}{"created_at": time.Now()},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Or,
				Children: []rules.Node{
					{Operator: rules.Eq, Field: "firstName", Value: "John"},
					{Operator: rules.Eq, Field: "lastName", Value: "Doe"},
				},
			},
			inputData: map[string]interface{}{
				"firstName": "Mustermann",
				"lastName":  "Doe",
			},
			expResult: rules.NodeEvaluation{Result: true},
		},
		{
			ruleNodes: rules.Node{
				Operator: rules.Not,
				Children: []rules.Node{
					{Operator: rules.Eq, Field: "firstName", Value: "John"},
				},
			},
			inputData: map[string]interface{}{
				"firstName": "Mustermann",
			},
			expResult: rules.NodeEvaluation{Result: true},
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

	rules.RegisterFunc("isEmail", func(args ...interface{}) bool {
		dataEmail, ok := args[0].(string)
		if !ok {
			return false
		}
		passedEmail, ok := args[1].(string)
		if !ok {
			return false
		}

		return strings.Contains(dataEmail, "@") &&
			strings.Contains(dataEmail, ".") &&
			passedEmail == "floating.tester@domain.ext" &&
			dataEmail == "some.email@domain.ext"
	})

	ruleNode := rules.Node{
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
