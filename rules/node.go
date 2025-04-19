package rules

import "time"

type (
	Node struct {
		Operator Operator    `json:"operator"`
		Field    string      `json:"field,omitempty"`
		Value    interface{} `json:"value,omitempty"`
		Children []Node      `json:"children,omitempty"`
	}

	NodeEvaluation struct {
		Node      Node
		Result    bool
		Children  []NodeEvaluation
		TimeTaken time.Duration
	}
)
