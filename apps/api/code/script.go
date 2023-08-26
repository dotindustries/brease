package code

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"go.dot.industries/brease/models"
	"go.dot.industries/brease/pb"
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
	isReference := ref != nil

	paramCode := string(condition.Value)
	fnWithParamLine := func(fnName string) string {
		if isReference {
			return fmt.Sprintf(`%s.%s(dref(jsonpath("%s", %s), "%s"), %s)`, tengoModuleName, fnName, ref.Src, objectVariable, ref.Dst, paramCode)
		}
		return fmt.Sprintf(`%s.%s(jsonpath("%s", %s), %s)`, tengoModuleName, fnName, jsonPath, objectVariable, paramCode)
	}

	switch condition.Kind {
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
			code = fmt.Sprintf(`dref(jsonpath("%s", %s), "%s") == %s`, ref.Src, objectVariable, ref.Dst, paramCode)
		} else {
			code = fmt.Sprintf(`jsonpath("%s", %s) == %s`, jsonPath, objectVariable, paramCode)
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
