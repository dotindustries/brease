package code

import (
	"fmt"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/interpreter"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
)

func compileCEL(expression string, expectedOutputType *types.Type, obj *structpb.Struct) (cel.Program, error) {
	env, err := cel.NewEnv(
		declareContextFromStruct(obj),
		// FIXME: this is not working for dynamic struct
		// cel.DeclareContextProto(obj.ProtoReflect().Descriptor()),
	)
	if err != nil {
		return nil, err
	}
	ast, iss := env.Parse(expression)
	// Report syntactic errors, if present.
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	checked, iss := env.Check(ast)
	// Report semantic errors, if present.
	if iss.Err() != nil {
		return nil, iss.Err()
	}
	// Check the output type is a boolean.
	if !reflect.DeepEqual(checked.OutputType(), expectedOutputType) {
		return nil, fmt.Errorf(
			"got %v, wanted %v output type",
			checked.OutputType(), expectedOutputType,
		)
	}
	program, err := env.Program(checked)
	if err != nil {
		return nil, err
	}
	return program, nil
}

func evalCEL(program cel.Program, obj *structpb.Struct) (any, error) {
	vars, err := structToCELVariables(obj)
	if err != nil {
		return nil, err
	}
	// Evaluate the program without any additional arguments.
	evalRef, _, err := program.Eval(
		vars,
	)
	if err != nil {
		return nil, err
	}
	return evalRef.Value(), nil
}

func fieldToCELType(value *structpb.Value) (*types.Type, error) {
	switch value.Kind.(type) {
	case *structpb.Value_BoolValue:
		return types.BoolType, nil
	case *structpb.Value_NumberValue:
		return types.DoubleType, nil
	case *structpb.Value_StringValue:
		return types.StringType, nil
	case *structpb.Value_StructValue:
		return types.DynType, nil
	case *structpb.Value_ListValue:
		return types.DynType, nil
	default:
		return types.DynType, nil
	}
}

func fieldToVariable(key string, field *structpb.Value) (cel.EnvOption, error) {
	switch field.Kind.(type) {
	case *structpb.Value_StructValue:
		// TODO: this is for sure not good.
		//   how to represent an arbitrary json object in CEL?
		return cel.Variable(key, cel.DynType), nil
	case *structpb.Value_ListValue:
		elemType, err := fieldToCELType(field.GetListValue().Values[0])
		if err != nil {
			return nil, err
		}
		return cel.Variable(key, cel.ListType(elemType)), nil
	default:
		celType, err := fieldToCELType(field)
		if err != nil {
			return nil, err
		}
		return cel.Variable(key, celType), nil
	}
}

func declareContextFromStruct(s *structpb.Struct) cel.EnvOption {
	return func(env *cel.Env) (*cel.Env, error) {
		for key, value := range s.Fields {
			variable, err := fieldToVariable(key, value)
			if err != nil {
				return nil, err
			}
			env, err = variable(env)
			if err != nil {
				return nil, err
			}
		}
		// TODO: add type definitions
		return cel.Types()(env)
	}
}

func getCELValue(field *structpb.Value) (any, error) {
	switch field.Kind.(type) {
	case *structpb.Value_StructValue:
		return field.GetStructValue().AsMap(), nil
	case *structpb.Value_ListValue:
		return field.GetListValue().AsSlice(), nil
	default:
		return field.AsInterface(), nil
	}
}

// structToCELVariables create CEL variables from structpb.Struct
func structToCELVariables(s *structpb.Struct) (interpreter.Activation, error) {
	if s == nil {
		return interpreter.EmptyActivation(), nil
	}

	vars := make(map[string]any, len(s.Fields))
	for k, v := range s.Fields {
		vars[k], _ = getCELValue(v)
	}

	return interpreter.NewActivation(vars)
}
