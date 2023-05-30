package code

import (
	"context"
	"fmt"
	jpath "github.com/PaesslerAG/jsonpath"
	"github.com/d5/tengo/v2"
	"go.dot.industries/brease/rref"
	"log"
	"regexp"
	"strings"
)

var (
	codeHeader = `brease := import("brease")

// aliases
jsonpath := brease.jsonpath
dref := brease.dref

results := []

action := func(action, targetType, targetID, value) {
	results = append(results, { targetID: targetID, targetType: targetType, action: action, value: value })
}

`
	breaseModule = map[string]tengo.Object{
		"hasValue": &tengo.UserFunction{
			Name:  "hasValue",
			Value: hasValue,
		},
		"hasPrefix": &tengo.UserFunction{
			Name:  "hasPrefix",
			Value: hasPrefix,
		},
		"hasSuffix": &tengo.UserFunction{
			Name:  "hasSuffix",
			Value: hasSuffix,
		},
		"inList": &tengo.UserFunction{
			Name:  "inList",
			Value: inList,
		},
		"some": &tengo.UserFunction{
			Name:  "some",
			Value: some,
		},
		"all": &tengo.UserFunction{
			Name:  "all",
			Value: all,
		},
		"none": &tengo.UserFunction{
			Name:  "none",
			Value: none,
		},
		"regex": &tengo.UserFunction{
			Name:  "equals",
			Value: regex,
		},
		"jsonpath": &tengo.UserFunction{
			Name:  "jsonpath",
			Value: jsonpath,
		},
		"dref": &tengo.UserFunction{
			Name:  "dref",
			Value: dref,
		},
	}
)

func hasValue(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}

	// the object
	val := tengo.ToInterface(args[0])

	ret = tengo.TrueValue
	if val == nil {
		ret = tengo.FalseValue
	}

	log.Printf("hasValue: %v\nvalue: %v", ret, val)
	return
}

func hasPrefix(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	val := tengo.ToInterface(args[0])
	prefix := ""
	switch o := args[1].(type) {
	case *tengo.String:
		prefix = o.Value
	default:
		prefix = o.String()
	}

	stringifiedVal := fmt.Sprintf("%v", val)
	if strings.HasPrefix(stringifiedVal, prefix) {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}
	log.Printf("hasPrefix: %v\nvalue: %v\ntest: %v", ret, stringifiedVal, prefix)
	return
}

func hasSuffix(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	val := tengo.ToInterface(args[0])
	suffix := ""
	switch o := args[1].(type) {
	case *tengo.String:
		suffix = o.Value
	default:
		suffix = o.String()
	}

	stringifiedVal := fmt.Sprintf("%v", val)
	if strings.HasSuffix(stringifiedVal, suffix) {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}
	log.Printf("hasSuffix: %v\nvalue: %v\ntest:%v\n", ret, stringifiedVal, suffix)
	return
}

func inList(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	val := args[0]
	var arr []tengo.Object
	switch o := args[1].(type) {
	case *tengo.Array:
		arr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "array",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}

	ret = tengo.FalseValue
	for _, o := range arr {
		log.Printf("comparing objects\no type: %s\no val: %v\nval type: %s\nval: %v", o.TypeName(), o, val.TypeName(), val)
		if o.Equals(val) {
			ret = tengo.TrueValue
			break
		}
	}

	log.Printf("inList: %v\nvalue: %v\ntest: %v", ret, val, arr)
	return
}

// some is a function helper which evaluates to true or false whether the value array contains at least one of the elements of the provided array
func some(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	var arr []tengo.Object
	switch o := args[0].(type) {
	case *tengo.Array:
		arr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	var inArr []tengo.Object
	switch o := args[1].(type) {
	case *tengo.Array:
		inArr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "inArray",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	set := make(map[string]bool)
	for _, o := range arr {
		set[o.String()] = true
	}
	ret = tengo.FalseValue
	for _, o := range inArr {
		if set[o.String()] {
			ret = tengo.TrueValue // first match wins
			break
		}
	}

	log.Printf("some: %v\nvalue: %v\ntest: %v\nsearchSet: %v\n", ret, arr, inArr, set)
	return
}

// all is a function helper which evaluates to true or false whether the value array contains every element of the provided array
func all(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	var arr []tengo.Object
	switch o := args[0].(type) {
	case *tengo.Array:
		arr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	var allOf []tengo.Object
	switch o := args[1].(type) {
	case *tengo.Array:
		allOf = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "inArray",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	set := make(map[string]bool)
	for _, o := range arr {
		set[o.String()] = true
	}
	ret = tengo.FalseValue
	c := 0
	for _, o := range allOf {
		if set[o.String()] {
			c++
		}
	}
	if len(set) == c {
		ret = tengo.TrueValue
	}

	log.Printf("all: %v\nvalue: %v\ntest: %v", ret, arr, allOf)
	return
}

// none is a function helper which evaluates to true or false whether the value array does not contain any element of the provided array
func none(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}
	var arr []tengo.Object
	switch o := args[0].(type) {
	case *tengo.Array:
		arr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	var noneOf []tengo.Object
	switch o := args[1].(type) {
	case *tengo.Array:
		noneOf = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "inArray",
			Expected: "array",
			Found:    o.TypeName(),
		}
	}
	set := make(map[string]bool)
	for _, o := range arr {
		set[o.String()] = true
	}
	ret = tengo.TrueValue
	for _, o := range noneOf {
		if set[o.String()] {
			ret = tengo.FalseValue // first match fails
			break
		}
	}

	log.Printf("none: %v\nvalue: %v\ntest: %v", ret, arr, noneOf)
	return
}

func regex(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	val := tengo.ToInterface(args[0])
	expr := ""
	switch o := args[1].(type) {
	case *tengo.String:
		expr = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "expression",
			Expected: "string",
			Found:    o.TypeName(),
		}
	}

	rgx, err := regexp.Compile(expr)
	if err != nil {
		return nil, fmt.Errorf("cannot compile regex: %w", err)
	}
	stringifiedVal := fmt.Sprintf("%v", val)
	if rgx.MatchString(stringifiedVal) {
		ret = tengo.TrueValue
	} else {
		ret = tengo.FalseValue
	}
	log.Printf("regex: %v\nvalue: %v\ntest: %v", ret, stringifiedVal, expr)
	return
}

func jsonpath(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	path := ""
	switch o := args[0].(type) {
	case *tengo.String:
		path = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    o.TypeName(),
		}
	}

	v := tengo.ToInterface(args[1])

	value, err := jpath.Get(path, v)
	if err != nil {
		return nil, err
	}

	return tengo.FromInterface(value)
}

func dref(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) != 2 {
		return nil, tengo.ErrWrongNumArguments
	}

	input := tengo.ToInterface(args[0])

	dst := ""
	switch o := args[1].(type) {
	case *tengo.String:
		dst = o.Value
	default:
		return nil, tengo.ErrInvalidArgumentType{
			Name:     "path",
			Expected: "string",
			Found:    o.TypeName(),
		}
	}

	if rref.IsConfigured() {
		log.Println("TODO: lookup reference value", dst, "based on input", input)
		_ = rref.LookupReferenceValue(context.Background(), nil)
	} else {
		log.Println("WARN: Lookup for reference valued expressions is off. The dref is not configured")
	}

	return nil, fmt.Errorf("dref not yet implemented")
}
