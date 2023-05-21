package code

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/goccy/go-json"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
	"reflect"
	"strconv"
)

const (
	objectVariable  = "object"
	tengoModuleName = "brease"
)

type Script struct {
	compiled  *tengo.Compiled
	codeBlock string
}

func moduleMaps(names ...string) *tengo.ModuleMap {
	modules := stdlib.GetModuleMap(names...)
	modules.AddBuiltinModule(tengoModuleName, breaseModule)
	return modules
}

func conditionToScript(condition *pb.Condition) (code string) {
	jsonPath, ref := extractScriptBase(condition)
	parameterValue, isParamObj := parameterToScriptValue(condition)
	isReference := ref != nil

	paramCode := ""
	if isParamObj {
		paramCode = fmt.Sprintf("%s", parameterValue)
	} else {
		paramCode = fmt.Sprintf("%v", parameterValue)
	}
	fnWithParamLine := func(fnName string) string {
		if isReference {
			return fmt.Sprintf(`%s.%s(dref(jsonpath("%s", %s), "%s"), %s)`, tengoModuleName, fnName, ref.Src, objectVariable, ref.Dst, paramCode)
		}
		return fmt.Sprintf(`%s.%s(jsonpath("%s", %s), %s)`, tengoModuleName, fnName, jsonPath, objectVariable, paramCode)
	}

	switch condition.Type {
	case models.ConditionEmpty:
		return // nothind to do
	case models.ConditionHasValue:
		if isReference {
			code = fmt.Sprintf(`%s.hasValue(dref(jsonpath("%s", %s), "%s"))`, tengoModuleName, ref.Src, objectVariable, ref.Dst)
		} else {
			code = fmt.Sprintf(`%s.hasValue(jsonpath("%s", %s))`, tengoModuleName, jsonPath, objectVariable)
		}
	case models.ConditionEquals:
		if isReference {
			code = fmt.Sprintf(`dref(jsonpath("%s", %s), "%s") == %v`, ref.Src, objectVariable, ref.Dst, parameterValue)
		} else {
			code = fmt.Sprintf(`jsonpath("%s", %s) == %v`, jsonPath, objectVariable, parameterValue)
		}
	case models.ConditionDoesNotEqual:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("equals"))
	case models.ConditionHasPrefix:
		code = fnWithParamLine("hasPrefix")
	case models.ConditionDoesNotHavePrefix:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("hasPrefix"))
	case models.ConditionHasSuffix:
		code = fnWithParamLine("hasSuffix")
	case models.ConditionDoesNotHaveSuffix:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("hasSuffix"))
	case models.ConditionInList:
		code = fnWithParamLine("inList")
	case models.ConditionNotInList:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("inList"))
	case models.ConditionRegex:
		code = fnWithParamLine("regex")
	case models.ConditionNotRegex:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("regex"))
	case models.ConditionSome:
		code = fnWithParamLine("some")
	case models.ConditionAll:
		code = fnWithParamLine("all")
	case models.ConditionNone:
		code = fnWithParamLine("none")
	}
	return
}

func extractScriptBase(condition *pb.Condition) (string, *pb.ConditionBaseRef) {
	var ref *pb.ConditionBaseRef
	jsonPath := ""
	if rf, ok := condition.Base.(*pb.Condition_Ref); ok {
		ref = rf.Ref
		jsonPath = rf.Ref.Src
	} else if base, ok := condition.Base.(*pb.Condition_Key); ok {
		jsonPath = base.Key
	}
	return jsonPath, ref
}

func parameterToScriptValue(condition *pb.Condition) (parameterValue any, isObj bool) {
	switch condition.Parameter.(type) {
	case *pb.Condition_BoolValue:
		parameterValue = condition.GetBoolValue()
	case *pb.Condition_IntValue:
		parameterValue = condition.GetIntValue()
	case *pb.Condition_StringValue:
		// return arr/obj in tengo format
		var v interface{}
		val := condition.GetStringValue()
		if err := json.Unmarshal([]byte(val), &v); err == nil {
			rv := reflect.ValueOf(v)
			kind := rv.Kind()
			if kind == reflect.Array || kind == reflect.Slice || kind == reflect.Map {
				parameterValue = val
				isObj = true
				return
			}
		}
		parameterValue = strconv.Quote(condition.GetStringValue())
	case *pb.Condition_ByteValue:
		parameterValue = condition.GetByteValue()
	default:
		parameterValue = nil
	}
	return
}
