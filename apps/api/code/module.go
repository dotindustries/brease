package code

import (
	"fmt"

	"github.com/d5/tengo/v2"
)

var (
	codeHeader = `brease := import("brease")

results := []

action := func(action, targetType, targetID, value) {
	results = append(results, { targetID: targetID, targetType: targetType, action: action, value: value })
}

`
	breaseModule = map[string]tengo.Object{
		"hasValue": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "hasValue",
			Value:      hasValue,
		},
		"hasPrefix": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "hasPrefix",
			Value:      hasPrefix,
		},
		"hasSuffix": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "hasSuffix",
			Value:      hasSuffix,
		},
		"equals": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "equals",
			Value:      equals,
		},
		"inList": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "inList",
			Value:      inList,
		},
		"regex": &tengo.UserFunction{
			ObjectImpl: tengo.ObjectImpl{},
			Name:       "equals",
			Value:      regex,
		},
	}
)

func hasValue(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}

func hasPrefix(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}

func hasSuffix(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}

func equals(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}

func inList(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}

func regex(args ...tengo.Object) (ret tengo.Object, err error) {
	return nil, fmt.Errorf("not yet implemented")
}
