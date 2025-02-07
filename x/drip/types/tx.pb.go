// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: juno/drip/v1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// MsgDistributeTokens defines a message that registers a Distribution of tokens.
type MsgDistributeTokens struct {
	// sender_address is the bech32 address of message sender.
	SenderAddress string `protobuf:"bytes,1,opt,name=sender_address,json=senderAddress,proto3" json:"sender_address,omitempty"`
	// amount is the amount being airdropped to stakers
	Amount github_com_cosmos_cosmos_sdk_types.Coins `protobuf:"bytes,2,rep,name=amount,proto3,castrepeated=github.com/cosmos/cosmos-sdk/types.Coins" json:"amount"`
}

func (m *MsgDistributeTokens) Reset()         { *m = MsgDistributeTokens{} }
func (m *MsgDistributeTokens) String() string { return proto.CompactTextString(m) }
func (*MsgDistributeTokens) ProtoMessage()    {}
func (*MsgDistributeTokens) Descriptor() ([]byte, []int) {
	return fileDescriptor_73c0f1d75f17f4bc, []int{0}
}
func (m *MsgDistributeTokens) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgDistributeTokens) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgDistributeTokens.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgDistributeTokens) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgDistributeTokens.Merge(m, src)
}
func (m *MsgDistributeTokens) XXX_Size() int {
	return m.Size()
}
func (m *MsgDistributeTokens) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgDistributeTokens.DiscardUnknown(m)
}

var xxx_messageInfo_MsgDistributeTokens proto.InternalMessageInfo

func (m *MsgDistributeTokens) GetSenderAddress() string {
	if m != nil {
		return m.SenderAddress
	}
	return ""
}

func (m *MsgDistributeTokens) GetAmount() github_com_cosmos_cosmos_sdk_types.Coins {
	if m != nil {
		return m.Amount
	}
	return nil
}

// MsgDistributeTokensResponse defines the MsgDistributeTokens response type
type MsgDistributeTokensResponse struct {
}

func (m *MsgDistributeTokensResponse) Reset()         { *m = MsgDistributeTokensResponse{} }
func (m *MsgDistributeTokensResponse) String() string { return proto.CompactTextString(m) }
func (*MsgDistributeTokensResponse) ProtoMessage()    {}
func (*MsgDistributeTokensResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_73c0f1d75f17f4bc, []int{1}
}
func (m *MsgDistributeTokensResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgDistributeTokensResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgDistributeTokensResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgDistributeTokensResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgDistributeTokensResponse.Merge(m, src)
}
func (m *MsgDistributeTokensResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgDistributeTokensResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgDistributeTokensResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgDistributeTokensResponse proto.InternalMessageInfo

// MsgUpdateParams is the Msg/UpdateParams request type.
type MsgUpdateParams struct {
	// authority is the address that controls the module (defaults to x/gov unless overwritten).
	Authority string `protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	// params defines the x/auth parameters to update.
	//
	// NOTE: All parameters must be supplied.
	Params Params `protobuf:"bytes,2,opt,name=params,proto3" json:"params"`
}

func (m *MsgUpdateParams) Reset()         { *m = MsgUpdateParams{} }
func (m *MsgUpdateParams) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateParams) ProtoMessage()    {}
func (*MsgUpdateParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_73c0f1d75f17f4bc, []int{2}
}
func (m *MsgUpdateParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateParams.Merge(m, src)
}
func (m *MsgUpdateParams) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateParams) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateParams.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateParams proto.InternalMessageInfo

// MsgUpdateParamsResponse defines the response structure for executing a
// MsgUpdateParams message.
type MsgUpdateParamsResponse struct {
}

func (m *MsgUpdateParamsResponse) Reset()         { *m = MsgUpdateParamsResponse{} }
func (m *MsgUpdateParamsResponse) String() string { return proto.CompactTextString(m) }
func (*MsgUpdateParamsResponse) ProtoMessage()    {}
func (*MsgUpdateParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_73c0f1d75f17f4bc, []int{3}
}
func (m *MsgUpdateParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgUpdateParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgUpdateParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgUpdateParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgUpdateParamsResponse.Merge(m, src)
}
func (m *MsgUpdateParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgUpdateParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgUpdateParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgUpdateParamsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgDistributeTokens)(nil), "juno.drip.v1.MsgDistributeTokens")
	proto.RegisterType((*MsgDistributeTokensResponse)(nil), "juno.drip.v1.MsgDistributeTokensResponse")
	proto.RegisterType((*MsgUpdateParams)(nil), "juno.drip.v1.MsgUpdateParams")
	proto.RegisterType((*MsgUpdateParamsResponse)(nil), "juno.drip.v1.MsgUpdateParamsResponse")
}

func init() { proto.RegisterFile("juno/drip/v1/tx.proto", fileDescriptor_73c0f1d75f17f4bc) }

var fileDescriptor_73c0f1d75f17f4bc = []byte{
	// 575 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x53, 0x4f, 0x6b, 0x13, 0x41,
	0x1c, 0xdd, 0x4d, 0x31, 0x90, 0x69, 0xfc, 0xb7, 0x46, 0x9a, 0x6c, 0xed, 0x26, 0x2e, 0x0a, 0x31,
	0xd8, 0x1d, 0x12, 0xa1, 0x42, 0x3d, 0x88, 0x89, 0xf4, 0x16, 0x90, 0x58, 0x2f, 0x5e, 0xc2, 0x24,
	0x3b, 0x4c, 0xc7, 0x76, 0x67, 0x96, 0x9d, 0x49, 0x68, 0x6e, 0xd2, 0x93, 0x88, 0x88, 0xe0, 0xcd,
	0x53, 0x8f, 0xe2, 0x29, 0x07, 0x4f, 0x7e, 0x82, 0x1e, 0x8b, 0x5e, 0x3c, 0xa9, 0x24, 0x42, 0xbc,
	0xfa, 0x0d, 0x64, 0x77, 0x27, 0xed, 0x26, 0x0d, 0xf6, 0xb2, 0xbb, 0xf3, 0xde, 0xef, 0xdf, 0x7b,
	0xfb, 0x1b, 0x70, 0xfd, 0x45, 0x8f, 0x71, 0xe8, 0x06, 0xd4, 0x87, 0xfd, 0x2a, 0x94, 0xfb, 0x8e,
	0x1f, 0x70, 0xc9, 0x8d, 0x6c, 0x08, 0x3b, 0x21, 0xec, 0xf4, 0xab, 0xe6, 0x55, 0xe4, 0x51, 0xc6,
	0x61, 0xf4, 0x8c, 0x03, 0x4c, 0xab, 0xcb, 0x85, 0xc7, 0x05, 0xec, 0x20, 0x81, 0x61, 0xbf, 0xda,
	0xc1, 0x12, 0x55, 0x61, 0x97, 0x53, 0xa6, 0xf8, 0x15, 0xc5, 0x7b, 0x82, 0x84, 0x85, 0x3d, 0x41,
	0x14, 0x51, 0x88, 0x89, 0x76, 0x74, 0x82, 0xf1, 0x41, 0x51, 0x39, 0xc2, 0x09, 0x8f, 0xf1, 0xf0,
	0x4b, 0xa1, 0x37, 0x08, 0xe7, 0x64, 0x0f, 0x43, 0xe4, 0x53, 0x88, 0x18, 0xe3, 0x12, 0x49, 0xca,
	0xd9, 0x34, 0xc7, 0x9c, 0x99, 0x9f, 0x60, 0x86, 0x05, 0x55, 0x9c, 0xfd, 0x36, 0x05, 0xae, 0x35,
	0x05, 0x79, 0x4c, 0x85, 0x0c, 0x68, 0xa7, 0x27, 0xf1, 0x36, 0xdf, 0xc5, 0x4c, 0x18, 0x0f, 0xc1,
	0x25, 0x81, 0x99, 0x8b, 0x83, 0x36, 0x72, 0xdd, 0x00, 0x0b, 0x91, 0xd7, 0x4b, 0x7a, 0x39, 0x53,
	0xcf, 0x7f, 0xfd, 0xbc, 0x9e, 0x53, 0x13, 0x3d, 0x8a, 0x99, 0xa7, 0x32, 0xa0, 0x8c, 0xb4, 0x2e,
	0xc6, 0xf1, 0x0a, 0x34, 0x06, 0x20, 0x8d, 0x3c, 0xde, 0x63, 0x32, 0x9f, 0x2a, 0x2d, 0x95, 0x97,
	0x6b, 0x05, 0x47, 0x65, 0x85, 0x6e, 0x38, 0xca, 0x0d, 0xa7, 0xc1, 0x29, 0xab, 0x6f, 0x1d, 0xfd,
	0x28, 0x6a, 0x9f, 0x7e, 0x16, 0xcb, 0x84, 0xca, 0x9d, 0x5e, 0xc7, 0xe9, 0x72, 0x4f, 0x89, 0x56,
	0xaf, 0x75, 0xe1, 0xee, 0x42, 0x39, 0xf0, 0xb1, 0x88, 0x12, 0xc4, 0x87, 0xc9, 0xb0, 0x92, 0xdd,
	0xc3, 0x04, 0x75, 0x07, 0xed, 0xd0, 0x4f, 0xf1, 0x71, 0x32, 0xac, 0xe8, 0x2d, 0xd5, 0x70, 0xf3,
	0xc1, 0x9f, 0xc3, 0xa2, 0x76, 0x30, 0x19, 0x56, 0xe6, 0x24, 0xbc, 0x9e, 0x0c, 0x2b, 0xc5, 0xc8,
	0x8b, 0xfd, 0xd8, 0x8d, 0x05, 0xc2, 0xed, 0x35, 0xb0, 0xba, 0x00, 0x6e, 0x61, 0xe1, 0x73, 0x26,
	0xb0, 0xfd, 0x45, 0x07, 0x97, 0x9b, 0x82, 0x3c, 0xf3, 0x5d, 0x24, 0xf1, 0x13, 0x14, 0x20, 0x4f,
	0x18, 0x1b, 0x20, 0x83, 0x7a, 0x72, 0x87, 0x07, 0x54, 0x0e, 0xce, 0xb5, 0xe9, 0x34, 0xd4, 0xb8,
	0x0f, 0xd2, 0x7e, 0x54, 0x21, 0x9f, 0x2a, 0xe9, 0xe5, 0xe5, 0x5a, 0xce, 0x49, 0x6e, 0x94, 0x13,
	0x57, 0xaf, 0x67, 0x42, 0x77, 0x94, 0xc0, 0x38, 0x7c, 0x73, 0xe3, 0xd5, 0x61, 0x51, 0x9b, 0x8a,
	0x3c, 0x2d, 0x18, 0xea, 0x5b, 0x9d, 0xd3, 0x97, 0x1c, 0xd4, 0x2e, 0x80, 0x95, 0x39, 0x68, 0xaa,
	0xab, 0xf6, 0x57, 0x07, 0x4b, 0x4d, 0x41, 0x8c, 0x37, 0x3a, 0xb8, 0x72, 0x66, 0x19, 0x6e, 0xce,
	0x0e, 0xb6, 0xc0, 0x1f, 0xf3, 0xce, 0xb9, 0x21, 0x27, 0x16, 0x56, 0x0e, 0xbe, 0xfd, 0x7e, 0x9f,
	0xba, 0x65, 0xdb, 0x70, 0xee, 0x5e, 0x41, 0xf7, 0x24, 0xa5, 0x2d, 0xe3, 0xce, 0xdb, 0x20, 0x3b,
	0x63, 0xf5, 0xda, 0x99, 0x36, 0x49, 0xda, 0xbc, 0xfd, 0x5f, 0x7a, 0x3a, 0x81, 0x79, 0xe1, 0x65,
	0x68, 0x67, 0x7d, 0xeb, 0x68, 0x64, 0xe9, 0xc7, 0x23, 0x4b, 0xff, 0x35, 0xb2, 0xf4, 0x77, 0x63,
	0x4b, 0x3b, 0x1e, 0x5b, 0xda, 0xf7, 0xb1, 0xa5, 0x3d, 0xbf, 0x9b, 0xd8, 0xc4, 0x46, 0xf4, 0x17,
	0x1b, 0x9c, 0xc9, 0x00, 0x75, 0xa5, 0x80, 0x49, 0x83, 0xa3, 0x9d, 0xec, 0xa4, 0xa3, 0xab, 0x74,
	0xef, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xf2, 0xaf, 0x29, 0xd0, 0x28, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	// DistributeTokens distribute the sent tokens to all stakers in the next block
	DistributeTokens(ctx context.Context, in *MsgDistributeTokens, opts ...grpc.CallOption) (*MsgDistributeTokensResponse, error)
	// UpdateParams defines a governance operation for updating the x/clock module
	// parameters. The authority is hard-coded to the x/gov module account.
	UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) DistributeTokens(ctx context.Context, in *MsgDistributeTokens, opts ...grpc.CallOption) (*MsgDistributeTokensResponse, error) {
	out := new(MsgDistributeTokensResponse)
	err := c.cc.Invoke(ctx, "/juno.drip.v1.Msg/DistributeTokens", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *msgClient) UpdateParams(ctx context.Context, in *MsgUpdateParams, opts ...grpc.CallOption) (*MsgUpdateParamsResponse, error) {
	out := new(MsgUpdateParamsResponse)
	err := c.cc.Invoke(ctx, "/juno.drip.v1.Msg/UpdateParams", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	// DistributeTokens distribute the sent tokens to all stakers in the next block
	DistributeTokens(context.Context, *MsgDistributeTokens) (*MsgDistributeTokensResponse, error)
	// UpdateParams defines a governance operation for updating the x/clock module
	// parameters. The authority is hard-coded to the x/gov module account.
	UpdateParams(context.Context, *MsgUpdateParams) (*MsgUpdateParamsResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) DistributeTokens(ctx context.Context, req *MsgDistributeTokens) (*MsgDistributeTokensResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DistributeTokens not implemented")
}
func (*UnimplementedMsgServer) UpdateParams(ctx context.Context, req *MsgUpdateParams) (*MsgUpdateParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateParams not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_DistributeTokens_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgDistributeTokens)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).DistributeTokens(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/juno.drip.v1.Msg/DistributeTokens",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).DistributeTokens(ctx, req.(*MsgDistributeTokens))
	}
	return interceptor(ctx, in, info, handler)
}

func _Msg_UpdateParams_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgUpdateParams)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).UpdateParams(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/juno.drip.v1.Msg/UpdateParams",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).UpdateParams(ctx, req.(*MsgUpdateParams))
	}
	return interceptor(ctx, in, info, handler)
}

var Msg_serviceDesc = _Msg_serviceDesc
var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "juno.drip.v1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "DistributeTokens",
			Handler:    _Msg_DistributeTokens_Handler,
		},
		{
			MethodName: "UpdateParams",
			Handler:    _Msg_UpdateParams_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "juno/drip/v1/tx.proto",
}

func (m *MsgDistributeTokens) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDistributeTokens) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDistributeTokens) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Amount) > 0 {
		for iNdEx := len(m.Amount) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Amount[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintTx(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if len(m.SenderAddress) > 0 {
		i -= len(m.SenderAddress)
		copy(dAtA[i:], m.SenderAddress)
		i = encodeVarintTx(dAtA, i, uint64(len(m.SenderAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgDistributeTokensResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgDistributeTokensResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgDistributeTokensResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *MsgUpdateParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Params.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintTx(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgUpdateParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgUpdateParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgUpdateParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgDistributeTokens) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.SenderAddress)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	if len(m.Amount) > 0 {
		for _, e := range m.Amount {
			l = e.Size()
			n += 1 + l + sovTx(uint64(l))
		}
	}
	return n
}

func (m *MsgDistributeTokensResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *MsgUpdateParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Params.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgUpdateParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgDistributeTokens) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgDistributeTokens: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDistributeTokens: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SenderAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.SenderAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Amount", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Amount = append(m.Amount, types.Coin{})
			if err := m.Amount[len(m.Amount)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgDistributeTokensResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgDistributeTokensResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgDistributeTokensResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUpdateParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUpdateParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Params.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MsgUpdateParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgUpdateParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgUpdateParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTx
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
