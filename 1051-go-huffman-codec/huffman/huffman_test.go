package huffman

import (
	"testing"
)

func TestSingleCharacter(t *testing.T) {
	tests := []string{
		"a",
		"aaaaa",
		"bbbbbbbb",
	}

	for _, original := range tests {
		t.Run("Test: "+original, func(t *testing.T) {
			freq := CountFrequencies(original)
			tree := BuildHuffmanTree(freq)

			if !tree.IsLeaf {
				t.Errorf("Expected tree to be a single leaf node")
			}

			codeTable := BuildCodeTable(tree)
			encoded, paddingBits, err := Encode(original, codeTable)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(encoded, paddingBits, tree)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded != original {
				t.Errorf("Decoded text mismatch. Expected: %q, Got: %q", original, decoded)
			}

			t.Logf("Original: %q, Decoded: %q ✓", original, decoded)
		})
	}
}

func TestEmptyString(t *testing.T) {
	original := ""

	freq := CountFrequencies(original)
	if len(freq) != 0 {
		t.Errorf("Expected empty frequency map, got size %d", len(freq))
	}

	tree := BuildHuffmanTree(freq)
	if tree != nil {
		t.Errorf("Expected nil tree for empty input")
	}

	codeTable := BuildCodeTable(tree)
	encoded, paddingBits, err := Encode(original, codeTable)

	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if len(encoded) != 0 {
		t.Errorf("Expected empty encoded data")
	}

	if paddingBits != 0 {
		t.Errorf("Expected 0 padding bits")
	}

	decoded, err := Decode(encoded, paddingBits, tree)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded != "" {
		t.Errorf("Expected empty decoded string")
	}

	t.Log("Empty string test passed ✓")
}

func TestNormalText(t *testing.T) {
	original := "The quick brown fox jumps over the lazy dog. Testing 123!"

	freq := CountFrequencies(original)
	tree := BuildHuffmanTree(freq)
	codeTable := BuildCodeTable(tree)

	encoded, paddingBits, err := Encode(original, codeTable)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded, paddingBits, tree)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded != original {
		t.Errorf("Decoded text mismatch. Expected: %q, Got: %q", original, decoded)
	}

	t.Logf("Normal text test passed ✓")
	t.Logf("  Original size: %d bytes", len(original))
	t.Logf("  Encoded size:  %d bytes", len(encoded))
}

func TestSpecialCharacters(t *testing.T) {
	tests := []string{
		"  \n\n\t\t",
		"Hello, World!\nNew line here.",
		"!!!@@@###$$$%%%",
		"   ",
	}

	for _, original := range tests {
		t.Run("Test special chars", func(t *testing.T) {
			freq := CountFrequencies(original)
			tree := BuildHuffmanTree(freq)
			codeTable := BuildCodeTable(tree)

			encoded, paddingBits, err := Encode(original, codeTable)
			if err != nil {
				t.Fatalf("Encode failed: %v", err)
			}

			decoded, err := Decode(encoded, paddingBits, tree)
			if err != nil {
				t.Fatalf("Decode failed: %v", err)
			}

			if decoded != original {
				t.Errorf("Decoded text mismatch")
			}

			t.Logf("Special characters test passed ✓")
		})
	}
}

func TestTreeSerialization(t *testing.T) {
	tests := []string{
		"a",
		"aaaaa",
		"hello world",
	}

	for _, original := range tests {
		t.Run("Serialize: "+original, func(t *testing.T) {
			freq := CountFrequencies(original)
			tree := BuildHuffmanTree(freq)

			serialized := SerializeTree(tree)
			deserialized, err := DeserializeTree(serialized)
			if err != nil {
				t.Fatalf("Deserialize failed: %v", err)
			}

			codeTable1 := BuildCodeTable(tree)
			codeTable2 := BuildCodeTable(deserialized)

			encoded1, padding1, _ := Encode(original, codeTable1)
			encoded2, padding2, _ := Encode(original, codeTable2)

			if len(encoded1) != len(encoded2) || padding1 != padding2 {
				t.Errorf("Serialization mismatch")
			}

			decoded, _ := Decode(encoded1, padding1, deserialized)
			if decoded != original {
				t.Errorf("Decoded with deserialized tree failed")
			}

			t.Logf("Tree serialization test passed ✓")
		})
	}
}
