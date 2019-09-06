// Code generated by protoc-gen-go. DO NOT EDIT.
// source: opr.proto

package oprencoding

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

type ProtoOPR struct {
	Address              string   `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	Id                   string   `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Height               int32    `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Winners              []string `protobuf:"bytes,4,rep,name=winners,proto3" json:"winners,omitempty"`
	PEG                  uint64   `protobuf:"varint,5,opt,name=PEG,proto3" json:"PEG,omitempty"`
	PUSD                 uint64   `protobuf:"varint,6,opt,name=pUSD,proto3" json:"pUSD,omitempty"`
	PEUR                 uint64   `protobuf:"varint,7,opt,name=pEUR,proto3" json:"pEUR,omitempty"`
	PJPY                 uint64   `protobuf:"varint,8,opt,name=pJPY,proto3" json:"pJPY,omitempty"`
	PGBP                 uint64   `protobuf:"varint,9,opt,name=pGBP,proto3" json:"pGBP,omitempty"`
	PCAD                 uint64   `protobuf:"varint,10,opt,name=pCAD,proto3" json:"pCAD,omitempty"`
	PCHF                 uint64   `protobuf:"varint,11,opt,name=pCHF,proto3" json:"pCHF,omitempty"`
	PINR                 uint64   `protobuf:"varint,12,opt,name=pINR,proto3" json:"pINR,omitempty"`
	PSGD                 uint64   `protobuf:"varint,13,opt,name=pSGD,proto3" json:"pSGD,omitempty"`
	PCNY                 uint64   `protobuf:"varint,14,opt,name=pCNY,proto3" json:"pCNY,omitempty"`
	PHKD                 uint64   `protobuf:"varint,15,opt,name=pHKD,proto3" json:"pHKD,omitempty"`
	PKRW                 uint64   `protobuf:"varint,16,opt,name=pKRW,proto3" json:"pKRW,omitempty"`
	PBRL                 uint64   `protobuf:"varint,17,opt,name=pBRL,proto3" json:"pBRL,omitempty"`
	PPHP                 uint64   `protobuf:"varint,18,opt,name=pPHP,proto3" json:"pPHP,omitempty"`
	PMXN                 uint64   `protobuf:"varint,19,opt,name=pMXN,proto3" json:"pMXN,omitempty"`
	PXAU                 uint64   `protobuf:"varint,20,opt,name=pXAU,proto3" json:"pXAU,omitempty"`
	PXAG                 uint64   `protobuf:"varint,21,opt,name=pXAG,proto3" json:"pXAG,omitempty"`
	PXBT                 uint64   `protobuf:"varint,22,opt,name=pXBT,proto3" json:"pXBT,omitempty"`
	PETH                 uint64   `protobuf:"varint,23,opt,name=pETH,proto3" json:"pETH,omitempty"`
	PLTC                 uint64   `protobuf:"varint,24,opt,name=pLTC,proto3" json:"pLTC,omitempty"`
	PRVN                 uint64   `protobuf:"varint,25,opt,name=pRVN,proto3" json:"pRVN,omitempty"`
	PXBC                 uint64   `protobuf:"varint,26,opt,name=pXBC,proto3" json:"pXBC,omitempty"`
	PFCT                 uint64   `protobuf:"varint,27,opt,name=pFCT,proto3" json:"pFCT,omitempty"`
	PBNB                 uint64   `protobuf:"varint,28,opt,name=pBNB,proto3" json:"pBNB,omitempty"`
	PXLM                 uint64   `protobuf:"varint,29,opt,name=pXLM,proto3" json:"pXLM,omitempty"`
	PADA                 uint64   `protobuf:"varint,30,opt,name=pADA,proto3" json:"pADA,omitempty"`
	PXMR                 uint64   `protobuf:"varint,31,opt,name=pXMR,proto3" json:"pXMR,omitempty"`
	PDASH                uint64   `protobuf:"varint,32,opt,name=pDASH,proto3" json:"pDASH,omitempty"`
	PZEC                 uint64   `protobuf:"varint,33,opt,name=pZEC,proto3" json:"pZEC,omitempty"`
	PDCR                 uint64   `protobuf:"varint,34,opt,name=pDCR,proto3" json:"pDCR,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProtoOPR) Reset()         { *m = ProtoOPR{} }
func (m *ProtoOPR) String() string { return proto.CompactTextString(m) }
func (*ProtoOPR) ProtoMessage()    {}
func (*ProtoOPR) Descriptor() ([]byte, []int) {
	return fileDescriptor_9c939583bc8284b5, []int{0}
}

func (m *ProtoOPR) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProtoOPR.Unmarshal(m, b)
}
func (m *ProtoOPR) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProtoOPR.Marshal(b, m, deterministic)
}
func (m *ProtoOPR) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProtoOPR.Merge(m, src)
}
func (m *ProtoOPR) XXX_Size() int {
	return xxx_messageInfo_ProtoOPR.Size(m)
}
func (m *ProtoOPR) XXX_DiscardUnknown() {
	xxx_messageInfo_ProtoOPR.DiscardUnknown(m)
}

var xxx_messageInfo_ProtoOPR proto.InternalMessageInfo

func (m *ProtoOPR) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *ProtoOPR) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *ProtoOPR) GetHeight() int32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *ProtoOPR) GetWinners() []string {
	if m != nil {
		return m.Winners
	}
	return nil
}

func (m *ProtoOPR) GetPEG() uint64 {
	if m != nil {
		return m.PEG
	}
	return 0
}

func (m *ProtoOPR) GetPUSD() uint64 {
	if m != nil {
		return m.PUSD
	}
	return 0
}

func (m *ProtoOPR) GetPEUR() uint64 {
	if m != nil {
		return m.PEUR
	}
	return 0
}

func (m *ProtoOPR) GetPJPY() uint64 {
	if m != nil {
		return m.PJPY
	}
	return 0
}

func (m *ProtoOPR) GetPGBP() uint64 {
	if m != nil {
		return m.PGBP
	}
	return 0
}

func (m *ProtoOPR) GetPCAD() uint64 {
	if m != nil {
		return m.PCAD
	}
	return 0
}

func (m *ProtoOPR) GetPCHF() uint64 {
	if m != nil {
		return m.PCHF
	}
	return 0
}

func (m *ProtoOPR) GetPINR() uint64 {
	if m != nil {
		return m.PINR
	}
	return 0
}

func (m *ProtoOPR) GetPSGD() uint64 {
	if m != nil {
		return m.PSGD
	}
	return 0
}

func (m *ProtoOPR) GetPCNY() uint64 {
	if m != nil {
		return m.PCNY
	}
	return 0
}

func (m *ProtoOPR) GetPHKD() uint64 {
	if m != nil {
		return m.PHKD
	}
	return 0
}

func (m *ProtoOPR) GetPKRW() uint64 {
	if m != nil {
		return m.PKRW
	}
	return 0
}

func (m *ProtoOPR) GetPBRL() uint64 {
	if m != nil {
		return m.PBRL
	}
	return 0
}

func (m *ProtoOPR) GetPPHP() uint64 {
	if m != nil {
		return m.PPHP
	}
	return 0
}

func (m *ProtoOPR) GetPMXN() uint64 {
	if m != nil {
		return m.PMXN
	}
	return 0
}

func (m *ProtoOPR) GetPXAU() uint64 {
	if m != nil {
		return m.PXAU
	}
	return 0
}

func (m *ProtoOPR) GetPXAG() uint64 {
	if m != nil {
		return m.PXAG
	}
	return 0
}

func (m *ProtoOPR) GetPXBT() uint64 {
	if m != nil {
		return m.PXBT
	}
	return 0
}

func (m *ProtoOPR) GetPETH() uint64 {
	if m != nil {
		return m.PETH
	}
	return 0
}

func (m *ProtoOPR) GetPLTC() uint64 {
	if m != nil {
		return m.PLTC
	}
	return 0
}

func (m *ProtoOPR) GetPRVN() uint64 {
	if m != nil {
		return m.PRVN
	}
	return 0
}

func (m *ProtoOPR) GetPXBC() uint64 {
	if m != nil {
		return m.PXBC
	}
	return 0
}

func (m *ProtoOPR) GetPFCT() uint64 {
	if m != nil {
		return m.PFCT
	}
	return 0
}

func (m *ProtoOPR) GetPBNB() uint64 {
	if m != nil {
		return m.PBNB
	}
	return 0
}

func (m *ProtoOPR) GetPXLM() uint64 {
	if m != nil {
		return m.PXLM
	}
	return 0
}

func (m *ProtoOPR) GetPADA() uint64 {
	if m != nil {
		return m.PADA
	}
	return 0
}

func (m *ProtoOPR) GetPXMR() uint64 {
	if m != nil {
		return m.PXMR
	}
	return 0
}

func (m *ProtoOPR) GetPDASH() uint64 {
	if m != nil {
		return m.PDASH
	}
	return 0
}

func (m *ProtoOPR) GetPZEC() uint64 {
	if m != nil {
		return m.PZEC
	}
	return 0
}

func (m *ProtoOPR) GetPDCR() uint64 {
	if m != nil {
		return m.PDCR
	}
	return 0
}

func init() {
	proto.RegisterType((*ProtoOPR)(nil), "oprencoding.ProtoOPR")
}

func init() { proto.RegisterFile("opr.proto", fileDescriptor_9c939583bc8284b5) }

var fileDescriptor_9c939583bc8284b5 = []byte{
	// 409 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0xd2, 0xcb, 0x72, 0xd3, 0x30,
	0x14, 0x06, 0xe0, 0x49, 0x73, 0x69, 0xe3, 0x42, 0x29, 0xa2, 0x94, 0x9f, 0xbb, 0xe9, 0x2a, 0x2b,
	0x36, 0x3c, 0x81, 0x2e, 0x8e, 0xd5, 0xc6, 0x16, 0x1a, 0xc5, 0x6e, 0x9d, 0xee, 0x00, 0x67, 0x5a,
	0x6f, 0x62, 0x8f, 0xd3, 0x19, 0xde, 0x8e, 0x67, 0x63, 0x22, 0x1d, 0x77, 0xa7, 0xf3, 0xcd, 0x7f,
	0xce, 0x91, 0x35, 0x8e, 0xe6, 0x6d, 0xd7, 0x7f, 0xef, 0xfa, 0xf6, 0xa9, 0x65, 0xa7, 0x6d, 0xd7,
	0x6f, 0x77, 0x7f, 0xda, 0xba, 0xd9, 0x3d, 0x5c, 0xfd, 0x9b, 0x46, 0x27, 0xf6, 0xc0, 0x3f, 0xad,
	0x63, 0x88, 0x8e, 0x7f, 0xd5, 0x75, 0xbf, 0xdd, 0xef, 0x31, 0x8a, 0x47, 0x8b, 0xb9, 0x1b, 0x4a,
	0x76, 0x16, 0x1d, 0x35, 0x35, 0x8e, 0x3c, 0x1e, 0x35, 0x35, 0xbb, 0x8c, 0x66, 0x8f, 0xdb, 0xe6,
	0xe1, 0xf1, 0x09, 0xe3, 0x78, 0xb4, 0x98, 0x3a, 0xaa, 0x0e, 0x13, 0xfe, 0x36, 0xbb, 0xdd, 0xb6,
	0xdf, 0x63, 0x12, 0x8f, 0x0f, 0x13, 0xa8, 0x64, 0xe7, 0xd1, 0xd8, 0x26, 0x29, 0xa6, 0xf1, 0x68,
	0x31, 0x71, 0x87, 0x23, 0x63, 0xd1, 0xa4, 0x2b, 0xd7, 0x0a, 0x33, 0x4f, 0xfe, 0xec, 0x2d, 0x29,
	0x1d, 0x8e, 0xc9, 0x92, 0xd2, 0x79, 0xbb, 0xb1, 0x1b, 0x9c, 0x90, 0xdd, 0xd8, 0x8d, 0xb7, 0x54,
	0x58, 0xcc, 0xc9, 0x52, 0x61, 0xbd, 0x49, 0xae, 0x10, 0x91, 0x49, 0x1e, 0xe6, 0x49, 0xbd, 0xc4,
	0xe9, 0x60, 0x7a, 0xe9, 0xed, 0xda, 0x38, 0xbc, 0x20, 0xbb, 0x36, 0x61, 0xc7, 0x3a, 0x55, 0x78,
	0x49, 0xb6, 0x4e, 0xa9, 0xd7, 0x6c, 0x70, 0x36, 0xf4, 0x9a, 0xb0, 0x57, 0xaf, 0x14, 0x5e, 0x91,
	0xe9, 0x55, 0xc8, 0xad, 0xdc, 0x1d, 0xce, 0xc9, 0x56, 0xee, 0xce, 0x9b, 0x70, 0x19, 0x5e, 0x93,
	0x09, 0x97, 0x79, 0xb3, 0xda, 0x82, 0x91, 0x59, 0x1d, 0xee, 0x9c, 0x57, 0x06, 0x6f, 0xc8, 0xf2,
	0xca, 0x78, 0xab, 0x78, 0x89, 0x0b, 0xb2, 0x8a, 0x97, 0x64, 0x29, 0xde, 0x3e, 0x5b, 0x78, 0xbf,
	0x4a, 0x14, 0xb8, 0x1c, 0x4c, 0x14, 0xe1, 0xfd, 0x0a, 0x8d, 0x77, 0xc3, 0xfb, 0x15, 0xda, 0x5b,
	0x56, 0x48, 0x80, 0x2c, 0x2b, 0xa4, 0x37, 0x77, 0x6b, 0xf0, 0x9e, 0xcc, 0xdd, 0xd2, 0x5e, 0x21,
	0xf1, 0xe1, 0x79, 0x5e, 0xc8, 0x2d, 0x65, 0x81, 0x8f, 0x64, 0x4b, 0x19, 0x76, 0x08, 0x23, 0xf0,
	0x69, 0xf8, 0x36, 0x23, 0x42, 0x6f, 0x96, 0xe3, 0xf3, 0xd0, 0x9b, 0xe5, 0xde, 0xb8, 0xe2, 0xf8,
	0x42, 0xc6, 0x15, 0x0f, 0xb9, 0xdc, 0xe1, 0xeb, 0x90, 0xcb, 0x1d, 0xbb, 0x88, 0xa6, 0x9d, 0xe2,
	0x6b, 0x8d, 0xd8, 0x63, 0x28, 0x7c, 0xf2, 0x3e, 0x91, 0xf8, 0x46, 0xc9, 0xfb, 0x24, 0xdc, 0x46,
	0x49, 0x87, 0x2b, 0x32, 0x25, 0xdd, 0xef, 0x99, 0xff, 0xa9, 0x7f, 0xfc, 0x0f, 0x00, 0x00, 0xff,
	0xff, 0xe9, 0xd7, 0xe4, 0xad, 0xe1, 0x02, 0x00, 0x00,
}