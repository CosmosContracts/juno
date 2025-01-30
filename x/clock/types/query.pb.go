// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: juno/clock/v1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	query "github.com/cosmos/cosmos-sdk/types/query"
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

// QueryClockContracts is the request type to get all contracts.
type QueryClockContractsRequest struct {
	// pagination defines an optional pagination for the request.
	Pagination *query.PageRequest `protobuf:"bytes,1,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryClockContractsRequest) Reset()         { *m = QueryClockContractsRequest{} }
func (m *QueryClockContractsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryClockContractsRequest) ProtoMessage()    {}
func (*QueryClockContractsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{0}
}
func (m *QueryClockContractsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClockContractsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClockContractsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClockContractsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClockContractsRequest.Merge(m, src)
}
func (m *QueryClockContractsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryClockContractsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClockContractsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClockContractsRequest proto.InternalMessageInfo

func (m *QueryClockContractsRequest) GetPagination() *query.PageRequest {
	if m != nil {
		return m.Pagination
	}
	return nil
}

// QueryClockContractsResponse is the response type for the Query/ClockContracts RPC method.
type QueryClockContractsResponse struct {
	// clock_contracts are the clock contracts.
	ClockContracts []ClockContract `protobuf:"bytes,1,rep,name=clock_contracts,json=clockContracts,proto3" json:"clock_contracts"`
	// pagination defines the pagination in the response.
	Pagination *query.PageResponse `protobuf:"bytes,2,opt,name=pagination,proto3" json:"pagination,omitempty"`
}

func (m *QueryClockContractsResponse) Reset()         { *m = QueryClockContractsResponse{} }
func (m *QueryClockContractsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryClockContractsResponse) ProtoMessage()    {}
func (*QueryClockContractsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{1}
}
func (m *QueryClockContractsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClockContractsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClockContractsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClockContractsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClockContractsResponse.Merge(m, src)
}
func (m *QueryClockContractsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryClockContractsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClockContractsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClockContractsResponse proto.InternalMessageInfo

func (m *QueryClockContractsResponse) GetClockContracts() []ClockContract {
	if m != nil {
		return m.ClockContracts
	}
	return nil
}

func (m *QueryClockContractsResponse) GetPagination() *query.PageResponse {
	if m != nil {
		return m.Pagination
	}
	return nil
}

// QueryClockContract is the request type to get a single contract.
type QueryClockContractRequest struct {
	// contract_address is the address of the contract to query.
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
}

func (m *QueryClockContractRequest) Reset()         { *m = QueryClockContractRequest{} }
func (m *QueryClockContractRequest) String() string { return proto.CompactTextString(m) }
func (*QueryClockContractRequest) ProtoMessage()    {}
func (*QueryClockContractRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{2}
}
func (m *QueryClockContractRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClockContractRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClockContractRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClockContractRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClockContractRequest.Merge(m, src)
}
func (m *QueryClockContractRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryClockContractRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClockContractRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClockContractRequest proto.InternalMessageInfo

func (m *QueryClockContractRequest) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

// QueryClockContractResponse is the response type for the Query/ClockContract RPC method.
type QueryClockContractResponse struct {
	// contract is the clock contract.
	ClockContract ClockContract `protobuf:"bytes,1,opt,name=clock_contract,json=clockContract,proto3" json:"clock_contract"`
}

func (m *QueryClockContractResponse) Reset()         { *m = QueryClockContractResponse{} }
func (m *QueryClockContractResponse) String() string { return proto.CompactTextString(m) }
func (*QueryClockContractResponse) ProtoMessage()    {}
func (*QueryClockContractResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{3}
}
func (m *QueryClockContractResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryClockContractResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryClockContractResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryClockContractResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryClockContractResponse.Merge(m, src)
}
func (m *QueryClockContractResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryClockContractResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryClockContractResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryClockContractResponse proto.InternalMessageInfo

func (m *QueryClockContractResponse) GetClockContract() ClockContract {
	if m != nil {
		return m.ClockContract
	}
	return ClockContract{}
}

// QueryParams is the request type to get all module params.
type QueryParamsRequest struct {
}

func (m *QueryParamsRequest) Reset()         { *m = QueryParamsRequest{} }
func (m *QueryParamsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryParamsRequest) ProtoMessage()    {}
func (*QueryParamsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{4}
}
func (m *QueryParamsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsRequest.Merge(m, src)
}
func (m *QueryParamsRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsRequest proto.InternalMessageInfo

// QueryClockContractsResponse is the response type for the Query/ClockContracts RPC method.
type QueryParamsResponse struct {
	Params Params `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
}

func (m *QueryParamsResponse) Reset()         { *m = QueryParamsResponse{} }
func (m *QueryParamsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryParamsResponse) ProtoMessage()    {}
func (*QueryParamsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_7da208f579d775c8, []int{5}
}
func (m *QueryParamsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryParamsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryParamsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryParamsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryParamsResponse.Merge(m, src)
}
func (m *QueryParamsResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryParamsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryParamsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryParamsResponse proto.InternalMessageInfo

func (m *QueryParamsResponse) GetParams() Params {
	if m != nil {
		return m.Params
	}
	return Params{}
}

func init() {
	proto.RegisterType((*QueryClockContractsRequest)(nil), "juno.clock.v1.QueryClockContractsRequest")
	proto.RegisterType((*QueryClockContractsResponse)(nil), "juno.clock.v1.QueryClockContractsResponse")
	proto.RegisterType((*QueryClockContractRequest)(nil), "juno.clock.v1.QueryClockContractRequest")
	proto.RegisterType((*QueryClockContractResponse)(nil), "juno.clock.v1.QueryClockContractResponse")
	proto.RegisterType((*QueryParamsRequest)(nil), "juno.clock.v1.QueryParamsRequest")
	proto.RegisterType((*QueryParamsResponse)(nil), "juno.clock.v1.QueryParamsResponse")
}

func init() { proto.RegisterFile("juno/clock/v1/query.proto", fileDescriptor_7da208f579d775c8) }

var fileDescriptor_7da208f579d775c8 = []byte{
	// 574 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x94, 0x3f, 0x6f, 0xd3, 0x4e,
	0x18, 0xc7, 0xe3, 0x56, 0xbf, 0x4a, 0xbd, 0x2a, 0xe9, 0x8f, 0x23, 0x15, 0x89, 0x5b, 0x99, 0x62,
	0x09, 0x48, 0x83, 0xe2, 0x53, 0xc2, 0xd2, 0x95, 0x44, 0xa2, 0x1b, 0x84, 0xb0, 0xb1, 0x84, 0x8b,
	0x73, 0x32, 0x86, 0xc4, 0xe7, 0xfa, 0x2e, 0x11, 0x15, 0x62, 0x61, 0x62, 0x42, 0x48, 0xec, 0xcc,
	0x4c, 0x88, 0x81, 0x81, 0x97, 0xd0, 0xb1, 0x82, 0x85, 0x09, 0xa1, 0x04, 0x89, 0xb7, 0x81, 0x7c,
	0x7f, 0x92, 0x9c, 0x49, 0x69, 0x96, 0xaa, 0x7e, 0xfe, 0x7d, 0x3f, 0xcf, 0xf7, 0x39, 0x05, 0x94,
	0x9f, 0x8e, 0x22, 0x8a, 0xfc, 0x01, 0xf5, 0x9f, 0xa1, 0x71, 0x1d, 0x1d, 0x8f, 0x48, 0x72, 0xe2,
	0xc5, 0x09, 0xe5, 0x14, 0xe6, 0xd3, 0x94, 0x27, 0x52, 0xde, 0xb8, 0x6e, 0x57, 0x7d, 0xca, 0x86,
	0x94, 0xa1, 0x1e, 0x66, 0x44, 0xd6, 0xa1, 0x71, 0xbd, 0x47, 0x38, 0xae, 0xa3, 0x18, 0x07, 0x61,
	0x84, 0x79, 0x48, 0x23, 0xd9, 0x6a, 0x97, 0x65, 0x6d, 0x57, 0x7c, 0x21, 0xf9, 0xa1, 0x52, 0xbb,
	0x6a, 0x8c, 0x9e, 0xb0, 0x28, 0x69, 0x17, 0x03, 0x1a, 0x50, 0xd9, 0x94, 0xfe, 0xa7, 0xa2, 0x7b,
	0x01, 0xa5, 0xc1, 0x80, 0x20, 0x1c, 0x87, 0x08, 0x47, 0x11, 0xe5, 0x42, 0x4a, 0x0f, 0xbc, 0x84,
	0x87, 0x61, 0x44, 0x91, 0xf8, 0xab, 0x35, 0xcc, 0xa5, 0x02, 0x12, 0x11, 0x16, 0xea, 0xfa, 0xcc,
	0xc6, 0x72, 0x3f, 0x91, 0x72, 0xfb, 0xc0, 0x7e, 0x90, 0xd2, 0xb4, 0xd2, 0x58, 0x8b, 0x46, 0x3c,
	0xc1, 0x3e, 0x67, 0x1d, 0x72, 0x3c, 0x22, 0x8c, 0xc3, 0xbb, 0x00, 0xcc, 0x17, 0x2d, 0x59, 0xfb,
	0x56, 0x65, 0xab, 0x71, 0xc3, 0x53, 0xcb, 0xa5, 0xae, 0x78, 0x72, 0x15, 0xe5, 0x8a, 0xd7, 0xc6,
	0x01, 0x51, 0xbd, 0x9d, 0x85, 0x4e, 0xf7, 0x8b, 0x05, 0x76, 0x97, 0xca, 0xb0, 0x98, 0x46, 0x8c,
	0xc0, 0x36, 0xd8, 0x16, 0x50, 0x5d, 0x5f, 0xa7, 0x4a, 0xd6, 0xfe, 0x7a, 0x65, 0xab, 0xb1, 0xe7,
	0x19, 0x17, 0xf1, 0x8c, 0xfe, 0xe6, 0xe6, 0xe9, 0x8f, 0xab, 0xb9, 0x0f, 0xbf, 0x3f, 0x55, 0xad,
	0x4e, 0xc1, 0x37, 0x26, 0xc3, 0x23, 0x83, 0x7c, 0x4d, 0x90, 0xdf, 0xbc, 0x90, 0x5c, 0xe2, 0x18,
	0xe8, 0x8f, 0x41, 0xf9, 0x6f, 0x72, 0xed, 0x4f, 0x0b, 0xfc, 0xaf, 0x89, 0xbb, 0xb8, 0xdf, 0x4f,
	0x08, 0x63, 0xc2, 0xa5, 0xcd, 0x66, 0xe9, 0xeb, 0xe7, 0x5a, 0x51, 0xc9, 0xdd, 0x91, 0x99, 0x87,
	0x3c, 0x09, 0xa3, 0xa0, 0xb3, 0xad, 0x3b, 0x54, 0xd8, 0x1d, 0x2c, 0x3b, 0xc1, 0xcc, 0x9a, 0x7b,
	0xa0, 0x60, 0x5a, 0xa3, 0xce, 0xb0, 0xb2, 0x33, 0x79, 0xc3, 0x19, 0xb7, 0x08, 0xa0, 0x50, 0x6b,
	0xe3, 0x04, 0x0f, 0xf5, 0xa1, 0xdd, 0xfb, 0xe0, 0xb2, 0x11, 0x55, 0xe2, 0x87, 0x60, 0x23, 0x16,
	0x11, 0x25, 0xba, 0x93, 0x11, 0x95, 0xe5, 0x8b, 0x6a, 0xaa, 0xbe, 0xf1, 0x71, 0x1d, 0xfc, 0x27,
	0x26, 0xc2, 0x37, 0x16, 0x28, 0x98, 0x67, 0x87, 0x07, 0x99, 0x31, 0xe7, 0xbf, 0x40, 0xbb, 0xba,
	0x4a, 0xa9, 0xa4, 0x75, 0xaf, 0xbf, 0x4e, 0x11, 0x5e, 0x7d, 0xfb, 0xf5, 0x6e, 0xcd, 0x86, 0x25,
	0x94, 0x79, 0xf4, 0x33, 0xf5, 0xf7, 0x16, 0xc8, 0x1b, 0x13, 0x60, 0xe5, 0x42, 0x11, 0x8d, 0x73,
	0xb0, 0x42, 0xa5, 0xa2, 0x39, 0x9c, 0xd3, 0xd4, 0xe0, 0xad, 0xf3, 0x68, 0xd0, 0x8b, 0xec, 0x03,
	0x7a, 0x09, 0x19, 0xd8, 0x90, 0xc6, 0xc2, 0x6b, 0xcb, 0xe4, 0x8c, 0xcb, 0xd9, 0xee, 0xbf, 0x4a,
	0x14, 0x8a, 0x3b, 0x47, 0xb9, 0x02, 0x77, 0x32, 0x28, 0xf2, 0x60, 0xcd, 0xa3, 0xd3, 0x89, 0x63,
	0x9d, 0x4d, 0x1c, 0xeb, 0xe7, 0xc4, 0xb1, 0xde, 0x4e, 0x9d, 0xdc, 0xd9, 0xd4, 0xc9, 0x7d, 0x9f,
	0x3a, 0xb9, 0x47, 0xb5, 0x20, 0xe4, 0x4f, 0x46, 0x3d, 0xcf, 0xa7, 0x43, 0xd4, 0x12, 0x2f, 0x7a,
	0x66, 0xbd, 0x1c, 0xf5, 0x5c, 0x0d, 0xe3, 0x27, 0x31, 0x61, 0xbd, 0x0d, 0xf1, 0xc3, 0x72, 0xfb,
	0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5a, 0xb8, 0xae, 0x5d, 0x67, 0x05, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// ClockContracts
	ClockContracts(ctx context.Context, in *QueryClockContractsRequest, opts ...grpc.CallOption) (*QueryClockContractsResponse, error)
	// ClockContract
	ClockContract(ctx context.Context, in *QueryClockContractRequest, opts ...grpc.CallOption) (*QueryClockContractResponse, error)
	// Params
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) ClockContracts(ctx context.Context, in *QueryClockContractsRequest, opts ...grpc.CallOption) (*QueryClockContractsResponse, error) {
	out := new(QueryClockContractsResponse)
	err := c.cc.Invoke(ctx, "/juno.clock.v1.Query/ClockContracts", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) ClockContract(ctx context.Context, in *QueryClockContractRequest, opts ...grpc.CallOption) (*QueryClockContractResponse, error) {
	out := new(QueryClockContractResponse)
	err := c.cc.Invoke(ctx, "/juno.clock.v1.Query/ClockContract", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, "/juno.clock.v1.Query/Params", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// ClockContracts
	ClockContracts(context.Context, *QueryClockContractsRequest) (*QueryClockContractsResponse, error)
	// ClockContract
	ClockContract(context.Context, *QueryClockContractRequest) (*QueryClockContractResponse, error)
	// Params
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) ClockContracts(ctx context.Context, req *QueryClockContractsRequest) (*QueryClockContractsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClockContracts not implemented")
}
func (*UnimplementedQueryServer) ClockContract(ctx context.Context, req *QueryClockContractRequest) (*QueryClockContractResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClockContract not implemented")
}
func (*UnimplementedQueryServer) Params(ctx context.Context, req *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_ClockContracts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryClockContractsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ClockContracts(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/juno.clock.v1.Query/ClockContracts",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ClockContracts(ctx, req.(*QueryClockContractsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_ClockContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryClockContractRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).ClockContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/juno.clock.v1.Query/ClockContract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).ClockContract(ctx, req.(*QueryClockContractRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Params_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryParamsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Params(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/juno.clock.v1.Query/Params",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var Query_serviceDesc = _Query_serviceDesc
var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "juno.clock.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ClockContracts",
			Handler:    _Query_ClockContracts_Handler,
		},
		{
			MethodName: "ClockContract",
			Handler:    _Query_ClockContract_Handler,
		},
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "juno/clock/v1/query.proto",
}

func (m *QueryClockContractsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClockContractsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClockContractsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryClockContractsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClockContractsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClockContractsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Pagination != nil {
		{
			size, err := m.Pagination.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintQuery(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.ClockContracts) > 0 {
		for iNdEx := len(m.ClockContracts) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ClockContracts[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryClockContractRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClockContractRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClockContractRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryClockContractResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryClockContractResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryClockContractResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.ClockContract.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryParamsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryParamsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryParamsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryParamsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryClockContractsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Pagination != nil {
		l = m.Pagination.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryClockContractsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.ClockContracts) > 0 {
		for _, e := range m.ClockContracts {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	if m.Pagination != nil {
		l = m.Pagination.Size()
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryClockContractRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryClockContractResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ClockContract.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryParamsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryParamsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Params.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryClockContractsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryClockContractsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClockContractsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pagination", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pagination == nil {
				m.Pagination = &query.PageRequest{}
			}
			if err := m.Pagination.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryClockContractsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryClockContractsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClockContractsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClockContracts", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ClockContracts = append(m.ClockContracts, ClockContract{})
			if err := m.ClockContracts[len(m.ClockContracts)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pagination", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Pagination == nil {
				m.Pagination = &query.PageResponse{}
			}
			if err := m.Pagination.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryClockContractRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryClockContractRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClockContractRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryClockContractResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryClockContractResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryClockContractResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ClockContract", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ClockContract.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryParamsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryParamsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func (m *QueryParamsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
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
			return fmt.Errorf("proto: QueryParamsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryParamsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Params", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
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
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
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
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
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
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
					return 0, ErrIntOverflowQuery
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
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
