package main

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"merkle-tree/pkg/api"
	"merkle-tree/pkg/merkle"
)

type Handler struct {
	store *TreeStore
}

func NewHandler(store *TreeStore) *Handler {
	return &Handler{store: store}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) BuildTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.BuildTreeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	leafHashes := make([][]byte, 0, len(req.LeafHashes))
	for _, hashStr := range req.LeafHashes {
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid hex string in leaf_hashes",
			})
			return
		}
		if len(hashBytes) != 32 {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid hash length, expected 32 bytes (SHA-256)",
			})
			return
		}
		leafHashes = append(leafHashes, hashBytes)
	}

	tree := merkle.NewMerkleTree(leafHashes)
	treeID := h.store.Store(tree)

	writeJSON(w, http.StatusOK, api.BuildTreeResponse{
		Success:  true,
		RootHash: tree.RootHex(),
		TreeID:   treeID,
	})
}

func (h *Handler) VerifyProof(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.VerifyProofRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	tree, ok := h.store.Get(req.TreeID)
	if !ok {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{
			Success: false,
			Error:   "tree not found",
		})
		return
	}

	leafHash, err := hex.DecodeString(req.LeafHash)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid leaf_hash hex string",
		})
		return
	}
	if len(leafHash) != 32 {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid leaf_hash length",
		})
		return
	}

	proof := make([]merkle.ProofStep, 0, len(req.Proof))
	for _, step := range req.Proof {
		stepHash, err := hex.DecodeString(step.Hash)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid proof step hash",
			})
			return
		}
		if len(stepHash) != 32 {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid proof step hash length",
			})
			return
		}
		proof = append(proof, merkle.ProofStep{
			Hash:    stepHash,
			IsRight: step.IsRight,
		})
	}

	valid := merkle.VerifyProof(leafHash, proof, tree.Root())

	writeJSON(w, http.StatusOK, api.VerifyProofResponse{
		Success: true,
		Valid:   valid,
	})
}

func (h *Handler) FindDifferences(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ErrorResponse{
			Success: false,
			Error:   "method not allowed",
		})
		return
	}

	var req api.FindDifferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "invalid request body",
		})
		return
	}

	sourceTree, ok := h.store.Get(req.SourceTreeID)
	if !ok {
		writeJSON(w, http.StatusNotFound, api.ErrorResponse{
			Success: false,
			Error:   "source tree not found",
		})
		return
	}

	if len(req.TargetLeafHashes) == 0 {
		targetRoot, err := hex.DecodeString(req.TargetRootHash)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid target_root_hash hex string",
			})
			return
		}
		if len(targetRoot) != 32 {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid target_root_hash length",
			})
			return
		}

		sourceRoot := sourceTree.Root()
		rootEqual := true
		for i := range sourceRoot {
			if sourceRoot[i] != targetRoot[i] {
				rootEqual = false
				break
			}
		}

		if rootEqual {
			writeJSON(w, http.StatusOK, api.FindDifferencesResponse{
				Success:     true,
				Differences: []int{},
			})
			return
		}

		writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
			Success: false,
			Error:   "target_leaf_hashes is required when root hashes differ",
		})
		return
	}

	targetLeafHashes := make([][]byte, 0, len(req.TargetLeafHashes))
	for _, hashStr := range req.TargetLeafHashes {
		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid hex string in target_leaf_hashes",
			})
			return
		}
		if len(hashBytes) != 32 {
			writeJSON(w, http.StatusBadRequest, api.ErrorResponse{
				Success: false,
				Error:   "invalid hash length in target_leaf_hashes",
			})
			return
		}
		targetLeafHashes = append(targetLeafHashes, hashBytes)
	}

	targetTree := merkle.NewMerkleTree(targetLeafHashes)
	differences := sourceTree.FindDifferences(targetTree)

	writeJSON(w, http.StatusOK, api.FindDifferencesResponse{
		Success:     true,
		Differences: differences,
	})
}
