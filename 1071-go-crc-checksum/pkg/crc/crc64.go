package crc

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type CRC64Variant struct {
	Name       string
	Polynomial uint64
	Init       uint64
	RefIn      bool
	RefOut     bool
	XorOut     uint64
}

var (
	CRC64VariantECMA = CRC64Variant{
		Name:       "CRC-64/ECMA-182",
		Polynomial: 0x42F0E1EBA9EA3693,
		Init:       0xFFFFFFFFFFFFFFFF,
		RefIn:      false,
		RefOut:     false,
		XorOut:     0xFFFFFFFFFFFFFFFF,
	}

	CRC64VariantWE = CRC64Variant{
		Name:       "CRC-64/WE",
		Polynomial: 0x42F0E1EBA9EA3693,
		Init:       0xFFFFFFFFFFFFFFFF,
		RefIn:      true,
		RefOut:     true,
		XorOut:     0xFFFFFFFFFFFFFFFF,
	}

	CRC64VariantISO = CRC64Variant{
		Name:       "CRC-64/ISO",
		Polynomial: 0x000000000000001B,
		Init:       0xFFFFFFFFFFFFFFFF,
		RefIn:      true,
		RefOut:     true,
		XorOut:     0xFFFFFFFFFFFFFFFF,
	}

	CRC64VariantGO = CRC64Variant{
		Name:       "CRC-64/GO",
		Polynomial: 0x000000000000001B,
		Init:       0xFFFFFFFFFFFFFFFF,
		RefIn:      true,
		RefOut:     true,
		XorOut:     0xFFFFFFFFFFFFFFFF,
	}
)

var CRC64Variants = map[string]CRC64Variant{
	"ecma":  CRC64VariantECMA,
	"we":    CRC64VariantWE,
	"iso":   CRC64VariantISO,
	"go":    CRC64VariantGO,
}

type CRC64Table [256]uint64

func MakeTable64(poly uint64, refIn bool) (*CRC64Table, error) {
	if poly&0xFF00000000000000 == 0 && poly != 0 {
		return nil, errors.New("polynomial must be 64-bit for CRC64")
	}

	table := &CRC64Table{}

	if refIn {
		reflectedPoly := reverseBits64(poly)
		for i := 0; i < 256; i++ {
			crc := uint64(i)
			for j := 0; j < 8; j++ {
				if (crc & 0x0000000000000001) != 0 {
					crc = (crc >> 1) ^ reflectedPoly
				} else {
					crc >>= 1
				}
			}
			table[i] = crc
		}
	} else {
		for i := 0; i < 256; i++ {
			crc := uint64(i) << 56
			for j := 0; j < 8; j++ {
				if (crc & 0x8000000000000000) != 0 {
					crc = (crc << 1) ^ poly
				} else {
					crc <<= 1
				}
			}
			table[i] = crc
		}
	}

	return table, nil
}

func reverseBits64(x uint64) uint64 {
	x = ((x >> 1) & 0x5555555555555555) | ((x & 0x5555555555555555) << 1)
	x = ((x >> 2) & 0x3333333333333333) | ((x & 0x3333333333333333) << 2)
	x = ((x >> 4) & 0x0F0F0F0F0F0F0F0F) | ((x & 0x0F0F0F0F0F0F0F0F) << 4)
	x = ((x >> 8) & 0x00FF00FF00FF00FF) | ((x & 0x00FF00FF00FF00FF) << 8)
	x = ((x >> 16) & 0x0000FFFF0000FFFF) | ((x & 0x0000FFFF0000FFFF) << 16)
	x = (x >> 32) | (x << 32)
	return x
}

func CalculateCRC64(data []byte, variant CRC64Variant, table *CRC64Table) (uint64, error) {
	if table == nil {
		var err error
		table, err = MakeTable64(variant.Polynomial, variant.RefIn)
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
		if !variant.RefOut {
			crc = reverseBits64(crc)
		}
	} else {
		for _, b := range data {
			index := byte(crc>>56) ^ b
			crc = (crc << 8) ^ table[index]
		}
		if variant.RefOut {
			crc = reverseBits64(crc)
		}
	}

	return crc ^ variant.XorOut, nil
}

func CalculateCRC64Stream(reader io.Reader, variant CRC64Variant, table *CRC64Table, bufferSize int) (uint64, error) {
	if table == nil {
		var err error
		table, err = MakeTable64(variant.Polynomial, variant.RefIn)
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
					index := byte(crc>>56) ^ buffer[i]
					crc = (crc << 8) ^ table[index]
				}
			}
		}

		if err == io.EOF {
			break
		}
	}

	if variant.RefIn {
		if !variant.RefOut {
			crc = reverseBits64(crc)
		}
	} else {
		if variant.RefOut {
			crc = reverseBits64(crc)
		}
	}

	return crc ^ variant.XorOut, nil
}

func VerifyCRC64(data []byte, expected uint64, variant CRC64Variant) (bool, error) {
	crc, err := CalculateCRC64(data, variant, nil)
	if err != nil {
		return false, err
	}
	return crc == expected, nil
}

func VerifyCRC64Stream(reader io.Reader, expected uint64, variant CRC64Variant) (bool, error) {
	crc, err := CalculateCRC64Stream(reader, variant, nil, 4096)
	if err != nil {
		return false, err
	}
	return crc == expected, nil
}

func FormatCRC64(crc uint64) string {
	return fmt.Sprintf("0x%016X", crc)
}

func ParseCRC64(s string) (uint64, error) {
	var value uint64
	_, err := fmt.Sscanf(s, "0x%016X", &value)
	if err != nil {
		_, err = fmt.Sscanf(s, "0X%016X", &value)
		if err != nil {
			return 0, err
		}
	}
	return value, nil
}

func CRC64ToBytes(crc uint64, bigEndian bool) []byte {
	bytes := make([]byte, 8)
	if bigEndian {
		binary.BigEndian.PutUint64(bytes, crc)
	} else {
		binary.LittleEndian.PutUint64(bytes, crc)
	}
	return bytes
}

func BytesToCRC64(bytes []byte, bigEndian bool) (uint64, error) {
	if len(bytes) != 8 {
		return 0, errors.New("CRC64 value must be exactly 8 bytes")
	}
	if bigEndian {
		return binary.BigEndian.Uint64(bytes), nil
	}
	return binary.LittleEndian.Uint64(bytes), nil
}
