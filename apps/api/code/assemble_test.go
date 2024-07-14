package code

import (
	rulev1 "buf.build/gen/go/dot/brease/protocolbuffers/go/brease/rule/v1"
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.dot.industries/brease/cache"
	"go.uber.org/zap"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"gotest.tools/v3/assert"
	"testing"
)

func TestCode_Evaluate(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jsonStr := `{"pre_01httnykd6fjvt518c3yxvx3r8":{"prse_01hv6qqj1ve7zvpvq03ak1b3w8":"the hobbit goes a long way"}}`
	obj := &structpb.Struct{}
	err := protojson.Unmarshal([]byte(jsonStr), obj)
	assert.NilError(t, err)

	type fields struct {
		logger *zap.Logger
		cache  cache.Cache
	}
	type args struct {
		ctx    context.Context
		object *structpb.Struct
		rules  []*rulev1.VersionedRule
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*rulev1.EvaluationResult
		wantErr bool
	}{
		{
			name: "simple test",
			fields: fields{
				logger: logger,
			},
			args: args{
				ctx:    context.TODO(),
				object: obj,
				rules: []*rulev1.VersionedRule{
					{
						Id:          "rule_01h89qfdhbejtb3jwqq1gazbm5",
						Version:     0,
						Description: "rule_01h89qfdhbejtb3jwqq1gazbm5",
						Actions: []*rulev1.Action{
							{
								Kind: "setValue",
								Target: &rulev1.Target{
									Kind:  "jsonpath",
									Id:    "$.method",
									Value: []byte(`ZXhhbXBsZQ==`),
								},
							},
						},
						Expression: &rulev1.Expression{
							Expr: &rulev1.Expression_Condition{
								Condition: &rulev1.Condition{
									// base is not used for CEL
									Base: &rulev1.Condition_Key{
										Key: "",
									},
									Kind:  rulev1.ConditionKind_cel,
									Value: []byte(`InByZV8wMWh0dG55a2Q2Zmp2dDUxOGMzeXh2eDNyOC5wcnNlXzAxaHY2cXFqMXZlN3p2cHZxMDNhazFiM3c4LmNvbnRhaW5zKFwiaG9iYml0XCIpIg==`),
								},
							},
						},
					},
				},
			},
			want: []*rulev1.EvaluationResult{
				{
					Action: "setValue",
					Target: &rulev1.Target{Kind: "jsonpath", Id: "$.method", Value: []byte(`ZXhhbXBsZQ==`)},
					By:     &rulev1.RuleRef{Id: "rule_01h89qfdhbejtb3jwqq1gazbm5", Description: "rule_01h89qfdhbejtb3jwqq1gazbm5"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bts, _ := proto.Marshal(tt.args.object)
			slog.Info("dumping object", "object", string(bts))
			a := &Assembler{
				logger: tt.fields.logger,
				cache:  tt.fields.cache,
			}
			c := NewCompiler(tt.fields.logger)
			code, err := a.BuildCode(tt.args.ctx, tt.args.rules)
			assert.NilError(t, err)
			compiledScript, err := c.CompileCode(tt.args.ctx, code)
			assert.NilError(t, err)
			run, err := NewRun(tt.args.ctx, tt.fields.logger, tt.args.object)
			assert.NilError(t, err)
			got, err := run.Execute(tt.args.ctx, compiledScript)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.IgnoreUnexported(rulev1.EvaluationResult{}, rulev1.Target{}, rulev1.RuleRef{})); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
