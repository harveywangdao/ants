// Code generated by protoc-gen-go. DO NOT EDIT.
// source: strategymanager.proto

package strategymanager

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type StartTaskRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Exchange             string   `protobuf:"bytes,2,opt,name=exchange,proto3" json:"exchange,omitempty"`
	ApiKey               string   `protobuf:"bytes,3,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	SecretKey            string   `protobuf:"bytes,4,opt,name=secret_key,json=secretKey,proto3" json:"secret_key,omitempty"`
	Passphrase           string   `protobuf:"bytes,5,opt,name=passphrase,proto3" json:"passphrase,omitempty"`
	StrategyName         string   `protobuf:"bytes,6,opt,name=strategy_name,json=strategyName,proto3" json:"strategy_name,omitempty"`
	InstrumentId         string   `protobuf:"bytes,7,opt,name=instrument_id,json=instrumentId,proto3" json:"instrument_id,omitempty"`
	Endpoint             string   `protobuf:"bytes,8,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	WsEndpoint           string   `protobuf:"bytes,9,opt,name=ws_endpoint,json=wsEndpoint,proto3" json:"ws_endpoint,omitempty"`
	Params               string   `protobuf:"bytes,10,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StartTaskRequest) Reset()         { *m = StartTaskRequest{} }
func (m *StartTaskRequest) String() string { return proto.CompactTextString(m) }
func (*StartTaskRequest) ProtoMessage()    {}
func (*StartTaskRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{0}
}

func (m *StartTaskRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StartTaskRequest.Unmarshal(m, b)
}
func (m *StartTaskRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StartTaskRequest.Marshal(b, m, deterministic)
}
func (m *StartTaskRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StartTaskRequest.Merge(m, src)
}
func (m *StartTaskRequest) XXX_Size() int {
	return xxx_messageInfo_StartTaskRequest.Size(m)
}
func (m *StartTaskRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StartTaskRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StartTaskRequest proto.InternalMessageInfo

func (m *StartTaskRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *StartTaskRequest) GetExchange() string {
	if m != nil {
		return m.Exchange
	}
	return ""
}

func (m *StartTaskRequest) GetApiKey() string {
	if m != nil {
		return m.ApiKey
	}
	return ""
}

func (m *StartTaskRequest) GetSecretKey() string {
	if m != nil {
		return m.SecretKey
	}
	return ""
}

func (m *StartTaskRequest) GetPassphrase() string {
	if m != nil {
		return m.Passphrase
	}
	return ""
}

func (m *StartTaskRequest) GetStrategyName() string {
	if m != nil {
		return m.StrategyName
	}
	return ""
}

func (m *StartTaskRequest) GetInstrumentId() string {
	if m != nil {
		return m.InstrumentId
	}
	return ""
}

func (m *StartTaskRequest) GetEndpoint() string {
	if m != nil {
		return m.Endpoint
	}
	return ""
}

func (m *StartTaskRequest) GetWsEndpoint() string {
	if m != nil {
		return m.WsEndpoint
	}
	return ""
}

func (m *StartTaskRequest) GetParams() string {
	if m != nil {
		return m.Params
	}
	return ""
}

type StartTaskResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StartTaskResponse) Reset()         { *m = StartTaskResponse{} }
func (m *StartTaskResponse) String() string { return proto.CompactTextString(m) }
func (*StartTaskResponse) ProtoMessage()    {}
func (*StartTaskResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{1}
}

func (m *StartTaskResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StartTaskResponse.Unmarshal(m, b)
}
func (m *StartTaskResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StartTaskResponse.Marshal(b, m, deterministic)
}
func (m *StartTaskResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StartTaskResponse.Merge(m, src)
}
func (m *StartTaskResponse) XXX_Size() int {
	return xxx_messageInfo_StartTaskResponse.Size(m)
}
func (m *StartTaskResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StartTaskResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StartTaskResponse proto.InternalMessageInfo

type StopTaskRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Exchange             string   `protobuf:"bytes,2,opt,name=exchange,proto3" json:"exchange,omitempty"`
	ApiKey               string   `protobuf:"bytes,3,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	StrategyName         string   `protobuf:"bytes,4,opt,name=strategy_name,json=strategyName,proto3" json:"strategy_name,omitempty"`
	InstrumentId         string   `protobuf:"bytes,5,opt,name=instrument_id,json=instrumentId,proto3" json:"instrument_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopTaskRequest) Reset()         { *m = StopTaskRequest{} }
func (m *StopTaskRequest) String() string { return proto.CompactTextString(m) }
func (*StopTaskRequest) ProtoMessage()    {}
func (*StopTaskRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{2}
}

func (m *StopTaskRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopTaskRequest.Unmarshal(m, b)
}
func (m *StopTaskRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopTaskRequest.Marshal(b, m, deterministic)
}
func (m *StopTaskRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopTaskRequest.Merge(m, src)
}
func (m *StopTaskRequest) XXX_Size() int {
	return xxx_messageInfo_StopTaskRequest.Size(m)
}
func (m *StopTaskRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_StopTaskRequest.DiscardUnknown(m)
}

var xxx_messageInfo_StopTaskRequest proto.InternalMessageInfo

func (m *StopTaskRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *StopTaskRequest) GetExchange() string {
	if m != nil {
		return m.Exchange
	}
	return ""
}

func (m *StopTaskRequest) GetApiKey() string {
	if m != nil {
		return m.ApiKey
	}
	return ""
}

func (m *StopTaskRequest) GetStrategyName() string {
	if m != nil {
		return m.StrategyName
	}
	return ""
}

func (m *StopTaskRequest) GetInstrumentId() string {
	if m != nil {
		return m.InstrumentId
	}
	return ""
}

type StopTaskResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *StopTaskResponse) Reset()         { *m = StopTaskResponse{} }
func (m *StopTaskResponse) String() string { return proto.CompactTextString(m) }
func (*StopTaskResponse) ProtoMessage()    {}
func (*StopTaskResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{3}
}

func (m *StopTaskResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_StopTaskResponse.Unmarshal(m, b)
}
func (m *StopTaskResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_StopTaskResponse.Marshal(b, m, deterministic)
}
func (m *StopTaskResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StopTaskResponse.Merge(m, src)
}
func (m *StopTaskResponse) XXX_Size() int {
	return xxx_messageInfo_StopTaskResponse.Size(m)
}
func (m *StopTaskResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_StopTaskResponse.DiscardUnknown(m)
}

var xxx_messageInfo_StopTaskResponse proto.InternalMessageInfo

type TaskCommandExecRequest struct {
	UserId               string   `protobuf:"bytes,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	Exchange             string   `protobuf:"bytes,2,opt,name=exchange,proto3" json:"exchange,omitempty"`
	ApiKey               string   `protobuf:"bytes,3,opt,name=api_key,json=apiKey,proto3" json:"api_key,omitempty"`
	StrategyName         string   `protobuf:"bytes,4,opt,name=strategy_name,json=strategyName,proto3" json:"strategy_name,omitempty"`
	InstrumentId         string   `protobuf:"bytes,5,opt,name=instrument_id,json=instrumentId,proto3" json:"instrument_id,omitempty"`
	Params               string   `protobuf:"bytes,6,opt,name=params,proto3" json:"params,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TaskCommandExecRequest) Reset()         { *m = TaskCommandExecRequest{} }
func (m *TaskCommandExecRequest) String() string { return proto.CompactTextString(m) }
func (*TaskCommandExecRequest) ProtoMessage()    {}
func (*TaskCommandExecRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{4}
}

func (m *TaskCommandExecRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TaskCommandExecRequest.Unmarshal(m, b)
}
func (m *TaskCommandExecRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TaskCommandExecRequest.Marshal(b, m, deterministic)
}
func (m *TaskCommandExecRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskCommandExecRequest.Merge(m, src)
}
func (m *TaskCommandExecRequest) XXX_Size() int {
	return xxx_messageInfo_TaskCommandExecRequest.Size(m)
}
func (m *TaskCommandExecRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskCommandExecRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TaskCommandExecRequest proto.InternalMessageInfo

func (m *TaskCommandExecRequest) GetUserId() string {
	if m != nil {
		return m.UserId
	}
	return ""
}

func (m *TaskCommandExecRequest) GetExchange() string {
	if m != nil {
		return m.Exchange
	}
	return ""
}

func (m *TaskCommandExecRequest) GetApiKey() string {
	if m != nil {
		return m.ApiKey
	}
	return ""
}

func (m *TaskCommandExecRequest) GetStrategyName() string {
	if m != nil {
		return m.StrategyName
	}
	return ""
}

func (m *TaskCommandExecRequest) GetInstrumentId() string {
	if m != nil {
		return m.InstrumentId
	}
	return ""
}

func (m *TaskCommandExecRequest) GetParams() string {
	if m != nil {
		return m.Params
	}
	return ""
}

type TaskCommandExecResponse struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TaskCommandExecResponse) Reset()         { *m = TaskCommandExecResponse{} }
func (m *TaskCommandExecResponse) String() string { return proto.CompactTextString(m) }
func (*TaskCommandExecResponse) ProtoMessage()    {}
func (*TaskCommandExecResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_4e316c09d5f1b9b5, []int{5}
}

func (m *TaskCommandExecResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TaskCommandExecResponse.Unmarshal(m, b)
}
func (m *TaskCommandExecResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TaskCommandExecResponse.Marshal(b, m, deterministic)
}
func (m *TaskCommandExecResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TaskCommandExecResponse.Merge(m, src)
}
func (m *TaskCommandExecResponse) XXX_Size() int {
	return xxx_messageInfo_TaskCommandExecResponse.Size(m)
}
func (m *TaskCommandExecResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TaskCommandExecResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TaskCommandExecResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*StartTaskRequest)(nil), "strategymanager.StartTaskRequest")
	proto.RegisterType((*StartTaskResponse)(nil), "strategymanager.StartTaskResponse")
	proto.RegisterType((*StopTaskRequest)(nil), "strategymanager.StopTaskRequest")
	proto.RegisterType((*StopTaskResponse)(nil), "strategymanager.StopTaskResponse")
	proto.RegisterType((*TaskCommandExecRequest)(nil), "strategymanager.TaskCommandExecRequest")
	proto.RegisterType((*TaskCommandExecResponse)(nil), "strategymanager.TaskCommandExecResponse")
}

func init() { proto.RegisterFile("strategymanager.proto", fileDescriptor_4e316c09d5f1b9b5) }

var fileDescriptor_4e316c09d5f1b9b5 = []byte{
	// 394 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x54, 0xc1, 0x4e, 0xea, 0x40,
	0x14, 0x7d, 0xe5, 0x41, 0xa1, 0xf7, 0xbd, 0x17, 0xde, 0x9b, 0x17, 0xa1, 0x36, 0x51, 0xa1, 0x2e,
	0x64, 0xc5, 0x42, 0x3f, 0xc1, 0xb0, 0x30, 0x46, 0x13, 0x81, 0x7d, 0x33, 0xd2, 0x2b, 0x34, 0xa4,
	0xd3, 0x71, 0x66, 0x08, 0xf0, 0x3b, 0xee, 0xfc, 0x09, 0xd7, 0x7e, 0x96, 0x69, 0xa7, 0xad, 0xd0,
	0x12, 0x74, 0x63, 0xe2, 0xf2, 0x9e, 0x73, 0x6e, 0xe7, 0xdc, 0x73, 0x67, 0x0a, 0x07, 0x52, 0x09,
	0xaa, 0x70, 0xba, 0x0e, 0x29, 0xa3, 0x53, 0x14, 0x7d, 0x2e, 0x22, 0x15, 0x91, 0x66, 0x01, 0x76,
	0x5f, 0x2a, 0xf0, 0x77, 0xa4, 0xa8, 0x50, 0x63, 0x2a, 0xe7, 0x43, 0x7c, 0x5c, 0xa0, 0x54, 0xa4,
	0x0d, 0xf5, 0x85, 0x44, 0xe1, 0x05, 0xbe, 0x6d, 0x74, 0x8c, 0x9e, 0x35, 0x34, 0xe3, 0xf2, 0xca,
	0x27, 0x0e, 0x34, 0x70, 0x35, 0x99, 0x51, 0x36, 0x45, 0xbb, 0x92, 0x30, 0x79, 0x1d, 0x37, 0x51,
	0x1e, 0x78, 0x73, 0x5c, 0xdb, 0x3f, 0x75, 0x13, 0xe5, 0xc1, 0x35, 0xae, 0xc9, 0x11, 0x80, 0xc4,
	0x89, 0x40, 0x95, 0x70, 0xd5, 0x84, 0xb3, 0x34, 0x12, 0xd3, 0xc7, 0x00, 0x9c, 0x4a, 0xc9, 0x67,
	0x82, 0x4a, 0xb4, 0x6b, 0x09, 0xbd, 0x81, 0x90, 0x53, 0xf8, 0x93, 0x99, 0xf6, 0x18, 0x0d, 0xd1,
	0x36, 0x13, 0xc9, 0xef, 0x0c, 0xbc, 0xa5, 0x61, 0x22, 0x0a, 0x98, 0x54, 0x62, 0x11, 0x22, 0x53,
	0xb1, 0xef, 0xba, 0x16, 0xbd, 0x83, 0xa9, 0x7b, 0xe6, 0xf3, 0x28, 0x60, 0xca, 0x6e, 0xa4, 0xee,
	0xd3, 0x9a, 0x9c, 0xc0, 0xaf, 0xa5, 0xf4, 0x72, 0xda, 0xd2, 0x36, 0x96, 0x72, 0x90, 0x09, 0x5a,
	0x60, 0x72, 0x2a, 0x68, 0x28, 0x6d, 0xd0, 0xd3, 0xe9, 0xca, 0xfd, 0x0f, 0xff, 0x36, 0xf2, 0x93,
	0x3c, 0x62, 0x12, 0xdd, 0x67, 0x03, 0x9a, 0x23, 0x15, 0xf1, 0xaf, 0x0b, 0xb5, 0x94, 0x4a, 0xf5,
	0x33, 0xa9, 0xd4, 0xca, 0xa9, 0xb8, 0x24, 0xbe, 0x00, 0x99, 0xd5, 0xd4, 0xff, 0xab, 0x01, 0xad,
	0x18, 0xb8, 0x8c, 0xc2, 0x90, 0x32, 0x7f, 0xb0, 0xc2, 0xc9, 0x37, 0x1f, 0x63, 0x63, 0x3f, 0xe6,
	0xd6, 0x7e, 0x0e, 0xa1, 0x5d, 0x9a, 0x44, 0x4f, 0x79, 0xfe, 0x54, 0x89, 0xb7, 0xa4, 0x0f, 0xba,
	0xd1, 0xef, 0x81, 0x8c, 0xc1, 0xca, 0xd7, 0x49, 0xba, 0xfd, 0xe2, 0x2b, 0x2a, 0x3e, 0x15, 0xc7,
	0xdd, 0x27, 0x49, 0xd3, 0xfc, 0x41, 0xee, 0xa0, 0x91, 0x65, 0x4c, 0x3a, 0x3b, 0x3a, 0xb6, 0x6e,
	0x8a, 0xd3, 0xdd, 0xa3, 0xc8, 0x3f, 0xf9, 0x00, 0xcd, 0xc2, 0x5c, 0xe4, 0xac, 0xd4, 0xb7, 0x7b,
	0x87, 0x4e, 0xef, 0x63, 0x61, 0x76, 0xce, 0xbd, 0x99, 0xfc, 0x38, 0x2e, 0xde, 0x02, 0x00, 0x00,
	0xff, 0xff, 0xcd, 0x90, 0x75, 0x6c, 0x51, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// StrategyManagerClient is the client API for StrategyManager service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type StrategyManagerClient interface {
	StartTask(ctx context.Context, in *StartTaskRequest, opts ...grpc.CallOption) (*StartTaskResponse, error)
	StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskResponse, error)
	TaskCommandExec(ctx context.Context, in *TaskCommandExecRequest, opts ...grpc.CallOption) (*TaskCommandExecResponse, error)
}

type strategyManagerClient struct {
	cc *grpc.ClientConn
}

func NewStrategyManagerClient(cc *grpc.ClientConn) StrategyManagerClient {
	return &strategyManagerClient{cc}
}

func (c *strategyManagerClient) StartTask(ctx context.Context, in *StartTaskRequest, opts ...grpc.CallOption) (*StartTaskResponse, error) {
	out := new(StartTaskResponse)
	err := c.cc.Invoke(ctx, "/strategymanager.StrategyManager/StartTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyManagerClient) StopTask(ctx context.Context, in *StopTaskRequest, opts ...grpc.CallOption) (*StopTaskResponse, error) {
	out := new(StopTaskResponse)
	err := c.cc.Invoke(ctx, "/strategymanager.StrategyManager/StopTask", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *strategyManagerClient) TaskCommandExec(ctx context.Context, in *TaskCommandExecRequest, opts ...grpc.CallOption) (*TaskCommandExecResponse, error) {
	out := new(TaskCommandExecResponse)
	err := c.cc.Invoke(ctx, "/strategymanager.StrategyManager/TaskCommandExec", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// StrategyManagerServer is the server API for StrategyManager service.
type StrategyManagerServer interface {
	StartTask(context.Context, *StartTaskRequest) (*StartTaskResponse, error)
	StopTask(context.Context, *StopTaskRequest) (*StopTaskResponse, error)
	TaskCommandExec(context.Context, *TaskCommandExecRequest) (*TaskCommandExecResponse, error)
}

func RegisterStrategyManagerServer(s *grpc.Server, srv StrategyManagerServer) {
	s.RegisterService(&_StrategyManager_serviceDesc, srv)
}

func _StrategyManager_StartTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StartTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyManagerServer).StartTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/strategymanager.StrategyManager/StartTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyManagerServer).StartTask(ctx, req.(*StartTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StrategyManager_StopTask_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StopTaskRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyManagerServer).StopTask(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/strategymanager.StrategyManager/StopTask",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyManagerServer).StopTask(ctx, req.(*StopTaskRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _StrategyManager_TaskCommandExec_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskCommandExecRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(StrategyManagerServer).TaskCommandExec(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/strategymanager.StrategyManager/TaskCommandExec",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(StrategyManagerServer).TaskCommandExec(ctx, req.(*TaskCommandExecRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _StrategyManager_serviceDesc = grpc.ServiceDesc{
	ServiceName: "strategymanager.StrategyManager",
	HandlerType: (*StrategyManagerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "StartTask",
			Handler:    _StrategyManager_StartTask_Handler,
		},
		{
			MethodName: "StopTask",
			Handler:    _StrategyManager_StopTask_Handler,
		},
		{
			MethodName: "TaskCommandExec",
			Handler:    _StrategyManager_TaskCommandExec_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "strategymanager.proto",
}
