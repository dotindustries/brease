package api

type Rule struct {
	ID          string  `json:"id" validate:"required"`
	Description *string `json:"description,omitempty"`
	// The action to be reported for the Target with its TargetType
	Action     string `json:"action" validate:"required"`
	TargetType string `json:"targetType" validate:"required"`
	Target     string `json:"target" validate:"required"`
	Parameter  []byte `json:"targetValue,omitempty"`
	// A variadic condition expression
	Expression *Expression `json:"expression" validate:"required"`
}

type EvaluationResult struct {
	TargetID   string `json:"targetID"`
	TargetType string `json:"actionTargetType"`
	Action     string `json:"action"`
	value      string `json:"Value"`
}

type Expression struct {
}
