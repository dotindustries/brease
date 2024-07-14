package code

import (
	"github.com/google/cel-go/common/types"
	"google.golang.org/protobuf/types/known/structpb"
	"reflect"
	"testing"
)

func Test_evalCEL(t *testing.T) {
	type args struct {
		expression         string
		expectedOutputType *types.Type
		val                *structpb.Struct
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "bool test",
			args: args{
				expression:         "method == 'GET'",
				expectedOutputType: types.BoolType,
				val: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"method": {Kind: &structpb.Value_StringValue{StringValue: "GET"}},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prg, err := compileCEL(tt.args.expression, tt.args.expectedOutputType, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("compileCEL() error = %v, wantErr %v", err, tt.wantErr)
			}
			got, err := evalCEL(prg, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("evalCEL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("evalCEL() got = %v, want %v", got, tt.want)
			}
		})
	}
}
