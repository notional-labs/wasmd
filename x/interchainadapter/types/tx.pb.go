// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: interchainadapter/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	grpc1 "github.com/gogo/protobuf/grpc"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

func init() { proto.RegisterFile("interchainadapter/tx.proto", fileDescriptor_d2a3ccc31c423c76) }

var fileDescriptor_d2a3ccc31c423c76 = []byte{
	// 139 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xca, 0xcc, 0x2b, 0x49,
	0x2d, 0x4a, 0xce, 0x48, 0xcc, 0xcc, 0x4b, 0x4c, 0x49, 0x2c, 0x28, 0x49, 0x2d, 0xd2, 0x2f, 0xa9,
	0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xd2, 0xcf, 0xc9, 0x4c, 0xc9, 0x4f, 0xcb, 0xcc, 0x4b,
	0xcc, 0x4b, 0x4e, 0xd5, 0xc3, 0x50, 0x87, 0x29, 0x62, 0xc4, 0xca, 0xc5, 0xec, 0x5b, 0x9c, 0xee,
	0x14, 0x7b, 0xe2, 0x91, 0x1c, 0xe3, 0x85, 0x47, 0x72, 0x8c, 0x0f, 0x1e, 0xc9, 0x31, 0x4e, 0x78,
	0x2c, 0xc7, 0x70, 0xe1, 0xb1, 0x1c, 0xc3, 0x8d, 0xc7, 0x72, 0x0c, 0x51, 0xce, 0xe9, 0x99, 0x25,
	0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9, 0xc8, 0x86, 0xeb, 0x23, 0x8c, 0xd2, 0x85, 0xb9, 0xa2,
	0x42, 0x1f, 0x8b, 0xcb, 0x2a, 0x0b, 0x52, 0x8b, 0x93, 0xd8, 0xc0, 0xae, 0x33, 0x06, 0x04, 0x00,
	0x00, 0xff, 0xff, 0xc5, 0xda, 0x78, 0x46, 0xbb, 0x00, 0x00, 0x00,
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
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "lidofinance.interchainadapter.interchainadapter.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "interchainadapter/tx.proto",
}
