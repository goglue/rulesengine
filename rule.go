package rulesengine

import "time"

type (
	Rule struct {
		Operator Operator    `json:"operator"`
		Field    string      `json:"field,omitempty"`
		Value    interface{} `json:"value,omitempty"`
		Children []Rule      `json:"children,omitempty"`
	}

	RuleResult struct {
		Node      Rule
		Result    bool
		Children  []RuleResult
		Mismatch  interface{}
		TimeTaken time.Duration
	}
)
