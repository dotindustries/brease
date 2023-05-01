package models

import (
	"fmt"

	"github.com/goccy/go-json"
	"go.dot.industries/brease/pb"
	"google.golang.org/protobuf/encoding/protojson"
)

func ValidateExpression(expression map[string]interface{}) (*pb.Expression, error) {
	exprBytes, err := json.Marshal(expression)
	if err != nil {
		return nil, fmt.Errorf("expression is not base64 encoded: %v", err)
	}
	expr := &pb.Expression{}
	if unmarshalErr := protojson.Unmarshal(exprBytes, expr); err != nil {
		return nil, fmt.Errorf("expression cannot be read: %v", unmarshalErr)
	}
	return expr, nil
}
