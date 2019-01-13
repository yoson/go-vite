// Code generated by protoc-gen-go. DO NOT EDIT.
// source: vitepb/snapshot_additional_list.proto

package vitepb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type HashHeight struct {
	Hash                 []byte   `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	Height               uint64   `protobuf:"varint,2,opt,name=height,proto3" json:"height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HashHeight) Reset()         { *m = HashHeight{} }
func (m *HashHeight) String() string { return proto.CompactTextString(m) }
func (*HashHeight) ProtoMessage()    {}
func (*HashHeight) Descriptor() ([]byte, []int) {
	return fileDescriptor_snapshot_additional_list_8011c913dc472a9e, []int{0}
}
func (m *HashHeight) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HashHeight.Unmarshal(m, b)
}
func (m *HashHeight) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HashHeight.Marshal(b, m, deterministic)
}
func (dst *HashHeight) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HashHeight.Merge(dst, src)
}
func (m *HashHeight) XXX_Size() int {
	return xxx_messageInfo_HashHeight.Size(m)
}
func (m *HashHeight) XXX_DiscardUnknown() {
	xxx_messageInfo_HashHeight.DiscardUnknown(m)
}

var xxx_messageInfo_HashHeight proto.InternalMessageInfo

func (m *HashHeight) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *HashHeight) GetHeight() uint64 {
	if m != nil {
		return m.Height
	}
	return 0
}

type SnapshotAdditionalItem struct {
	Quota                uint64      `protobuf:"varint,1,opt,name=quota,proto3" json:"quota,omitempty"`
	AggregateQuota       uint64      `protobuf:"varint,2,opt,name=aggregateQuota,proto3" json:"aggregateQuota,omitempty"`
	SnapshotHashHeight   *HashHeight `protobuf:"bytes,3,opt,name=snapshotHashHeight,proto3" json:"snapshotHashHeight,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *SnapshotAdditionalItem) Reset()         { *m = SnapshotAdditionalItem{} }
func (m *SnapshotAdditionalItem) String() string { return proto.CompactTextString(m) }
func (*SnapshotAdditionalItem) ProtoMessage()    {}
func (*SnapshotAdditionalItem) Descriptor() ([]byte, []int) {
	return fileDescriptor_snapshot_additional_list_8011c913dc472a9e, []int{1}
}
func (m *SnapshotAdditionalItem) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotAdditionalItem.Unmarshal(m, b)
}
func (m *SnapshotAdditionalItem) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotAdditionalItem.Marshal(b, m, deterministic)
}
func (dst *SnapshotAdditionalItem) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotAdditionalItem.Merge(dst, src)
}
func (m *SnapshotAdditionalItem) XXX_Size() int {
	return xxx_messageInfo_SnapshotAdditionalItem.Size(m)
}
func (m *SnapshotAdditionalItem) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotAdditionalItem.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotAdditionalItem proto.InternalMessageInfo

func (m *SnapshotAdditionalItem) GetQuota() uint64 {
	if m != nil {
		return m.Quota
	}
	return 0
}

func (m *SnapshotAdditionalItem) GetAggregateQuota() uint64 {
	if m != nil {
		return m.AggregateQuota
	}
	return 0
}

func (m *SnapshotAdditionalItem) GetSnapshotHashHeight() *HashHeight {
	if m != nil {
		return m.SnapshotHashHeight
	}
	return nil
}

type SnapshotAdditionalFragment struct {
	List                 []*SnapshotAdditionalItem `protobuf:"bytes,1,rep,name=list,proto3" json:"list,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                  `json:"-"`
	XXX_unrecognized     []byte                    `json:"-"`
	XXX_sizecache        int32                     `json:"-"`
}

func (m *SnapshotAdditionalFragment) Reset()         { *m = SnapshotAdditionalFragment{} }
func (m *SnapshotAdditionalFragment) String() string { return proto.CompactTextString(m) }
func (*SnapshotAdditionalFragment) ProtoMessage()    {}
func (*SnapshotAdditionalFragment) Descriptor() ([]byte, []int) {
	return fileDescriptor_snapshot_additional_list_8011c913dc472a9e, []int{2}
}
func (m *SnapshotAdditionalFragment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SnapshotAdditionalFragment.Unmarshal(m, b)
}
func (m *SnapshotAdditionalFragment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SnapshotAdditionalFragment.Marshal(b, m, deterministic)
}
func (dst *SnapshotAdditionalFragment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SnapshotAdditionalFragment.Merge(dst, src)
}
func (m *SnapshotAdditionalFragment) XXX_Size() int {
	return xxx_messageInfo_SnapshotAdditionalFragment.Size(m)
}
func (m *SnapshotAdditionalFragment) XXX_DiscardUnknown() {
	xxx_messageInfo_SnapshotAdditionalFragment.DiscardUnknown(m)
}

var xxx_messageInfo_SnapshotAdditionalFragment proto.InternalMessageInfo

func (m *SnapshotAdditionalFragment) GetList() []*SnapshotAdditionalItem {
	if m != nil {
		return m.List
	}
	return nil
}

func init() {
	proto.RegisterType((*HashHeight)(nil), "vitepb.HashHeight")
	proto.RegisterType((*SnapshotAdditionalItem)(nil), "vitepb.SnapshotAdditionalItem")
	proto.RegisterType((*SnapshotAdditionalFragment)(nil), "vitepb.SnapshotAdditionalFragment")
}

func init() {
	proto.RegisterFile("vitepb/snapshot_additional_list.proto", fileDescriptor_snapshot_additional_list_8011c913dc472a9e)
}

var fileDescriptor_snapshot_additional_list_8011c913dc472a9e = []byte{
	// 230 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x41, 0x4b, 0x03, 0x31,
	0x10, 0x85, 0x89, 0x5d, 0xf7, 0x30, 0x15, 0x0f, 0x83, 0x94, 0xe0, 0x41, 0xc2, 0x82, 0x92, 0xd3,
	0x0a, 0xeb, 0xc5, 0xab, 0x1e, 0xa4, 0xde, 0x34, 0xfe, 0x80, 0x32, 0xa5, 0x21, 0x09, 0xb4, 0x9b,
	0x75, 0x33, 0xfa, 0x63, 0xfc, 0xb5, 0x62, 0xd2, 0x5a, 0xd0, 0xbd, 0x65, 0xf2, 0x1e, 0x6f, 0xbe,
	0x79, 0x70, 0xfd, 0x19, 0xd8, 0x0e, 0xeb, 0xdb, 0xd4, 0xd3, 0x90, 0x7c, 0xe4, 0x15, 0x6d, 0x36,
	0x81, 0x43, 0xec, 0x69, 0xbb, 0xda, 0x86, 0xc4, 0xed, 0x30, 0x46, 0x8e, 0x58, 0x17, 0x5b, 0x73,
	0x0f, 0xb0, 0xa4, 0xe4, 0x97, 0x36, 0x38, 0xcf, 0x88, 0x50, 0x79, 0x4a, 0x5e, 0x0a, 0x25, 0xf4,
	0x99, 0xc9, 0x6f, 0x5c, 0x40, 0xed, 0xb3, 0x2a, 0x4f, 0x94, 0xd0, 0x95, 0xd9, 0x4f, 0xcd, 0x97,
	0x80, 0xc5, 0xdb, 0x7e, 0xc9, 0xc3, 0xef, 0x8e, 0x67, 0xb6, 0x3b, 0xbc, 0x80, 0xd3, 0xf7, 0x8f,
	0xc8, 0x94, 0x73, 0x2a, 0x53, 0x06, 0xbc, 0x81, 0x73, 0x72, 0x6e, 0xb4, 0x8e, 0xd8, 0xbe, 0x66,
	0xb9, 0x04, 0xfe, 0xf9, 0xc5, 0x47, 0xc0, 0x03, 0xfc, 0x11, 0x4d, 0xce, 0x94, 0xd0, 0xf3, 0x0e,
	0xdb, 0xc2, 0xdd, 0x1e, 0x15, 0x33, 0xe1, 0x6e, 0x5e, 0xe0, 0xf2, 0x3f, 0xdb, 0xd3, 0x48, 0x6e,
	0x67, 0x7b, 0xc6, 0x0e, 0xaa, 0x9f, 0x2a, 0xa4, 0x50, 0x33, 0x3d, 0xef, 0xae, 0x0e, 0x99, 0xd3,
	0xd7, 0x98, 0xec, 0x5d, 0xd7, 0xb9, 0xb7, 0xbb, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0xc8, 0x11,
	0xc9, 0x79, 0x60, 0x01, 0x00, 0x00,
}