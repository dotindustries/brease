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
	compiled *tengo.Compiled
}

func moduleMaps(names ...string) *tengo.ModuleMap {
	modules := stdlib.GetModuleMap(names...)
	modules.AddBuiltinModule(tengoModuleName, breaseModule)
	return modules
}

func conditionToScript(condition *pb.Condition) (code string) {
	isReference := false
	if _, ok := condition.Base.(*pb.Condition_Ref); ok {
		isReference = true
	}

	var parameterValue any
	switch condition.Parameter.(type) {
	case *pb.Condition_BoolValue:
		parameterValue = condition.GetBoolValue()
	case *pb.Condition_IntValue:
		parameterValue = condition.GetIntValue()
	case *pb.Condition_StringValue:
		// TODO check if parameter is a JSON array or a JSON map and return accordingly
		parameterValue = "\"" + condition.GetStringValue() + "\""
	default:
		parameterValue = nil
	}

	switch condition.Type {
	case models.ConditionEmpty:
		return // nothind to do
	case models.ConditionHasValue:
		code = fmt.Sprintf(`%s.hasValue(%s, "%s", %t)`, tengoModuleName, objectVariable, condition.Base, isReference)
	case models.ConditionEquals:
		code = fmt.Sprintf(`%s.equals(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionDoesNotEqual:
		code = fmt.Sprintf(`!%s.equals(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionHasPrefix:
		code = fmt.Sprintf(`%s.hasPrefix(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionDoesNotHavePrefix:
		code = fmt.Sprintf(`!%s.hasPrefix(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionHasSuffix:
		code = fmt.Sprintf(`%s.hasSuffix(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionDoesNotHaveSuffix:
		code = fmt.Sprintf(`!%s.hasSuffix(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionInList:
		code = fmt.Sprintf(`%s.inList(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionNotInList:
		code = fmt.Sprintf(`!%s.inList(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionRegex:
		code = fmt.Sprintf(`%s.regexMatch(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	case models.ConditionNotRegex:
		code = fmt.Sprintf(`!%s.regexMatch(%s, "%s", %t, %v)`, tengoModuleName, objectVariable, condition.Base, isReference, parameterValue)
	}
	return
}
