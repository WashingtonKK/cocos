// Copyright (c) Ultraviolet
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.2
// source: manager/manager.proto

package manager

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ManagerService_Run_FullMethodName = "/manager.ManagerService/Run"
)

// ManagerServiceClient is the client API for ManagerService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ManagerServiceClient interface {
	Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*RunResponse, error)
}

type managerServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewManagerServiceClient(cc grpc.ClientConnInterface) ManagerServiceClient {
	return &managerServiceClient{cc}
}

func (c *managerServiceClient) Run(ctx context.Context, in *RunRequest, opts ...grpc.CallOption) (*RunResponse, error) {
	out := new(RunResponse)
	err := c.cc.Invoke(ctx, ManagerService_Run_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ManagerServiceServer is the server API for ManagerService service.
// All implementations must embed UnimplementedManagerServiceServer
// for forward compatibility
type ManagerServiceServer interface {
	Run(context.Context, *RunRequest) (*RunResponse, error)
	mustEmbedUnimplementedManagerServiceServer()
}

// UnimplementedManagerServiceServer must be embedded to have forward compatible implementations.
type UnimplementedManagerServiceServer struct {
}

func (UnimplementedManagerServiceServer) Run(context.Context, *RunRequest) (*RunResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Run not implemented")
}
func (UnimplementedManagerServiceServer) mustEmbedUnimplementedManagerServiceServer() {}

// UnsafeManagerServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ManagerServiceServer will
// result in compilation errors.
type UnsafeManagerServiceServer interface {
	mustEmbedUnimplementedManagerServiceServer()
}

func RegisterManagerServiceServer(s grpc.ServiceRegistrar, srv ManagerServiceServer) {
	s.RegisterService(&ManagerService_ServiceDesc, srv)
}

func _ManagerService_Run_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RunRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ManagerServiceServer).Run(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ManagerService_Run_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ManagerServiceServer).Run(ctx, req.(*RunRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ManagerService_ServiceDesc is the grpc.ServiceDesc for ManagerService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ManagerService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "manager.ManagerService",
	HandlerType: (*ManagerServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Run",
			Handler:    _ManagerService_Run_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "manager/manager.proto",
}
