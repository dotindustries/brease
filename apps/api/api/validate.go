package api

import (
	"fmt"

	"github.com/goccy/go-json"
	"go.dot.industries/brease/pb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

func (b *BreaseHandler) validateExpression(expression map[string]interface{}) error {
	exprBytes, err := json.Marshal(expression)
	if err != nil {
		b.logger.Error("expression is not base64 encoded", zap.Error(err), zap.Any("expression", expression))
		return fmt.Errorf("expression is not base64 encoded: %v", err)
	}
	expr := &pb.Expression{}
	if unmarshalErr := protojson.Unmarshal(exprBytes, expr); err != nil {
		b.logger.Error("expression cannot be read", zap.Error(unmarshalErr), zap.Any("expression", expression))
		return fmt.Errorf("expression cannot be read: %v", unmarshalErr)
	}
	b.logger.Debug("Valid expression", zap.Any("expression", expr))
	return nil
}
