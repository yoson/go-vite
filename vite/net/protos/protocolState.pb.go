// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protocolState.proto

package protos

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

type ProtocolState_PeerStatus int32

const (
	ProtocolState_Connected    ProtocolState_PeerStatus = 0
	ProtocolState_Disconnected ProtocolState_PeerStatus = 1
)

var ProtocolState_PeerStatus_name = map[int32]string{
	0: "Connected",
	1: "Disconnected",
}

var ProtocolState_PeerStatus_value = map[string]int32{
	"Connected":    0,
	"Disconnected": 1,
}

func (x ProtocolState_PeerStatus) String() string {
	return proto.EnumName(ProtocolState_PeerStatus_name, int32(x))
}

func (ProtocolState_PeerStatus) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8ffcde59fe093887, []int{0, 0}
}

type ProtocolState struct {
	Peers                []*ProtocolState_Peer `protobuf:"bytes,1,rep,name=Peers,proto3" json:"Peers,omitempty"`
	Patch                bool                  `protobuf:"varint,2,opt,name=Patch,proto3" json:"Patch,omitempty"`
	Head                 []byte                `protobuf:"bytes,3,opt,name=Head,proto3" json:"Head,omitempty"`
	Height               uint64                `protobuf:"varint,4,opt,name=Height,proto3" json:"Height,omitempty"`
	Timestamp            int64                 `protobuf:"varint,10,opt,name=Timestamp,proto3" json:"Timestamp,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *ProtocolState) Reset()         { *m = ProtocolState{} }
func (m *ProtocolState) String() string { return proto.CompactTextString(m) }
func (*ProtocolState) ProtoMessage()    {}
func (*ProtocolState) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ffcde59fe093887, []int{0}
}

func (m *ProtocolState) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProtocolState.Unmarshal(m, b)
}
func (m *ProtocolState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProtocolState.Marshal(b, m, deterministic)
}
func (m *ProtocolState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProtocolState.Merge(m, src)
}
func (m *ProtocolState) XXX_Size() int {
	return xxx_messageInfo_ProtocolState.Size(m)
}
func (m *ProtocolState) XXX_DiscardUnknown() {
	xxx_messageInfo_ProtocolState.DiscardUnknown(m)
}

var xxx_messageInfo_ProtocolState proto.InternalMessageInfo

func (m *ProtocolState) GetPeers() []*ProtocolState_Peer {
	if m != nil {
		return m.Peers
	}
	return nil
}

func (m *ProtocolState) GetPatch() bool {
	if m != nil {
		return m.Patch
	}
	return false
}

func (m *ProtocolState) GetHead() []byte {
	if m != nil {
		return m.Head
	}
	return nil
}

func (m *ProtocolState) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ProtocolState) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type ProtocolState_Peer struct {
	ID                   []byte                   `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Status               ProtocolState_PeerStatus `protobuf:"varint,3,opt,name=Status,proto3,enum=protos.ProtocolState_PeerStatus" json:"Status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                 `json:"-"`
	XXX_unrecognized     []byte                   `json:"-"`
	XXX_sizecache        int32                    `json:"-"`
}

func (m *ProtocolState_Peer) Reset()         { *m = ProtocolState_Peer{} }
func (m *ProtocolState_Peer) String() string { return proto.CompactTextString(m) }
func (*ProtocolState_Peer) ProtoMessage()    {}
func (*ProtocolState_Peer) Descriptor() ([]byte, []int) {
	return fileDescriptor_8ffcde59fe093887, []int{0, 0}
}

func (m *ProtocolState_Peer) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProtocolState_Peer.Unmarshal(m, b)
}
func (m *ProtocolState_Peer) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProtocolState_Peer.Marshal(b, m, deterministic)
}
func (m *ProtocolState_Peer) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProtocolState_Peer.Merge(m, src)
}
func (m *ProtocolState_Peer) XXX_Size() int {
	return xxx_messageInfo_ProtocolState_Peer.Size(m)
}
func (m *ProtocolState_Peer) XXX_DiscardUnknown() {
	xxx_messageInfo_ProtocolState_Peer.DiscardUnknown(m)
}

var xxx_messageInfo_ProtocolState_Peer proto.InternalMessageInfo

func (m *ProtocolState_Peer) GetID() []byte {
	if m != nil {
		return m.ID
	}
	return nil
}

func (m *ProtocolState_Peer) GetStatus() ProtocolState_PeerStatus {
	if m != nil {
		return m.Status
	}
	return ProtocolState_Connected
}

func init() {
	proto.RegisterEnum("protos.ProtocolState_PeerStatus", ProtocolState_PeerStatus_name, ProtocolState_PeerStatus_value)
	proto.RegisterType((*ProtocolState)(nil), "protos.ProtocolState")
	proto.RegisterType((*ProtocolState_Peer)(nil), "protos.ProtocolState.Peer")
}

func init() { proto.RegisterFile("protocolState.proto", fileDescriptor_8ffcde59fe093887) }

var fileDescriptor_8ffcde59fe093887 = []byte{
	// 232 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x90, 0x41, 0x4f, 0x03, 0x21,
	0x10, 0x85, 0x85, 0xa5, 0x1b, 0x3b, 0xb6, 0x4d, 0x33, 0x1a, 0x43, 0x1a, 0x0f, 0xa4, 0x27, 0x2e,
	0x6e, 0x4c, 0xbd, 0x78, 0x77, 0x0f, 0xed, 0x8d, 0xa0, 0x7f, 0x00, 0xe9, 0xc4, 0x6e, 0x62, 0x4b,
	0x53, 0xf0, 0x0f, 0xf9, 0x4b, 0xcd, 0x42, 0xe3, 0xea, 0xc1, 0x13, 0xbc, 0xc7, 0xc7, 0xbc, 0x97,
	0x81, 0xeb, 0xe3, 0x29, 0xa4, 0xe0, 0xc3, 0xc7, 0x4b, 0x72, 0x89, 0x9a, 0xac, 0xb0, 0xce, 0x47,
	0x5c, 0x7e, 0x71, 0x98, 0x9a, 0xdf, 0xef, 0xf8, 0x00, 0x23, 0x43, 0x74, 0x8a, 0x92, 0xa9, 0x4a,
	0x5f, 0xad, 0x16, 0xe5, 0x43, 0x6c, 0xfe, 0x50, 0x4d, 0x8f, 0xd8, 0x02, 0xe2, 0x0d, 0x8c, 0x8c,
	0x4b, 0x7e, 0x27, 0xb9, 0x62, 0xfa, 0xd2, 0x16, 0x81, 0x08, 0x62, 0x4d, 0x6e, 0x2b, 0x2b, 0xc5,
	0xf4, 0xc4, 0xe6, 0x3b, 0xde, 0x42, 0xbd, 0xa6, 0xee, 0x7d, 0x97, 0xa4, 0x50, 0x4c, 0x0b, 0x7b,
	0x56, 0x78, 0x07, 0xe3, 0xd7, 0x6e, 0x4f, 0x31, 0xb9, 0xfd, 0x51, 0x82, 0x62, 0xba, 0xb2, 0x83,
	0xb1, 0x30, 0x20, 0xfa, 0x20, 0x9c, 0x01, 0xdf, 0xb4, 0x92, 0xe5, 0x79, 0x7c, 0xd3, 0xe2, 0x13,
	0xd4, 0x7d, 0x99, 0xcf, 0x98, 0x33, 0x66, 0x2b, 0xf5, 0x7f, 0xd5, 0xc2, 0xd9, 0x33, 0xbf, 0xbc,
	0x07, 0x18, 0x5c, 0x9c, 0xc2, 0xf8, 0x39, 0x1c, 0x0e, 0xe4, 0x13, 0x6d, 0xe7, 0x17, 0x38, 0x87,
	0x49, 0xdb, 0x45, 0xff, 0xe3, 0xb0, 0xb7, 0xb2, 0xac, 0xc7, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x8e, 0x69, 0xc3, 0x3b, 0x4a, 0x01, 0x00, 0x00,
}
