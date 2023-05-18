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
