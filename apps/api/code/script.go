package code

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"encoding/base64"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
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

func conditionToScript(condition *rulev1.Condition) (code string) {
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
	case rulev1.ConditionKind_e:
		return // nothind to do
	case rulev1.ConditionKind_hv:
		if isReference {
			code = fmt.Sprintf(`%s.hasValue(dref(jsonpath("%s", %s), "%s"))`, tengoModuleName, ref.Src, objectVariable, ref.Dst)
		} else {
			code = fmt.Sprintf(`%s.hasValue(jsonpath("%s", %s))`, tengoModuleName, jsonPath, objectVariable)
		}
	case rulev1.ConditionKind_eq:
		if isReference {
			code = fmt.Sprintf(`dref(jsonpath("%s", %s), "%s") == %s`, ref.Src, objectVariable, ref.Dst, paramCode)
		} else {
			code = fmt.Sprintf(`jsonpath("%s", %s) == %s`, jsonPath, objectVariable, paramCode)
		}
	case rulev1.ConditionKind_neq:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("equals"))
	case rulev1.ConditionKind_px:
		code = fnWithParamLine("hasPrefix")
	case rulev1.ConditionKind_npx:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("hasPrefix"))
	case rulev1.ConditionKind_sx:
		code = fnWithParamLine("hasSuffix")
	case rulev1.ConditionKind_nsx:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("hasSuffix"))
	case rulev1.ConditionKind_in:
		code = fnWithParamLine("inList")
	case rulev1.ConditionKind_nin:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("inList"))
	case rulev1.ConditionKind_rgx:
		code = fnWithParamLine("regex")
	case rulev1.ConditionKind_nrgx:
		code = fmt.Sprintf(`!%s`, fnWithParamLine("regex"))
	case rulev1.ConditionKind_some:
		code = fnWithParamLine("some")
	case rulev1.ConditionKind_all:
		code = fnWithParamLine("all")
	case rulev1.ConditionKind_cel:
		// base64 decode the paramCode
		bts, err := base64.StdEncoding.DecodeString(paramCode)
		if err != nil {
			panic(err)
		}
		code = fmt.Sprintf(`%s.cel(%s, %s)`, tengoModuleName, objectVariable, string(bts))
	case rulev1.ConditionKind_none:
		code = fnWithParamLine("none")
	}
	return
}

func extractScriptBase(condition *rulev1.Condition) (jsonPath string, ref *rulev1.ConditionBaseRef) {
	if rf, ok := condition.Base.(*rulev1.Condition_Ref); ok {
		ref = rf.Ref
		jsonPath = rf.Ref.Src
	} else if base, ok := condition.Base.(*rulev1.Condition_Key); ok {
		jsonPath = base.Key
	}
	return jsonPath, ref
}
