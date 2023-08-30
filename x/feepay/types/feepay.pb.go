// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: juno/feepay/v1/feepay.proto

package types

import (
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
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

// FeeShare defines an instance that organizes fee distribution conditions for
// the owner of a given smart contract
type FeePayContract struct {
	ContractAddress string `protobuf:"bytes,1,opt,name=contract_address,json=contractAddress,proto3" json:"contract_address,omitempty"`
}

func (m *FeePayContract) Reset()         { *m = FeePayContract{} }
func (m *FeePayContract) String() string { return proto.CompactTextString(m) }
func (*FeePayContract) ProtoMessage()    {}
func (*FeePayContract) Descriptor() ([]byte, []int) {
	return fileDescriptor_14ea6771eacbfed1, []int{0}
}
func (m *FeePayContract) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FeePayContract) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FeePayContract.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FeePayContract) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeePayContract.Merge(m, src)
}
func (m *FeePayContract) XXX_Size() int {
	return m.Size()
}
func (m *FeePayContract) XXX_DiscardUnknown() {
	xxx_messageInfo_FeePayContract.DiscardUnknown(m)
}

var xxx_messageInfo_FeePayContract proto.InternalMessageInfo

func (m *FeePayContract) GetContractAddress() string {
	if m != nil {
		return m.ContractAddress
	}
	return ""
}

type FeePayUserUsage struct {
	// not used as a Tx, just a storage device
	UserAddress string `protobuf:"bytes,1,opt,name=user_address,json=userAddress,proto3" json:"user_address,omitempty"`
	NumUses     uint64 `protobuf:"varint,2,opt,name=num_uses,json=numUses,proto3" json:"num_uses,omitempty"`
}

func (m *FeePayUserUsage) Reset()         { *m = FeePayUserUsage{} }
func (m *FeePayUserUsage) String() string { return proto.CompactTextString(m) }
func (*FeePayUserUsage) ProtoMessage()    {}
func (*FeePayUserUsage) Descriptor() ([]byte, []int) {
	return fileDescriptor_14ea6771eacbfed1, []int{1}
}
func (m *FeePayUserUsage) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *FeePayUserUsage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_FeePayUserUsage.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *FeePayUserUsage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FeePayUserUsage.Merge(m, src)
}
func (m *FeePayUserUsage) XXX_Size() int {
	return m.Size()
}
func (m *FeePayUserUsage) XXX_DiscardUnknown() {
	xxx_messageInfo_FeePayUserUsage.DiscardUnknown(m)
}

var xxx_messageInfo_FeePayUserUsage proto.InternalMessageInfo

func (m *FeePayUserUsage) GetUserAddress() string {
	if m != nil {
		return m.UserAddress
	}
	return ""
}

func (m *FeePayUserUsage) GetNumUses() uint64 {
	if m != nil {
		return m.NumUses
	}
	return 0
}

func init() {
	proto.RegisterType((*FeePayContract)(nil), "juno.feepay.v1.FeePayContract")
	proto.RegisterType((*FeePayUserUsage)(nil), "juno.feepay.v1.FeePayUserUsage")
}

func init() { proto.RegisterFile("juno/feepay/v1/feepay.proto", fileDescriptor_14ea6771eacbfed1) }

var fileDescriptor_14ea6771eacbfed1 = []byte{
	// 226 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xce, 0x2a, 0xcd, 0xcb,
	0xd7, 0x4f, 0x4b, 0x4d, 0x2d, 0x48, 0xac, 0xd4, 0x2f, 0x33, 0x84, 0xb2, 0xf4, 0x0a, 0x8a, 0xf2,
	0x4b, 0xf2, 0x85, 0xf8, 0x40, 0x92, 0x7a, 0x50, 0xa1, 0x32, 0x43, 0x25, 0x6b, 0x2e, 0x3e, 0xb7,
	0xd4, 0xd4, 0x80, 0xc4, 0x4a, 0xe7, 0xfc, 0xbc, 0x92, 0xa2, 0xc4, 0xe4, 0x12, 0x21, 0x4d, 0x2e,
	0x81, 0x64, 0x28, 0x3b, 0x3e, 0x31, 0x25, 0xa5, 0x28, 0xb5, 0xb8, 0x58, 0x82, 0x51, 0x81, 0x51,
	0x83, 0x33, 0x88, 0x1f, 0x26, 0xee, 0x08, 0x11, 0x56, 0xf2, 0xe7, 0xe2, 0x87, 0x68, 0x0e, 0x2d,
	0x4e, 0x2d, 0x0a, 0x2d, 0x4e, 0x4c, 0x4f, 0x15, 0x52, 0xe4, 0xe2, 0x29, 0x2d, 0x4e, 0x2d, 0x42,
	0xd3, 0xc9, 0x0d, 0x12, 0x83, 0xea, 0x12, 0x92, 0xe4, 0xe2, 0xc8, 0x2b, 0xcd, 0x8d, 0x2f, 0x2d,
	0x4e, 0x2d, 0x96, 0x60, 0x52, 0x60, 0xd4, 0x60, 0x09, 0x62, 0xcf, 0x2b, 0xcd, 0x0d, 0x2d, 0x4e,
	0x2d, 0x76, 0xf2, 0x38, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f, 0xe4, 0x18,
	0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0, 0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28, 0xbd, 0xf4,
	0xcc, 0x92, 0x8c, 0xd2, 0x24, 0xbd, 0xe4, 0xfc, 0x5c, 0x7d, 0xe7, 0xfc, 0xe2, 0xdc, 0xfc, 0x62,
	0x98, 0x83, 0x8b, 0xf5, 0xc1, 0xfe, 0xad, 0x80, 0xf9, 0xb8, 0xa4, 0xb2, 0x20, 0xb5, 0x38, 0x89,
	0x0d, 0xec, 0x5d, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff, 0xff, 0x68, 0xa5, 0x09, 0x5f, 0x0d, 0x01,
	0x00, 0x00,
}

func (m *FeePayContract) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FeePayContract) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FeePayContract) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ContractAddress) > 0 {
		i -= len(m.ContractAddress)
		copy(dAtA[i:], m.ContractAddress)
		i = encodeVarintFeepay(dAtA, i, uint64(len(m.ContractAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *FeePayUserUsage) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *FeePayUserUsage) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *FeePayUserUsage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.NumUses != 0 {
		i = encodeVarintFeepay(dAtA, i, uint64(m.NumUses))
		i--
		dAtA[i] = 0x10
	}
	if len(m.UserAddress) > 0 {
		i -= len(m.UserAddress)
		copy(dAtA[i:], m.UserAddress)
		i = encodeVarintFeepay(dAtA, i, uint64(len(m.UserAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintFeepay(dAtA []byte, offset int, v uint64) int {
	offset -= sovFeepay(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *FeePayContract) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ContractAddress)
	if l > 0 {
		n += 1 + l + sovFeepay(uint64(l))
	}
	return n
}

func (m *FeePayUserUsage) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.UserAddress)
	if l > 0 {
		n += 1 + l + sovFeepay(uint64(l))
	}
	if m.NumUses != 0 {
		n += 1 + sovFeepay(uint64(m.NumUses))
	}
	return n
}

func sovFeepay(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozFeepay(x uint64) (n int) {
	return sovFeepay(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *FeePayContract) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFeepay
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
			return fmt.Errorf("proto: FeePayContract: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FeePayContract: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFeepay
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
				return ErrInvalidLengthFeepay
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFeepay
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipFeepay(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthFeepay
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
func (m *FeePayUserUsage) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowFeepay
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
			return fmt.Errorf("proto: FeePayUserUsage: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: FeePayUserUsage: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field UserAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFeepay
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
				return ErrInvalidLengthFeepay
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthFeepay
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.UserAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field NumUses", wireType)
			}
			m.NumUses = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowFeepay
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.NumUses |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipFeepay(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthFeepay
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
func skipFeepay(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowFeepay
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
					return 0, ErrIntOverflowFeepay
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
					return 0, ErrIntOverflowFeepay
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
				return 0, ErrInvalidLengthFeepay
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupFeepay
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthFeepay
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthFeepay        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowFeepay          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupFeepay = fmt.Errorf("proto: unexpected end of group")
)
