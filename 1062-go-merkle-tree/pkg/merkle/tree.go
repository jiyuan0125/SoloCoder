package merkle

import (
	"encoding/hex"
	"errors"
)

type MerkleTree struct {
	leaves    [][]byte
	levels    [][][]byte
	root      []byte
	leafCount int
}

func NewMerkleTree(leafHashes [][]byte) *MerkleTree {
	if len(leafHashes) == 0 {
		return &MerkleTree{
			leaves:    [][]byte{},
			levels:    [][][]byte{},
			root:      EmptyHash(),
			leafCount: 0,
		}
	}

	tree := &MerkleTree{
		leaves:    make([][]byte, len(leafHashes)),
		levels:    [][][]byte{},
		leafCount: len(leafHashes),
	}

	for i, h := range leafHashes {
		tree.leaves[i] = make([]byte, len(h))
		copy(tree.leaves[i], h)
	}

	currentLevel := make([][]byte, len(tree.leaves))
	for i, h := range tree.leaves {
		currentLevel[i] = make([]byte, len(h))
		copy(currentLevel[i], h)
	}

	tree.levels = append(tree.levels, currentLevel)

	for len(currentLevel) > 1 {
		nextLevel := make([][]byte, 0, (len(currentLevel)+1)/2)

		for i := 0; i < len(currentLevel); i += 2 {
			left := currentLevel[i]
			var right []byte

			if i+1 < len(currentLevel) {
				right = currentLevel[i+1]
			} else {
				right = make([]byte, len(left))
				copy(right, left)
			}

			parentHash := ConcatAndHash(left, right)
			nextLevel = append(nextLevel, parentHash)
		}

		tree.levels = append(tree.levels, nextLevel)
		currentLevel = nextLevel
	}

	tree.root = make([]byte, len(currentLevel[0]))
	copy(tree.root, currentLevel[0])

	return tree
}

func (t *MerkleTree) Root() []byte {
	if t.root == nil {
		return nil
	}
	rootCopy := make([]byte, len(t.root))
	copy(rootCopy, t.root)
	return rootCopy
}

func (t *MerkleTree) RootHex() string {
	return hex.EncodeToString(t.Root())
}

func (t *MerkleTree) LeafCount() int {
	return t.leafCount
}

func (t *MerkleTree) GetLeaf(index int) []byte {
	if index < 0 || index >= t.leafCount {
		return nil
	}
	leafCopy := make([]byte, len(t.leaves[index]))
	copy(leafCopy, t.leaves[index])
	return leafCopy
}

type ProofStep struct {
	Hash      []byte
	IsRight   bool
}

func (t *MerkleTree) GetProof(index int) ([]ProofStep, error) {
	if t.leafCount == 0 {
		return nil, errors.New("tree has no leaves")
	}
	if index < 0 || index >= t.leafCount {
		return nil, errors.New("index out of bounds")
	}

	if t.leafCount == 1 {
		return []ProofStep{}, nil
	}

	proof := []ProofStep{}
	currentIndex := index

	for levelIdx := 0; levelIdx < len(t.levels)-1; levelIdx++ {
		level := t.levels[levelIdx]
		var siblingIndex int
		var isRight bool

		if currentIndex%2 == 0 {
			siblingIndex = currentIndex + 1
			isRight = true
			if siblingIndex >= len(level) {
				siblingIndex = currentIndex
			}
		} else {
			siblingIndex = currentIndex - 1
			isRight = false
		}

		siblingHash := make([]byte, len(level[siblingIndex]))
		copy(siblingHash, level[siblingIndex])

		proof = append(proof, ProofStep{
			Hash:    siblingHash,
			IsRight: isRight,
		})

		currentIndex = currentIndex / 2
	}

	return proof, nil
}

func VerifyProof(leafHash []byte, proof []ProofStep, expectedRoot []byte) bool {
	currentHash := make([]byte, len(leafHash))
	copy(currentHash, leafHash)

	for _, step := range proof {
		siblingHash := step.Hash

		if step.IsRight {
			currentHash = ConcatAndHash(currentHash, siblingHash)
		} else {
			currentHash = ConcatAndHash(siblingHash, currentHash)
		}
	}

	if len(currentHash) != len(expectedRoot) {
		return false
	}

	for i := range currentHash {
		if currentHash[i] != expectedRoot[i] {
			return false
		}
	}

	return true
}

func (t *MerkleTree) FindDifferences(other *MerkleTree) []int {
	if t.leafCount == 0 || other.leafCount == 0 {
		if t.leafCount == 0 && other.leafCount == 0 {
			return []int{}
		}
		maxCount := t.leafCount
		if other.leafCount > maxCount {
			maxCount = other.leafCount
		}
		diffs := make([]int, 0, maxCount)
		for i := 0; i < maxCount; i++ {
			if i >= t.leafCount || i >= other.leafCount {
				diffs = append(diffs, i)
				continue
			}
			if !equalBytes(t.leaves[i], other.leaves[i]) {
				diffs = append(diffs, i)
			}
		}
		return diffs
	}

	if equalBytes(t.root, other.root) {
		return []int{}
	}

	if len(t.levels) == 0 || len(other.levels) == 0 {
		diffs := make([]int, 0, t.leafCount)
		for i := 0; i < t.leafCount; i++ {
			if i >= other.leafCount || !equalBytes(t.leaves[i], other.leaves[i]) {
				diffs = append(diffs, i)
			}
		}
		return diffs
	}

	diffs := []int{}
	t.findDiffsRecursive(len(t.levels)-1, 0, len(other.levels)-1, 0, other, &diffs)
	return diffs
}

func (t *MerkleTree) findDiffsRecursive(
	tLevelIdx, tNodeIdx int,
	oLevelIdx, oNodeIdx int,
	other *MerkleTree,
	diffs *[]int,
) {
	if tLevelIdx < 0 || oLevelIdx < 0 {
		return
	}

	tLevel := t.levels[tLevelIdx]
	oLevel := other.levels[oLevelIdx]

	if tNodeIdx >= len(tLevel) || oNodeIdx >= len(oLevel) {
		return
	}

	tNode := tLevel[tNodeIdx]
	oNode := oLevel[oNodeIdx]

	if equalBytes(tNode, oNode) {
		return
	}

	if tLevelIdx == 0 && oLevelIdx == 0 {
		if tNodeIdx < t.leafCount && tNodeIdx < other.leafCount {
			*diffs = append(*diffs, tNodeIdx)
		} else if tNodeIdx < t.leafCount || tNodeIdx < other.leafCount {
			if tNodeIdx < t.leafCount {
				*diffs = append(*diffs, tNodeIdx)
			}
		}
		return
	}

	if tLevelIdx == 0 {
		baseRange := getLeafRange(oLevelIdx, oNodeIdx, other.levels)
		for i := baseRange.start; i < baseRange.end && i < other.leafCount; i++ {
			if i >= t.leafCount {
				*diffs = append(*diffs, i)
			}
		}
		return
	}

	if oLevelIdx == 0 {
		baseRange := getLeafRange(tLevelIdx, tNodeIdx, t.levels)
		for i := baseRange.start; i < baseRange.end && i < t.leafCount; i++ {
			if i >= other.leafCount || !equalBytes(t.leaves[i], other.GetLeaf(i)) {
				*diffs = append(*diffs, i)
			}
		}
		return
	}

	tLeftChildIdx := tNodeIdx * 2
	tRightChildIdx := tLeftChildIdx + 1
	oLeftChildIdx := oNodeIdx * 2
	oRightChildIdx := oLeftChildIdx + 1

	t.findDiffsRecursive(tLevelIdx-1, tLeftChildIdx, oLevelIdx-1, oLeftChildIdx, other, diffs)
	t.findDiffsRecursive(tLevelIdx-1, tRightChildIdx, oLevelIdx-1, oRightChildIdx, other, diffs)
}

type leafRange struct {
	start int
	end   int
}

func getLeafRange(levelIdx, nodeIdx int, levels [][][]byte) leafRange {
	leavesPerNode := 1
	for i := 0; i < levelIdx; i++ {
		leavesPerNode *= 2
	}
	return leafRange{
		start: nodeIdx * leavesPerNode,
		end:   (nodeIdx + 1) * leavesPerNode,
	}
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
