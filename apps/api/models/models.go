package models

import (
	"go.uber.org/zap/zapcore"
)

type Target struct {
	Type   string `json:"type" validate:"required"`
	Target string `json:"target" validate:"required"`
	// TODO: Should be anything
	// The target value to be set (it is the json serialized representation of the value
	Value string `json:"value,omitempty"`
}

type Rule struct {
	ID          string `json:"id" validate:"required"`
	Description string `json:"description,omitempty"`
	// The action to be reported for the Target
	Action string `json:"action" validate:"required"`
	Target Target `json:"target" validate:"required"`

	// Ugly workaround as base64 protobuf until https://github.com/wI2L/fizz/issues/80 is resolved
	// A variadic condition expression in a binary format.
	//  Expression:
	//    type: object
	//    oneOf:
	//      - $ref: '#/components/schemas/And'
	//      - $ref: '#/components/schemas/Or'
	//      - $ref: '#/components/schemas/Condition'
	Expression map[string]interface{} `json:"expression" validate:"required"`
}

type EvaluationResult struct {
	TargetID   string `json:"targetID"`
	TargetType string `json:"actionTargetType"`
	Action     string `json:"action"`
	Value      string `json:"value"`
}

func (e EvaluationResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("targetID", e.TargetID)
	enc.AddString("targetType", e.TargetType)
	enc.AddString("action", e.Action)
	enc.AddString("value", e.Value)

	return nil
}

type ConditionType = string

const (
	ConditionEmpty             ConditionType = "e"
	ConditionHasValue          ConditionType = "hv"
	ConditionEquals            ConditionType = "eq"
	ConditionDoesNotEqual      ConditionType = "neq"
	ConditionHasPrefix         ConditionType = "px"
	ConditionDoesNotHavePrefix ConditionType = "npx"
	ConditionHasSuffix         ConditionType = "sx"
	ConditionDoesNotHaveSuffix ConditionType = "nsx"
	ConditionInList            ConditionType = "in"
	ConditionNotInList         ConditionType = "nin"
	ConditionSome              ConditionType = "some"
	ConditionAll               ConditionType = "all"
	ConditionNone              ConditionType = "none"
	ConditionRegex             ConditionType = "rgx"
	ConditionNotRegex          ConditionType = "nrgx"
)
