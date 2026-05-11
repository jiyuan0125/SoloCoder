package siphash

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSipHash24_Vectors(t *testing.T) {
	keyBytes, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	key, _ := NewKeyFromBytes(keyBytes)

	expected := []uint64{
		0x726fdb47dd0e0e31, // 0 bytes
		0x74f839c593dc67fd, // 1 byte
		0x0d6c8009d9a94f5a, // 2 bytes
		0x85676696d7fb7e2d, // 3 bytes
		0xcf2794e0277187b7, // 4 bytes
		0x18765564cd99a68d, // 5 bytes
		0xcbc9466e58fee3ce, // 6 bytes
		0xab0200f58b01d137, // 7 bytes
		0x93f5f5799a932462, // 8 bytes
		0x9e0082df0ba9e4b0, // 9 bytes
		0x7a5dbbc594ddb9f3, // 10 bytes
		0xf4b32f46226bada7, // 11 bytes
		0x751e8fbc860ee5fb, // 12 bytes
		0x14ea5627c0843d90, // 13 bytes
		0xf723ca908e7af2ee, // 14 bytes
		0xa129ca6149be45e5, // 15 bytes
		0x3f2acc7f57c29bdb, // 16 bytes
	}

	for i := 0; i < len(expected); i++ {
		input := make([]byte, i)
		for j := 0; j < i; j++ {
			input[j] = byte(j)
		}
		got := Sum64(key, input)
		if got != expected[i] {
			t.Errorf("SipHash-2-4(%d bytes): got 0x%016x, expected 0x%016x", i, got, expected[i])
		} else {
			fmt.Printf("PASS: SipHash-2-4(%d bytes): 0x%016x\n", i, got)
		}
	}
}

func TestSipHash13_Vectors(t *testing.T) {
	keyBytes, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f")
	key, _ := NewKeyFromBytes(keyBytes)

	expected := []uint64{
		0xabac0158050fc4dc, // 0 bytes
		0xc9f49bf37d57ca93, // 1 byte
		0x82cb9b024dc7d44d, // 2 bytes
		0x8bf80ab8e7ddf7fb, // 3 bytes
		0xcf75576088d38328, // 4 bytes
		0xdef9d52f49533b67, // 5 bytes
		0xc50d2b50c59f22a7, // 6 bytes
		0xd3927d989bb11140, // 7 bytes
		0x369095118d299a8e, // 8 bytes
		0x25a48eb36c063de4, // 9 bytes
		0x79de85ee92ff097f, // 10 bytes
		0x70c118c1f94dc352, // 11 bytes
		0x78a384b157b4d9a2, // 12 bytes
		0x306f760c1229ffa7, // 13 bytes
		0x605aa111c0f95d34, // 14 bytes
		0xd320d86d2a519956, // 15 bytes
		0xcc4fdd1a7d908b66, // 16 bytes
	}

	for i := 0; i < len(expected); i++ {
		input := make([]byte, i)
		for j := 0; j < i; j++ {
			input[j] = byte(j)
		}
		got := Sum64WithRounds(key, input, 1, 3)
		if got != expected[i] {
			t.Errorf("SipHash-1-3(%d bytes): got 0x%016x, expected 0x%016x", i, got, expected[i])
		} else {
			fmt.Printf("PASS: SipHash-1-3(%d bytes): 0x%016x\n", i, got)
		}
	}
}
