// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.22.3
// source: proto/lnrpc/neutrinorpc/neutrino.proto

package neutrinorpc

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
	NeutrinoKit_Status_FullMethodName         = "/neutrinorpc.NeutrinoKit/Status"
	NeutrinoKit_AddPeer_FullMethodName        = "/neutrinorpc.NeutrinoKit/AddPeer"
	NeutrinoKit_DisconnectPeer_FullMethodName = "/neutrinorpc.NeutrinoKit/DisconnectPeer"
	NeutrinoKit_IsBanned_FullMethodName       = "/neutrinorpc.NeutrinoKit/IsBanned"
	NeutrinoKit_GetBlockHeader_FullMethodName = "/neutrinorpc.NeutrinoKit/GetBlockHeader"
	NeutrinoKit_GetBlock_FullMethodName       = "/neutrinorpc.NeutrinoKit/GetBlock"
	NeutrinoKit_GetCFilter_FullMethodName     = "/neutrinorpc.NeutrinoKit/GetCFilter"
	NeutrinoKit_GetBlockHash_FullMethodName   = "/neutrinorpc.NeutrinoKit/GetBlockHash"
)

// NeutrinoKitClient is the client API for NeutrinoKit service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type NeutrinoKitClient interface {
	// Status returns the status of the light client neutrino instance,
	// along with height and hash of the best block, and a list of connected
	// peers.
	Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error)
	// AddPeer adds a new peer that has already been connected to the server.
	AddPeer(ctx context.Context, in *AddPeerRequest, opts ...grpc.CallOption) (*AddPeerResponse, error)
	// DisconnectPeer disconnects a peer by target address. Both outbound and
	// inbound nodes will be searched for the target node. An error message will
	// be returned if the peer was not found.
	DisconnectPeer(ctx context.Context, in *DisconnectPeerRequest, opts ...grpc.CallOption) (*DisconnectPeerResponse, error)
	// IsBanned returns true if the peer is banned, otherwise false.
	IsBanned(ctx context.Context, in *IsBannedRequest, opts ...grpc.CallOption) (*IsBannedResponse, error)
	// GetBlockHeader returns a block header with a particular block hash.
	GetBlockHeader(ctx context.Context, in *GetBlockHeaderRequest, opts ...grpc.CallOption) (*GetBlockHeaderResponse, error)
	// GetBlock returns a block with a particular block hash.
	GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error)
	// GetCFilter returns a compact filter from a block.
	GetCFilter(ctx context.Context, in *GetCFilterRequest, opts ...grpc.CallOption) (*GetCFilterResponse, error)
	// GetBlockHash returns the header hash of a block at a given height.
	GetBlockHash(ctx context.Context, in *GetBlockHashRequest, opts ...grpc.CallOption) (*GetBlockHashResponse, error)
}

type neutrinoKitClient struct {
	cc grpc.ClientConnInterface
}

func NewNeutrinoKitClient(cc grpc.ClientConnInterface) NeutrinoKitClient {
	return &neutrinoKitClient{cc}
}

func (c *neutrinoKitClient) Status(ctx context.Context, in *StatusRequest, opts ...grpc.CallOption) (*StatusResponse, error) {
	out := new(StatusResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_Status_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) AddPeer(ctx context.Context, in *AddPeerRequest, opts ...grpc.CallOption) (*AddPeerResponse, error) {
	out := new(AddPeerResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_AddPeer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) DisconnectPeer(ctx context.Context, in *DisconnectPeerRequest, opts ...grpc.CallOption) (*DisconnectPeerResponse, error) {
	out := new(DisconnectPeerResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_DisconnectPeer_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) IsBanned(ctx context.Context, in *IsBannedRequest, opts ...grpc.CallOption) (*IsBannedResponse, error) {
	out := new(IsBannedResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_IsBanned_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) GetBlockHeader(ctx context.Context, in *GetBlockHeaderRequest, opts ...grpc.CallOption) (*GetBlockHeaderResponse, error) {
	out := new(GetBlockHeaderResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_GetBlockHeader_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) GetBlock(ctx context.Context, in *GetBlockRequest, opts ...grpc.CallOption) (*GetBlockResponse, error) {
	out := new(GetBlockResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_GetBlock_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) GetCFilter(ctx context.Context, in *GetCFilterRequest, opts ...grpc.CallOption) (*GetCFilterResponse, error) {
	out := new(GetCFilterResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_GetCFilter_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *neutrinoKitClient) GetBlockHash(ctx context.Context, in *GetBlockHashRequest, opts ...grpc.CallOption) (*GetBlockHashResponse, error) {
	out := new(GetBlockHashResponse)
	err := c.cc.Invoke(ctx, NeutrinoKit_GetBlockHash_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NeutrinoKitServer is the server API for NeutrinoKit service.
// All implementations must embed UnimplementedNeutrinoKitServer
// for forward compatibility
type NeutrinoKitServer interface {
	// Status returns the status of the light client neutrino instance,
	// along with height and hash of the best block, and a list of connected
	// peers.
	Status(context.Context, *StatusRequest) (*StatusResponse, error)
	// AddPeer adds a new peer that has already been connected to the server.
	AddPeer(context.Context, *AddPeerRequest) (*AddPeerResponse, error)
	// DisconnectPeer disconnects a peer by target address. Both outbound and
	// inbound nodes will be searched for the target node. An error message will
	// be returned if the peer was not found.
	DisconnectPeer(context.Context, *DisconnectPeerRequest) (*DisconnectPeerResponse, error)
	// IsBanned returns true if the peer is banned, otherwise false.
	IsBanned(context.Context, *IsBannedRequest) (*IsBannedResponse, error)
	// GetBlockHeader returns a block header with a particular block hash.
	GetBlockHeader(context.Context, *GetBlockHeaderRequest) (*GetBlockHeaderResponse, error)
	// GetBlock returns a block with a particular block hash.
	GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error)
	// GetCFilter returns a compact filter from a block.
	GetCFilter(context.Context, *GetCFilterRequest) (*GetCFilterResponse, error)
	// GetBlockHash returns the header hash of a block at a given height.
	GetBlockHash(context.Context, *GetBlockHashRequest) (*GetBlockHashResponse, error)
	mustEmbedUnimplementedNeutrinoKitServer()
}

// UnimplementedNeutrinoKitServer must be embedded to have forward compatible implementations.
type UnimplementedNeutrinoKitServer struct {
}

func (UnimplementedNeutrinoKitServer) Status(context.Context, *StatusRequest) (*StatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Status not implemented")
}
func (UnimplementedNeutrinoKitServer) AddPeer(context.Context, *AddPeerRequest) (*AddPeerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddPeer not implemented")
}
func (UnimplementedNeutrinoKitServer) DisconnectPeer(context.Context, *DisconnectPeerRequest) (*DisconnectPeerResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DisconnectPeer not implemented")
}
func (UnimplementedNeutrinoKitServer) IsBanned(context.Context, *IsBannedRequest) (*IsBannedResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsBanned not implemented")
}
func (UnimplementedNeutrinoKitServer) GetBlockHeader(context.Context, *GetBlockHeaderRequest) (*GetBlockHeaderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockHeader not implemented")
}
func (UnimplementedNeutrinoKitServer) GetBlock(context.Context, *GetBlockRequest) (*GetBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (UnimplementedNeutrinoKitServer) GetCFilter(context.Context, *GetCFilterRequest) (*GetCFilterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCFilter not implemented")
}
func (UnimplementedNeutrinoKitServer) GetBlockHash(context.Context, *GetBlockHashRequest) (*GetBlockHashResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlockHash not implemented")
}
func (UnimplementedNeutrinoKitServer) mustEmbedUnimplementedNeutrinoKitServer() {}

// UnsafeNeutrinoKitServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to NeutrinoKitServer will
// result in compilation errors.
type UnsafeNeutrinoKitServer interface {
	mustEmbedUnimplementedNeutrinoKitServer()
}

func RegisterNeutrinoKitServer(s grpc.ServiceRegistrar, srv NeutrinoKitServer) {
	s.RegisterService(&NeutrinoKit_ServiceDesc, srv)
}

func _NeutrinoKit_Status_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).Status(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_Status_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).Status(ctx, req.(*StatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_AddPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddPeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).AddPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_AddPeer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).AddPeer(ctx, req.(*AddPeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_DisconnectPeer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DisconnectPeerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).DisconnectPeer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_DisconnectPeer_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).DisconnectPeer(ctx, req.(*DisconnectPeerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_IsBanned_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IsBannedRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).IsBanned(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_IsBanned_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).IsBanned(ctx, req.(*IsBannedRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_GetBlockHeader_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockHeaderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).GetBlockHeader(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_GetBlockHeader_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).GetBlockHeader(ctx, req.(*GetBlockHeaderRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_GetBlock_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).GetBlock(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_GetBlock_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).GetBlock(ctx, req.(*GetBlockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_GetCFilter_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCFilterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).GetCFilter(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_GetCFilter_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).GetCFilter(ctx, req.(*GetCFilterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NeutrinoKit_GetBlockHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetBlockHashRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NeutrinoKitServer).GetBlockHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: NeutrinoKit_GetBlockHash_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NeutrinoKitServer).GetBlockHash(ctx, req.(*GetBlockHashRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// NeutrinoKit_ServiceDesc is the grpc.ServiceDesc for NeutrinoKit service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var NeutrinoKit_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "neutrinorpc.NeutrinoKit",
	HandlerType: (*NeutrinoKitServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Status",
			Handler:    _NeutrinoKit_Status_Handler,
		},
		{
			MethodName: "AddPeer",
			Handler:    _NeutrinoKit_AddPeer_Handler,
		},
		{
			MethodName: "DisconnectPeer",
			Handler:    _NeutrinoKit_DisconnectPeer_Handler,
		},
		{
			MethodName: "IsBanned",
			Handler:    _NeutrinoKit_IsBanned_Handler,
		},
		{
			MethodName: "GetBlockHeader",
			Handler:    _NeutrinoKit_GetBlockHeader_Handler,
		},
		{
			MethodName: "GetBlock",
			Handler:    _NeutrinoKit_GetBlock_Handler,
		},
		{
			MethodName: "GetCFilter",
			Handler:    _NeutrinoKit_GetCFilter_Handler,
		},
		{
			MethodName: "GetBlockHash",
			Handler:    _NeutrinoKit_GetBlockHash_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/lnrpc/neutrinorpc/neutrino.proto",
}
