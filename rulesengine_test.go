package rulesengine

import (
	"github.com/goglue/rulesengine/rules"
	assertion "github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestEvaluate(t *testing.T) {
	assert := assertion.New(t)

	ruleNode := rules.Node{
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
	}

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"firstName": "John",
			"lastName":  "Doe",
		},
	}

	result := Evaluate(ruleNode, data, DefaultOptions().WithTiming())
	assert.Equal(true, result.Result)
	assert.NotEmpty(result.TimeTaken)
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
