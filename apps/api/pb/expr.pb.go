// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.3
// source: expr.proto

package pb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A reference to point to third party data to be used as the value for the condition.
//
// Data retrieval requires dalí integration.
// Uses a dref format based on the dalí project.
// Stringified format: `$/map[0]/prop=>/~/datasource/table/searchExpr/selectExpr`
// Object format: { src: `$/map[0]/prop`, dst: `/~/datasource/table/searchExpr/selectExpr` }
type ConditionBaseRef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Src   string `protobuf:"bytes,1,opt,name=src,proto3" json:"src,omitempty"`
	Dst   string `protobuf:"bytes,2,opt,name=dst,proto3" json:"dst,omitempty"`
	Value []byte `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *ConditionBaseRef) Reset() {
	*x = ConditionBaseRef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_expr_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ConditionBaseRef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConditionBaseRef) ProtoMessage() {}

func (x *ConditionBaseRef) ProtoReflect() protoreflect.Message {
	mi := &file_expr_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ConditionBaseRef.ProtoReflect.Descriptor instead.
func (*ConditionBaseRef) Descriptor() ([]byte, []int) {
	return file_expr_proto_rawDescGZIP(), []int{0}
}

func (x *ConditionBaseRef) GetSrc() string {
	if x != nil {
		return x.Src
	}
	return ""
}

func (x *ConditionBaseRef) GetDst() string {
	if x != nil {
		return x.Dst
	}
	return ""
}

func (x *ConditionBaseRef) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

// A condition expression
type Condition struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Base:
	//
	//	*Condition_Key
	//	*Condition_Ref
	Base isCondition_Base `protobuf_oneof:"base"`
	// The type of this condition, based on the internally defined ConditionType
	Kind string `protobuf:"bytes,3,opt,name=kind,proto3" json:"kind,omitempty"`
	// A condition parameter holds a target value (right hand side) for a comparison
	Value []byte `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Condition) Reset() {
	*x = Condition{}
	if protoimpl.UnsafeEnabled {
		mi := &file_expr_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Condition) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Condition) ProtoMessage() {}

func (x *Condition) ProtoReflect() protoreflect.Message {
	mi := &file_expr_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Condition.ProtoReflect.Descriptor instead.
func (*Condition) Descriptor() ([]byte, []int) {
	return file_expr_proto_rawDescGZIP(), []int{1}
}

func (m *Condition) GetBase() isCondition_Base {
	if m != nil {
		return m.Base
	}
	return nil
}

func (x *Condition) GetKey() string {
	if x, ok := x.GetBase().(*Condition_Key); ok {
		return x.Key
	}
	return ""
}

func (x *Condition) GetRef() *ConditionBaseRef {
	if x, ok := x.GetBase().(*Condition_Ref); ok {
		return x.Ref
	}
	return nil
}

func (x *Condition) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Condition) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

type isCondition_Base interface {
	isCondition_Base()
}

type Condition_Key struct {
	// Holds a direct object key to be used the value extraction
	Key string `protobuf:"bytes,1,opt,name=key,proto3,oneof"`
}

type Condition_Ref struct {
	// Holds a dref reference (see dalí project for details)
	Ref *ConditionBaseRef `protobuf:"bytes,2,opt,name=ref,proto3,oneof"`
}

func (*Condition_Key) isCondition_Base() {}

func (*Condition_Ref) isCondition_Base() {}

type Expression struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Expr:
	//
	//	*Expression_And
	//	*Expression_Or
	//	*Expression_Condition
	Expr isExpression_Expr `protobuf_oneof:"expr"`
}

func (x *Expression) Reset() {
	*x = Expression{}
	if protoimpl.UnsafeEnabled {
		mi := &file_expr_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Expression) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Expression) ProtoMessage() {}

func (x *Expression) ProtoReflect() protoreflect.Message {
	mi := &file_expr_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Expression.ProtoReflect.Descriptor instead.
func (*Expression) Descriptor() ([]byte, []int) {
	return file_expr_proto_rawDescGZIP(), []int{2}
}

func (m *Expression) GetExpr() isExpression_Expr {
	if m != nil {
		return m.Expr
	}
	return nil
}

func (x *Expression) GetAnd() *And {
	if x, ok := x.GetExpr().(*Expression_And); ok {
		return x.And
	}
	return nil
}

func (x *Expression) GetOr() *Or {
	if x, ok := x.GetExpr().(*Expression_Or); ok {
		return x.Or
	}
	return nil
}

func (x *Expression) GetCondition() *Condition {
	if x, ok := x.GetExpr().(*Expression_Condition); ok {
		return x.Condition
	}
	return nil
}

type isExpression_Expr interface {
	isExpression_Expr()
}

type Expression_And struct {
	And *And `protobuf:"bytes,1,opt,name=and,proto3,oneof"`
}

type Expression_Or struct {
	Or *Or `protobuf:"bytes,2,opt,name=or,proto3,oneof"`
}

type Expression_Condition struct {
	Condition *Condition `protobuf:"bytes,3,opt,name=condition,proto3,oneof"`
}

func (*Expression_And) isExpression_Expr() {}

func (*Expression_Or) isExpression_Expr() {}

func (*Expression_Condition) isExpression_Expr() {}

type And struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Expression []*Expression `protobuf:"bytes,1,rep,name=expression,proto3" json:"expression,omitempty"`
}

func (x *And) Reset() {
	*x = And{}
	if protoimpl.UnsafeEnabled {
		mi := &file_expr_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *And) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*And) ProtoMessage() {}

func (x *And) ProtoReflect() protoreflect.Message {
	mi := &file_expr_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use And.ProtoReflect.Descriptor instead.
func (*And) Descriptor() ([]byte, []int) {
	return file_expr_proto_rawDescGZIP(), []int{3}
}

func (x *And) GetExpression() []*Expression {
	if x != nil {
		return x.Expression
	}
	return nil
}

type Or struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Expression []*Expression `protobuf:"bytes,1,rep,name=expression,proto3" json:"expression,omitempty"`
}

func (x *Or) Reset() {
	*x = Or{}
	if protoimpl.UnsafeEnabled {
		mi := &file_expr_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Or) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Or) ProtoMessage() {}

func (x *Or) ProtoReflect() protoreflect.Message {
	mi := &file_expr_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Or.ProtoReflect.Descriptor instead.
func (*Or) Descriptor() ([]byte, []int) {
	return file_expr_proto_rawDescGZIP(), []int{4}
}

func (x *Or) GetExpression() []*Expression {
	if x != nil {
		return x.Expression
	}
	return nil
}

var File_expr_proto protoreflect.FileDescriptor

var file_expr_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x62, 0x72,
	0x65, 0x61, 0x73, 0x65, 0x22, 0x4c, 0x0a, 0x10, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f,
	0x6e, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x66, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x72, 0x63, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x72, 0x63, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x73,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x64, 0x73, 0x74, 0x12, 0x14, 0x0a, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c,
	0x75, 0x65, 0x22, 0x7f, 0x0a, 0x09, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x12,
	0x12, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x48, 0x00, 0x52, 0x03,
	0x6b, 0x65, 0x79, 0x12, 0x2c, 0x0a, 0x03, 0x72, 0x65, 0x66, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x18, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x2e, 0x43, 0x6f, 0x6e, 0x64, 0x69, 0x74,
	0x69, 0x6f, 0x6e, 0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x66, 0x48, 0x00, 0x52, 0x03, 0x72, 0x65,
	0x66, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x69, 0x6e, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6b, 0x69, 0x6e, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x42, 0x06, 0x0a, 0x04, 0x62,
	0x61, 0x73, 0x65, 0x22, 0x86, 0x01, 0x0a, 0x0a, 0x45, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x12, 0x1f, 0x0a, 0x03, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0b, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x2e, 0x41, 0x6e, 0x64, 0x48, 0x00, 0x52, 0x03,
	0x61, 0x6e, 0x64, 0x12, 0x1c, 0x0a, 0x02, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x0a, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x2e, 0x4f, 0x72, 0x48, 0x00, 0x52, 0x02, 0x6f,
	0x72, 0x12, 0x31, 0x0a, 0x09, 0x63, 0x6f, 0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x2e, 0x43, 0x6f,
	0x6e, 0x64, 0x69, 0x74, 0x69, 0x6f, 0x6e, 0x48, 0x00, 0x52, 0x09, 0x63, 0x6f, 0x6e, 0x64, 0x69,
	0x74, 0x69, 0x6f, 0x6e, 0x42, 0x06, 0x0a, 0x04, 0x65, 0x78, 0x70, 0x72, 0x22, 0x39, 0x0a, 0x03,
	0x41, 0x6e, 0x64, 0x12, 0x32, 0x0a, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x12, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65,
	0x2e, 0x45, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x65, 0x78, 0x70,
	0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x38, 0x0a, 0x02, 0x4f, 0x72, 0x12, 0x32, 0x0a,
	0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x12, 0x2e, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x2e, 0x45, 0x78, 0x70, 0x72, 0x65,
	0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0a, 0x65, 0x78, 0x70, 0x72, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x42, 0x1d, 0x5a, 0x1b, 0x67, 0x6f, 0x2e, 0x64, 0x6f, 0x74, 0x2e, 0x69, 0x6e, 0x64, 0x75,
	0x73, 0x74, 0x72, 0x69, 0x65, 0x73, 0x2f, 0x62, 0x72, 0x65, 0x61, 0x73, 0x65, 0x3b, 0x70, 0x62,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_expr_proto_rawDescOnce sync.Once
	file_expr_proto_rawDescData = file_expr_proto_rawDesc
)

func file_expr_proto_rawDescGZIP() []byte {
	file_expr_proto_rawDescOnce.Do(func() {
		file_expr_proto_rawDescData = protoimpl.X.CompressGZIP(file_expr_proto_rawDescData)
	})
	return file_expr_proto_rawDescData
}

var file_expr_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_expr_proto_goTypes = []interface{}{
	(*ConditionBaseRef)(nil), // 0: brease.ConditionBaseRef
	(*Condition)(nil),        // 1: brease.Condition
	(*Expression)(nil),       // 2: brease.Expression
	(*And)(nil),              // 3: brease.And
	(*Or)(nil),               // 4: brease.Or
}
var file_expr_proto_depIdxs = []int32{
	0, // 0: brease.Condition.ref:type_name -> brease.ConditionBaseRef
	3, // 1: brease.Expression.and:type_name -> brease.And
	4, // 2: brease.Expression.or:type_name -> brease.Or
	1, // 3: brease.Expression.condition:type_name -> brease.Condition
	2, // 4: brease.And.expression:type_name -> brease.Expression
	2, // 5: brease.Or.expression:type_name -> brease.Expression
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_expr_proto_init() }
func file_expr_proto_init() {
	if File_expr_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_expr_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ConditionBaseRef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_expr_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Condition); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_expr_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Expression); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_expr_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*And); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_expr_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Or); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_expr_proto_msgTypes[1].OneofWrappers = []interface{}{
		(*Condition_Key)(nil),
		(*Condition_Ref)(nil),
	}
	file_expr_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Expression_And)(nil),
		(*Expression_Or)(nil),
		(*Expression_Condition)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_expr_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_expr_proto_goTypes,
		DependencyIndexes: file_expr_proto_depIdxs,
		MessageInfos:      file_expr_proto_msgTypes,
	}.Build()
	File_expr_proto = out.File
	file_expr_proto_rawDesc = nil
	file_expr_proto_goTypes = nil
	file_expr_proto_depIdxs = nil
}
