// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: juno/mint/v1/mint.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
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

// Minter represents the minting state.
type Minter struct {
	// current annual inflation rate
	Inflation       cosmossdk_io_math.LegacyDec `protobuf:"bytes,1,opt,name=inflation,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"inflation"`
	Phase           uint64                      `protobuf:"varint,2,opt,name=phase,proto3" json:"phase,omitempty"`
	StartPhaseBlock uint64                      `protobuf:"varint,3,opt,name=start_phase_block,json=startPhaseBlock,proto3" json:"start_phase_block,omitempty"`
	// current annual expected provisions
	AnnualProvisions cosmossdk_io_math.LegacyDec `protobuf:"bytes,4,opt,name=annual_provisions,json=annualProvisions,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"annual_provisions"`
	TargetSupply     cosmossdk_io_math.Int       `protobuf:"bytes,5,opt,name=target_supply,json=targetSupply,proto3,customtype=cosmossdk.io/math.Int" json:"target_supply"`
}

func (m *Minter) Reset()         { *m = Minter{} }
func (m *Minter) String() string { return proto.CompactTextString(m) }
func (*Minter) ProtoMessage()    {}
func (*Minter) Descriptor() ([]byte, []int) {
	return fileDescriptor_6f00ea321a7a5c34, []int{0}
}
func (m *Minter) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Minter) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Minter.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Minter) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Minter.Merge(m, src)
}
func (m *Minter) XXX_Size() int {
	return m.Size()
}
func (m *Minter) XXX_DiscardUnknown() {
	xxx_messageInfo_Minter.DiscardUnknown(m)
}

var xxx_messageInfo_Minter proto.InternalMessageInfo

func (m *Minter) GetPhase() uint64 {
	if m != nil {
		return m.Phase
	}
	return 0
}

func (m *Minter) GetStartPhaseBlock() uint64 {
	if m != nil {
		return m.StartPhaseBlock
	}
	return 0
}

// Params holds parameters for the mint module.
type Params struct {
	// type of coin to mint
	MintDenom string `protobuf:"bytes,1,opt,name=mint_denom,json=mintDenom,proto3" json:"mint_denom,omitempty"`
	// expected blocks per year
	BlocksPerYear uint64 `protobuf:"varint,2,opt,name=blocks_per_year,json=blocksPerYear,proto3" json:"blocks_per_year,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_6f00ea321a7a5c34, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetMintDenom() string {
	if m != nil {
		return m.MintDenom
	}
	return ""
}

func (m *Params) GetBlocksPerYear() uint64 {
	if m != nil {
		return m.BlocksPerYear
	}
	return 0
}

func init() {
	proto.RegisterType((*Minter)(nil), "juno.mint.v1.Minter")
	proto.RegisterType((*Params)(nil), "juno.mint.v1.Params")
}

func init() { proto.RegisterFile("juno/mint/v1/mint.proto", fileDescriptor_6f00ea321a7a5c34) }

var fileDescriptor_6f00ea321a7a5c34 = []byte{
	// 423 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x92, 0x41, 0xab, 0xd3, 0x40,
	0x10, 0x80, 0x93, 0xe7, 0x7b, 0x85, 0x2e, 0xef, 0x51, 0xbb, 0x54, 0x8c, 0x95, 0xa6, 0xa5, 0x07,
	0x29, 0xd5, 0x66, 0x29, 0xde, 0x3c, 0xb6, 0x45, 0x28, 0x28, 0x86, 0x78, 0xd2, 0x83, 0x61, 0x9b,
	0xae, 0xe9, 0xda, 0x64, 0x37, 0xec, 0x6e, 0x8a, 0xf9, 0x0b, 0x9e, 0xfc, 0x29, 0x1e, 0xfc, 0x11,
	0xbd, 0x08, 0xc5, 0x93, 0x78, 0x28, 0xd2, 0x1e, 0xfc, 0x1b, 0xb2, 0xbb, 0x41, 0x05, 0x6f, 0x5e,
	0x92, 0xcc, 0x37, 0x93, 0x6f, 0xb2, 0x93, 0x01, 0x77, 0xdf, 0x95, 0x8c, 0xa3, 0x9c, 0x32, 0x85,
	0x76, 0x53, 0x73, 0x0f, 0x0a, 0xc1, 0x15, 0x87, 0xd7, 0x3a, 0x11, 0x18, 0xb0, 0x9b, 0x76, 0xef,
	0x25, 0x5c, 0xe6, 0x5c, 0xc6, 0x26, 0x87, 0x6c, 0x60, 0x0b, 0xbb, 0x6d, 0x9c, 0x53, 0xc6, 0x91,
	0xb9, 0xd6, 0xa8, 0x93, 0xf2, 0x94, 0xdb, 0x52, 0xfd, 0x64, 0xe9, 0xf0, 0xcb, 0x05, 0x68, 0x3c,
	0xa7, 0x4c, 0x11, 0x01, 0x5f, 0x80, 0x26, 0x65, 0x6f, 0x33, 0xac, 0x28, 0x67, 0x9e, 0x3b, 0x70,
	0x47, 0xcd, 0xd9, 0x74, 0x7f, 0xec, 0x3b, 0xdf, 0x8f, 0xfd, 0xfb, 0x56, 0x2e, 0xd7, 0xdb, 0x80,
	0x72, 0x94, 0x63, 0xb5, 0x09, 0x9e, 0x91, 0x14, 0x27, 0xd5, 0x82, 0x24, 0x5f, 0x3f, 0x4f, 0x40,
	0xdd, 0x7b, 0x41, 0x92, 0xe8, 0x8f, 0x03, 0x76, 0xc0, 0x55, 0xb1, 0xc1, 0x92, 0x78, 0x17, 0x03,
	0x77, 0x74, 0x19, 0xd9, 0x00, 0x8e, 0x41, 0x5b, 0x2a, 0x2c, 0x54, 0x6c, 0xc2, 0x78, 0x95, 0xf1,
	0x64, 0xeb, 0xdd, 0x32, 0x15, 0x2d, 0x93, 0x08, 0x35, 0x9f, 0x69, 0x0c, 0xdf, 0x80, 0x36, 0x66,
	0xac, 0xc4, 0x99, 0x3e, 0xe3, 0x8e, 0x4a, 0xca, 0x99, 0xf4, 0x2e, 0xff, 0xf7, 0xd3, 0x6e, 0x5b,
	0x57, 0xf8, 0x5b, 0x05, 0x43, 0x70, 0xa3, 0xb0, 0x48, 0x89, 0x8a, 0x65, 0x59, 0x14, 0x59, 0xe5,
	0x5d, 0x19, 0xf7, 0xc3, 0xda, 0x7d, 0xe7, 0x5f, 0xf7, 0x92, 0xa9, 0xbf, 0xac, 0x4b, 0xa6, 0xa2,
	0x6b, 0x6b, 0x78, 0x69, 0x04, 0x43, 0x06, 0x1a, 0x21, 0x16, 0x38, 0x97, 0xb0, 0x07, 0x80, 0xfe,
	0x51, 0xf1, 0x9a, 0x30, 0x9e, 0xdb, 0x79, 0x46, 0x4d, 0x4d, 0x16, 0x1a, 0xc0, 0x07, 0xa0, 0x65,
	0x8e, 0x2e, 0xe3, 0x82, 0x88, 0xb8, 0x22, 0x58, 0xd4, 0x63, 0xba, 0xb1, 0x38, 0x24, 0xe2, 0x15,
	0xc1, 0xe2, 0x49, 0xef, 0xc3, 0xcf, 0x4f, 0x63, 0xcf, 0xf6, 0x9b, 0xc8, 0xf5, 0x16, 0xbd, 0xb7,
	0x8b, 0x61, 0xbb, 0xcc, 0x9e, 0xee, 0x4f, 0xbe, 0x7b, 0x38, 0xf9, 0xee, 0x8f, 0x93, 0xef, 0x7e,
	0x3c, 0xfb, 0xce, 0xe1, 0xec, 0x3b, 0xdf, 0xce, 0xbe, 0xf3, 0xfa, 0x51, 0x4a, 0xd5, 0xa6, 0x5c,
	0x05, 0x09, 0xcf, 0xd1, 0xdc, 0xbc, 0x3e, 0xe7, 0x4c, 0x09, 0x9c, 0x28, 0x89, 0xcc, 0x7e, 0xd5,
	0x22, 0x55, 0x15, 0x44, 0xae, 0x1a, 0x66, 0x1d, 0x1e, 0xff, 0x0a, 0x00, 0x00, 0xff, 0xff, 0xcb,
	0xb0, 0xfa, 0x72, 0x7b, 0x02, 0x00, 0x00,
}

func (m *Minter) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Minter) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Minter) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.TargetSupply.Size()
		i -= size
		if _, err := m.TargetSupply.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	{
		size := m.AnnualProvisions.Size()
		i -= size
		if _, err := m.AnnualProvisions.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.StartPhaseBlock != 0 {
		i = encodeVarintMint(dAtA, i, uint64(m.StartPhaseBlock))
		i--
		dAtA[i] = 0x18
	}
	if m.Phase != 0 {
		i = encodeVarintMint(dAtA, i, uint64(m.Phase))
		i--
		dAtA[i] = 0x10
	}
	{
		size := m.Inflation.Size()
		i -= size
		if _, err := m.Inflation.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintMint(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.BlocksPerYear != 0 {
		i = encodeVarintMint(dAtA, i, uint64(m.BlocksPerYear))
		i--
		dAtA[i] = 0x10
	}
	if len(m.MintDenom) > 0 {
		i -= len(m.MintDenom)
		copy(dAtA[i:], m.MintDenom)
		i = encodeVarintMint(dAtA, i, uint64(len(m.MintDenom)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMint(dAtA []byte, offset int, v uint64) int {
	offset -= sovMint(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Minter) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Inflation.Size()
	n += 1 + l + sovMint(uint64(l))
	if m.Phase != 0 {
		n += 1 + sovMint(uint64(m.Phase))
	}
	if m.StartPhaseBlock != 0 {
		n += 1 + sovMint(uint64(m.StartPhaseBlock))
	}
	l = m.AnnualProvisions.Size()
	n += 1 + l + sovMint(uint64(l))
	l = m.TargetSupply.Size()
	n += 1 + l + sovMint(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.MintDenom)
	if l > 0 {
		n += 1 + l + sovMint(uint64(l))
	}
	if m.BlocksPerYear != 0 {
		n += 1 + sovMint(uint64(m.BlocksPerYear))
	}
	return n
}

func sovMint(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMint(x uint64) (n int) {
	return sovMint(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Minter) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: Minter: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Minter: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Inflation", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Inflation.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Phase", wireType)
			}
			m.Phase = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Phase |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field StartPhaseBlock", wireType)
			}
			m.StartPhaseBlock = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.StartPhaseBlock |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AnnualProvisions", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.AnnualProvisions.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TargetSupply", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TargetSupply.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMint
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MintDenom", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
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
				return ErrInvalidLengthMint
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMint
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MintDenom = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field BlocksPerYear", wireType)
			}
			m.BlocksPerYear = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMint
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.BlocksPerYear |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMint(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMint
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
func skipMint(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
					return 0, ErrIntOverflowMint
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
				return 0, ErrInvalidLengthMint
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMint
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMint
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMint        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMint          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMint = fmt.Errorf("proto: unexpected end of group")
)
