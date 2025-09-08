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
	last10Sec, _      = time.ParseDuration("10s")
	testData          = []struct {
		ruleNodes Rule
		inputData map[string]interface{}
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
			inputData: map[string]interface{}{
				"user": map[string]interface{}{
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
			inputData: map[string]interface{}{"name": "John"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Neq, Field: "name", Value: "Johny",
			},
			inputData: map[string]interface{}{"name": "John"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Gt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 11},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Lt, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 9},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Gte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Lte, Field: "age", Value: 10,
			},
			inputData: map[string]interface{}{"age": 10},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Between, Field: "age", Value: []interface{}{10, 20},
			},
			inputData: map[string]interface{}{"age": 15},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "admin"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: In, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "editor"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: NotIn, Field: "role", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"role": "manager"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "roles", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"roles": []interface{}{"editor", "admin"}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "roles", Value: []interface{}{"admin", "manager"},
			},
			inputData: map[string]interface{}{"roles": []interface{}{"editor", "dev"}},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []int{1, 2},
			},
			inputData: map[string]interface{}{"nums": []int{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []interface{}{1, 2},
			},
			inputData: map[string]interface{}{"nums": []int{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: AnyIn, Field: "nums", Value: []int{1, 2},
			},
			inputData: map[string]interface{}{"nums": []interface{}{1, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Contains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some_password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: NotContains, Field: "secret", Value: "%",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: StartsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: StartsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: EndsWith, Field: "secret", Value: "password",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: EndsWith, Field: "secret", Value: "some",
			},
			inputData: map[string]interface{}{"secret": "some%password"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "peach"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: Matches, Field: "secret", Value: "p([a-z]+)ch",
			},
			inputData: map[string]interface{}{"secret": "pencil"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthEq, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abcd"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthGt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "ab"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: LengthLt, Field: "secret", Value: "3",
			},
			inputData: map[string]interface{}{"secret": "abc"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsTrue, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": false},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsFalse, Field: "secret",
			},
			inputData: map[string]interface{}{"secret": true},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: Before, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: After, Field: "startDate", Value: parsedTimeExp,
			},
			inputData: map[string]interface{}{"startDate": parsedTimeData},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(3 * time.Hour)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: DateBetween, Field: "startDate", Value: []time.Time{parsedTimeExp, parsedTimeExp.Add(24 * time.Hour)},
			},
			inputData: map[string]interface{}{"startDate": parsedTimeExp.Add(-3 * time.Hour)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-5 * time.Second)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: WithinLast, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(-15 * time.Second)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(5 * time.Second)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: WithinNext, Field: "startDate", Value: last10Sec,
			},
			inputData: map[string]interface{}{"startDate": time.Now().Add(11 * time.Second)},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"someField": "something"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNotNull, Field: "someField",
			},
			inputData: map[string]interface{}{"anotherField": "something"},
			expResult: RuleResult{Result: false},
		},
		{
			ruleNodes: Rule{
				Operator: IsString, Field: "zipCode",
			},
			inputData: map[string]interface{}{"zipCode": "1300"},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsNumber, Field: "age",
			},
			inputData: map[string]interface{}{"age": uint32(20)},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsBool, Field: "agreed",
			},
			inputData: map[string]interface{}{"agreed": true},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsList, Field: "roles",
			},
			inputData: map[string]interface{}{"roles": []int{1, 2, 3}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": map[string]interface{}{}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsObject, Field: "address",
			},
			inputData: map[string]interface{}{"address": struct{}{}},
			expResult: RuleResult{Result: true},
		},
		{
			ruleNodes: Rule{
				Operator: IsDate, Field: "created_at",
			},
			inputData: map[string]interface{}{"created_at": time.Now()},
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
				"user": map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
			inputData: map[string]interface{}{
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
							actual, expected interface{},
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

	RegisterFunc("isEmail", func(args ...interface{}) (bool, any, error) {
		dataEmail, ok := args[0].(string)
		if !ok {
			return false, nil, nil
		}
		passedEmail, ok := args[1].(string)
		if !ok {
			return false, nil, nil
		}

		return strings.Contains(dataEmail, "@") &&
			strings.Contains(dataEmail, ".") &&
			passedEmail == "floating.tester@domain.ext" &&
			dataEmail == "some.email@domain.ext", nil, nil
	})

	ruleNode := Rule{
		Operator: Custom,
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
