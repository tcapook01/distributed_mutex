// tokenring.proto

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.3
// source: tokenring.proto

package proto

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
	TokenRing_RequestToken_FullMethodName = "/proto.TokenRing/RequestToken"
	TokenRing_PassToken_FullMethodName    = "/proto.TokenRing/PassToken"
)

// TokenRingClient is the client API for TokenRing service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TokenRingClient interface {
	RequestToken(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	PassToken(ctx context.Context, in *TokenMessage, opts ...grpc.CallOption) (*Response, error)
}

type tokenRingClient struct {
	cc grpc.ClientConnInterface
}

func NewTokenRingClient(cc grpc.ClientConnInterface) TokenRingClient {
	return &tokenRingClient{cc}
}

func (c *tokenRingClient) RequestToken(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, TokenRing_RequestToken_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tokenRingClient) PassToken(ctx context.Context, in *TokenMessage, opts ...grpc.CallOption) (*Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Response)
	err := c.cc.Invoke(ctx, TokenRing_PassToken_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TokenRingServer is the server API for TokenRing service.
// All implementations must embed UnimplementedTokenRingServer
// for forward compatibility.
type TokenRingServer interface {
	RequestToken(context.Context, *Request) (*Response, error)
	PassToken(context.Context, *TokenMessage) (*Response, error)
	mustEmbedUnimplementedTokenRingServer()
}

// UnimplementedTokenRingServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTokenRingServer struct{}

func (UnimplementedTokenRingServer) RequestToken(context.Context, *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestToken not implemented")
}
func (UnimplementedTokenRingServer) PassToken(context.Context, *TokenMessage) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PassToken not implemented")
}
func (UnimplementedTokenRingServer) mustEmbedUnimplementedTokenRingServer() {}
func (UnimplementedTokenRingServer) testEmbeddedByValue()                   {}

// UnsafeTokenRingServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TokenRingServer will
// result in compilation errors.
type UnsafeTokenRingServer interface {
	mustEmbedUnimplementedTokenRingServer()
}

func RegisterTokenRingServer(s grpc.ServiceRegistrar, srv TokenRingServer) {
	// If the following call pancis, it indicates UnimplementedTokenRingServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TokenRing_ServiceDesc, srv)
}

func _TokenRing_RequestToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenRingServer).RequestToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TokenRing_RequestToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenRingServer).RequestToken(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _TokenRing_PassToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenRingServer).PassToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TokenRing_PassToken_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenRingServer).PassToken(ctx, req.(*TokenMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// TokenRing_ServiceDesc is the grpc.ServiceDesc for TokenRing service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TokenRing_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "proto.TokenRing",
	HandlerType: (*TokenRingServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RequestToken",
			Handler:    _TokenRing_RequestToken_Handler,
		},
		{
			MethodName: "PassToken",
			Handler:    _TokenRing_PassToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tokenring.proto",
}