// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             (unknown)
// source: juno/feeshare/v1/query.proto

package feesharev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	Query_FeeShares_FullMethodName           = "/juno.feeshare.v1.Query/FeeShares"
	Query_FeeShare_FullMethodName            = "/juno.feeshare.v1.Query/FeeShare"
	Query_Params_FullMethodName              = "/juno.feeshare.v1.Query/Params"
	Query_DeployerFeeShares_FullMethodName   = "/juno.feeshare.v1.Query/DeployerFeeShares"
	Query_WithdrawerFeeShares_FullMethodName = "/juno.feeshare.v1.Query/WithdrawerFeeShares"
)

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// Query defines the gRPC querier service.
type QueryClient interface {
	// FeeShares retrieves all registered FeeShares
	FeeShares(ctx context.Context, in *QueryFeeSharesRequest, opts ...grpc.CallOption) (*QueryFeeSharesResponse, error)
	// FeeShare retrieves a registered FeeShare for a given contract address
	FeeShare(ctx context.Context, in *QueryFeeShareRequest, opts ...grpc.CallOption) (*QueryFeeShareResponse, error)
	// Params retrieves the FeeShare module params
	Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error)
	// DeployerFeeShares retrieves all FeeShares that a given deployer has
	// registered
	DeployerFeeShares(ctx context.Context, in *QueryDeployerFeeSharesRequest, opts ...grpc.CallOption) (*QueryDeployerFeeSharesResponse, error)
	// WithdrawerFeeShares retrieves all FeeShares with a given withdrawer
	// address
	WithdrawerFeeShares(ctx context.Context, in *QueryWithdrawerFeeSharesRequest, opts ...grpc.CallOption) (*QueryWithdrawerFeeSharesResponse, error)
}

type queryClient struct {
	cc grpc.ClientConnInterface
}

func NewQueryClient(cc grpc.ClientConnInterface) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) FeeShares(ctx context.Context, in *QueryFeeSharesRequest, opts ...grpc.CallOption) (*QueryFeeSharesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryFeeSharesResponse)
	err := c.cc.Invoke(ctx, Query_FeeShares_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) FeeShare(ctx context.Context, in *QueryFeeShareRequest, opts ...grpc.CallOption) (*QueryFeeShareResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryFeeShareResponse)
	err := c.cc.Invoke(ctx, Query_FeeShare_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Params(ctx context.Context, in *QueryParamsRequest, opts ...grpc.CallOption) (*QueryParamsResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryParamsResponse)
	err := c.cc.Invoke(ctx, Query_Params_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) DeployerFeeShares(ctx context.Context, in *QueryDeployerFeeSharesRequest, opts ...grpc.CallOption) (*QueryDeployerFeeSharesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryDeployerFeeSharesResponse)
	err := c.cc.Invoke(ctx, Query_DeployerFeeShares_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) WithdrawerFeeShares(ctx context.Context, in *QueryWithdrawerFeeSharesRequest, opts ...grpc.CallOption) (*QueryWithdrawerFeeSharesResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(QueryWithdrawerFeeSharesResponse)
	err := c.cc.Invoke(ctx, Query_WithdrawerFeeShares_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
// All implementations must embed UnimplementedQueryServer
// for forward compatibility.
//
// Query defines the gRPC querier service.
type QueryServer interface {
	// FeeShares retrieves all registered FeeShares
	FeeShares(context.Context, *QueryFeeSharesRequest) (*QueryFeeSharesResponse, error)
	// FeeShare retrieves a registered FeeShare for a given contract address
	FeeShare(context.Context, *QueryFeeShareRequest) (*QueryFeeShareResponse, error)
	// Params retrieves the FeeShare module params
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
	// DeployerFeeShares retrieves all FeeShares that a given deployer has
	// registered
	DeployerFeeShares(context.Context, *QueryDeployerFeeSharesRequest) (*QueryDeployerFeeSharesResponse, error)
	// WithdrawerFeeShares retrieves all FeeShares with a given withdrawer
	// address
	WithdrawerFeeShares(context.Context, *QueryWithdrawerFeeSharesRequest) (*QueryWithdrawerFeeSharesResponse, error)
	mustEmbedUnimplementedQueryServer()
}

// UnimplementedQueryServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedQueryServer struct{}

func (UnimplementedQueryServer) FeeShares(context.Context, *QueryFeeSharesRequest) (*QueryFeeSharesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeShares not implemented")
}
func (UnimplementedQueryServer) FeeShare(context.Context, *QueryFeeShareRequest) (*QueryFeeShareResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FeeShare not implemented")
}
func (UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Params not implemented")
}
func (UnimplementedQueryServer) DeployerFeeShares(context.Context, *QueryDeployerFeeSharesRequest) (*QueryDeployerFeeSharesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeployerFeeShares not implemented")
}
func (UnimplementedQueryServer) WithdrawerFeeShares(context.Context, *QueryWithdrawerFeeSharesRequest) (*QueryWithdrawerFeeSharesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawerFeeShares not implemented")
}
func (UnimplementedQueryServer) mustEmbedUnimplementedQueryServer() {}
func (UnimplementedQueryServer) testEmbeddedByValue()               {}

// UnsafeQueryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to QueryServer will
// result in compilation errors.
type UnsafeQueryServer interface {
	mustEmbedUnimplementedQueryServer()
}

func RegisterQueryServer(s grpc.ServiceRegistrar, srv QueryServer) {
	// If the following call pancis, it indicates UnimplementedQueryServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&Query_ServiceDesc, srv)
}

func _Query_FeeShares_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeSharesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeShares(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_FeeShares_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeShares(ctx, req.(*QueryFeeSharesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_FeeShare_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryFeeShareRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).FeeShare(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_FeeShare_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).FeeShare(ctx, req.(*QueryFeeShareRequest))
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
		FullMethod: Query_Params_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Params(ctx, req.(*QueryParamsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_DeployerFeeShares_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryDeployerFeeSharesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).DeployerFeeShares(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_DeployerFeeShares_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).DeployerFeeShares(ctx, req.(*QueryDeployerFeeSharesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_WithdrawerFeeShares_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWithdrawerFeeSharesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WithdrawerFeeShares(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Query_WithdrawerFeeShares_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WithdrawerFeeShares(ctx, req.(*QueryWithdrawerFeeSharesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Query_ServiceDesc is the grpc.ServiceDesc for Query service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Query_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "juno.feeshare.v1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FeeShares",
			Handler:    _Query_FeeShares_Handler,
		},
		{
			MethodName: "FeeShare",
			Handler:    _Query_FeeShare_Handler,
		},
		{
			MethodName: "Params",
			Handler:    _Query_Params_Handler,
		},
		{
			MethodName: "DeployerFeeShares",
			Handler:    _Query_DeployerFeeShares_Handler,
		},
		{
			MethodName: "WithdrawerFeeShares",
			Handler:    _Query_WithdrawerFeeShares_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "juno/feeshare/v1/query.proto",
}
