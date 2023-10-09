package models

import (
	typeid "go.jetpack.io/typeid/typed"
)

type rulePrefix struct{}

func (rulePrefix) Type() string { return "rule" }

type RuleID struct{ typeid.TypeID[rulePrefix] }

func NewRuleID() typeid.TypeID[RuleID] {
	tid, _ := typeid.New[RuleID]()
	return tid
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
