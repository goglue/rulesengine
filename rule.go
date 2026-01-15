package rulesengine

import "time"

type (
	Rule struct {
		// Operator attribute is the operator to be used for the evaluation
		// process, check [Operator] constants.
		Operator Operator `json:"operator"`
		// Field attribute is the path for the variable that needs to be
		// evaluated, the path can be defined in a JSON path `path.to.variable`.
		Field string `json:"field,omitempty"`
		// Value attribute is complementary to the [Operator], it can be a value
		// to be compared against.
		Value any `json:"value,omitempty"`
		// Children attribute is the nested (if needed) set of rules, in case of
		// len(Children) > 0, the [Operator] can only be logic: [And],[Or],[Not].
		Children []Rule `json:"children,omitempty"`
	}

	RuleResult struct {
		// Rule attribute references the rule the [RuleResult] belongs to.
		Rule Rule `json:"rule"`
		// Result attribute is a boolean indicator whether the rule has passed
		// or not.
		Result bool `json:"result"`
		// IsEmpty attribute indicates whether there was a value to compare or
		// not, e.g. nil pointer
		IsEmpty bool `json:"IsEmpty,omitempty"`
		// Children attribute is the nested results of the nested rules.
		Children []RuleResult `json:"children,omitempty"`
		// Input attribute holds the original value as given by the user.
		Input any `json:"input,omitempty"`
		// TimeTaken is a debugging attribute and holds the duration of the
		// rule evaluation.
		TimeTaken time.Duration `json:"timeTaken,omitempty"`
		// Error attribute contains an error occurred during the evaluation of a
		// rule, this can be any of the errors defined in the library. The error
		// entails that the evaluation did not finish successfully. The
		// underlying type is [Error].
		Error error `json:"error,omitempty"`
	}
)
