package crc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type CRC32Variant struct {
	Name      string
	Polynomial uint32
	Init      uint32
	RefIn     bool
	RefOut    bool
	XorOut    uint32
}

var (
	CRC32VariantISO_HDLC = CRC32Variant{
		Name:      "CRC-32/ISO-HDLC",
		Polynomial: 0x04C11DB7,
		Init:      0xFFFFFFFF,
		RefIn:     true,
		RefOut:    true,
		XorOut:    0xFFFFFFFF,
	}

	CRC32VariantBZIP2 = CRC32Variant{
		Name:      "CRC-32/BZIP2",
		Polynomial: 0x04C11DB7,
		Init:      0xFFFFFFFF,
		RefIn:     false,
		RefOut:    false,
		XorOut:    0xFFFFFFFF,
	}

	CRC32VariantMPEG2 = CRC32Variant{
		Name:      "CRC-32/MPEG-2",
		Polynomial: 0x04C11DB7,
		Init:      0xFFFFFFFF,
		RefIn:     false,
		RefOut:    false,
		XorOut:    0x00000000,
	}

	CRC32VariantPOSIX = CRC32Variant{
		Name:      "CRC-32/POSIX",
		Polynomial: 0x04C11DB7,
		Init:      0x00000000,
		RefIn:     false,
		RefOut:    false,
		XorOut:    0xFFFFFFFF,
	}
)

var CRC32Variants = map[string]CRC32Variant{
	"iso-hdlc": CRC32VariantISO_HDLC,
	"bzip2":    CRC32VariantBZIP2,
	"mpeg2":    CRC32VariantMPEG2,
	"posix":    CRC32VariantPOSIX,
}

type CRC32Table [256]uint32

func MakeTable32(poly uint32, refIn bool) (*CRC32Table, error) {
	if poly&0xFF000000 == 0 && poly != 0 {
		return nil, errors.New("polynomial must be 32-bit for CRC32")
	}

	table := &CRC32Table{}
	for i := 0; i < 256; i++ {
		var crc uint32
		if refIn {
			crc = reverseBits32(uint32(i))
		} else {
			crc = uint32(i) << 24
		}

		for j := 0; j < 8; j++ {
			if refIn {
				if (crc & 0x00000001) != 0 {
					crc = (crc >> 1) ^ poly
				} else {
					crc >>= 1
				}
			} else {
				if (crc & 0x80000000) != 0 {
					crc = (crc << 1) ^ poly
				} else {
					crc <<= 1
				}
			}
		}

		if refIn {
			crc = reverseBits32(crc)
		}

		table[i] = crc
	}

	return table, nil
}

func reverseBits32(x uint32) uint32 {
	x = ((x >> 1) & 0x55555555) | ((x & 0x55555555) << 1)
	x = ((x >> 2) & 0x33333333) | ((x & 0x33333333) << 2)
	x = ((x >> 4) & 0x0F0F0F0F) | ((x & 0x0F0F0F0F) << 4)
	x = ((x >> 8) & 0x00FF00FF) | ((x & 0x00FF00FF) << 8)
	x = (x >> 16) | (x << 16)
	return x
}

func CalculateCRC32(data []byte, variant CRC32Variant, table *CRC32Table) (uint32, error) {
	if table == nil {
		var err error
		table, err = MakeTable32(variant.Polynomial, variant.RefIn)
		if err != nil {
			return 0, err
		}
	}

	crc := variant.Init

	if variant.RefIn {
		for _, b := range data {
			index := byte(crc) ^ b
			crc = (crc >> 8) ^ table[index]
		}
	} else {
		for _, b := range data {
			index := byte(crc>>24) ^ b
			crc = (crc << 8) ^ table[index]
		}
	}

	if variant.RefOut {
		crc = reverseBits32(crc)
	}

	return crc ^ variant.XorOut, nil
}

func CalculateCRC32Stream(reader io.Reader, variant CRC32Variant, table *CRC32Table, bufferSize int) (uint32, error) {
	if table == nil {
		var err error
		table, err = MakeTable32(variant.Polynomial, variant.RefIn)
		if err != nil {
			return 0, err
		}
	}

	if bufferSize <= 0 {
		bufferSize = 4096
	}

	crc := variant.Init
	buffer := make([]byte, bufferSize)

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return 0, err
		}

		if n > 0 {
			if variant.RefIn {
				for i := 0; i < n; i++ {
					index := byte(crc) ^ buffer[i]
					crc = (crc >> 8) ^ table[index]
				}
			} else {
				for i := 0; i < n; i++ {
					index := byte(crc>>24) ^ buffer[i]
					crc = (crc << 8) ^ table[index]
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	if variant.RefOut {
		crc = reverseBits32(crc)
	}

	return crc ^ variant.XorOut, nil
}

func VerifyCRC32(data []byte, expected uint32, variant CRC32Variant) (bool, error) {
	crc, err := CalculateCRC32(data, variant, nil)
	if err != nil {
		return false, err
	}
	return crc == expected, nil
}

func VerifyCRC32Stream(reader io.Reader, expected uint32, variant CRC32Variant) (bool, error) {
	crc, err := CalculateCRC32Stream(reader, variant, nil, 4096)
	if err != nil {
		return false, err
	}
	return crc == expected, nil
}

func FormatCRC32(crc uint32) string {
	return fmt.Sprintf("0x%08X", crc)
}

func ParseCRC32(s string) (uint32, error) {
	var value uint32
	_, err := fmt.Sscanf(s, "0x%08X", &value)
	if err != nil {
		_, err = fmt.Sscanf(s, "0X%08X", &value)
		if err != nil {
			return 0, err
		}
	}
	return value, nil
}

func CRC32ToBytes(crc uint32, bigEndian bool) []byte {
	bytes := make([]byte, 4)
	if bigEndian {
		binary.BigEndian.PutUint32(bytes, crc)
	} else {
		binary.LittleEndian.PutUint32(bytes, crc)
	}
	return bytes
}

func BytesToCRC32(bytes []byte, bigEndian bool) (uint32, error) {
	if len(bytes) != 4 {
		return 0, errors.New("CRC32 value must be exactly 4 bytes")
	}
	if bigEndian {
		return binary.BigEndian.Uint32(bytes), nil
	}
	return binary.LittleEndian.Uint32(bytes), nil
}
