package api

type BuildTreeRequest struct {
	LeafHashes []string `json:"leaf_hashes"`
}

type BuildTreeResponse struct {
	Success    bool   `json:"success"`
	RootHash   string `json:"root_hash,omitempty"`
	Error      string `json:"error,omitempty"`
	TreeID     string `json:"tree_id,omitempty"`
}

type VerifyProofRequest struct {
	TreeID      string       `json:"tree_id"`
	LeafIndex   int          `json:"leaf_index"`
	LeafHash    string       `json:"leaf_hash"`
	Proof       []ProofStep  `json:"proof"`
}

type ProofStep struct {
	Hash    string `json:"hash"`
	IsRight bool   `json:"is_right"`
}

type VerifyProofResponse struct {
	Success bool   `json:"success"`
	Valid   bool   `json:"valid,omitempty"`
	Error   string `json:"error,omitempty"`
}

type FindDifferencesRequest struct {
	SourceTreeID string `json:"source_tree_id"`
	TargetRootHash string `json:"target_root_hash"`
	TargetLeafHashes []string `json:"target_leaf_hashes,omitempty"`
}

type FindDifferencesResponse struct {
	Success      bool   `json:"success"`
	Differences  []int  `json:"differences,omitempty"`
	Error        string `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
