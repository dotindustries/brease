package api

type Rule struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	// The action to be reported for the Target with its TargetType
	Action     string `json:"action"`
	TargetType string `json:"targetType"`
	Target     string `json:"target"`
	Parameter  []byte `json:"targetValue,omitempty"`
	// A variadic condition expression
	Expression *Expression `json:"expression"`
}

type Expression struct {
}
