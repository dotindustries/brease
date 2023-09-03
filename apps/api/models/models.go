package models

import (
	"go.uber.org/zap/zapcore"
)

type Target struct {
	Kind   string `json:"kind" validate:"required" example:"customTargetKind"`
	Target string `json:"target" validate:"required" example:"$.prop" example:"propKey" example:"target_01h89qgxe5e7wregw6gb94d5p6"`
	// TODO: Should be anything
	// The target value to be set (it is the json serialized representation of the value
	Value string `json:"value,omitempty" example:"ZXhhbXBsZQ=="`
}

func (Target) TypeName() string { return "Target" }

func (e Target) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("target", e.Target)
	enc.AddString("kind", e.Kind)
	enc.AddString("value", e.Value)
	return nil
}

type Action struct {
	Action string `json:"action" validate:"required" example:"$set"`
	Target Target `json:"target" validate:"required"`
}

type Rule struct {
	ID          string   `json:"id" validate:"required" example:"rule_01h89qfdhbejtb3jwqq1gazbm5"`
	Description string   `json:"description,omitempty" example:"Rule short description"`
	Actions     []Action `json:"actions" validate:"required"`
	// Ugly workaround as base64 protobuf until https://github.com/wI2L/fizz/issues/80 is resolved
	// A variadic condition expression in a binary format.
	//  Expression:
	//    type: object
	//    oneOf:
	//      - $ref: '#/components/schemas/And'
	//      - $ref: '#/components/schemas/Or'
	//      - $ref: '#/components/schemas/Condition'
	Expression map[string]interface{} `json:"expression" validate:"required" example:""`
}

func (*Rule) TypeName() string { return "Rule" }

type VersionedRule struct {
	Rule
	Version int64 `json:"version" validate:"required" example:"1"`
}

func (*VersionedRule) TypeName() string { return "VersionedRule" }

type EvaluationResult struct {
	Action string `json:"action" validate:"required" example:"$set"`
	Target Target `json:"target" validate:"required"`
	By     string `json:"by" example:"rule_01h89qfdhbejtb3jwqq1gazbm5"`
}

func (EvaluationResult) TypeName() string { return "EvaluationResult" }

func (e EvaluationResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	_ = enc.AddObject("target", e.Target)
	enc.AddString("action", e.Action)
	enc.AddString("by", e.By)
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
