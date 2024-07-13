package models

import (
	"go.jetify.com/typeid"
)

type RulePrefix struct{}

func (RulePrefix) Prefix() string { return "rule" }

type RuleID struct{ typeid.TypeID[RulePrefix] }

func NewRuleID() RuleID {
	tid, _ := typeid.New[RuleID]()
	return tid
}
