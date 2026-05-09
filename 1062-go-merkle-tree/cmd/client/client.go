package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"merkle-tree/pkg/api"
	"merkle-tree/pkg/merkle"
)

type Client struct {
	serverAddr string
	httpClient *http.Client
}

func NewClient(serverAddr string) *Client {
	return &Client{
		serverAddr: serverAddr,
		httpClient: &http.Client{},
	}
}

func (c *Client) BuildTree(leafHashes []string) (*api.BuildTreeResponse, error) {
	req := api.BuildTreeRequest{
		LeafHashes: leafHashes,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		c.serverAddr+"/api/build-tree",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.BuildTreeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return &result, nil
}

func (c *Client) VerifyProof(treeID string, leafIndex int, leafHash string, allLeafHashes []string) (*api.VerifyProofResponse, error) {
	leafHashesBytes := make([][]byte, 0, len(allLeafHashes))
	for _, h := range allLeafHashes {
		b, _ := hex.DecodeString(h)
		leafHashesBytes = append(leafHashesBytes, b)
	}

	localTree := merkle.NewMerkleTree(leafHashesBytes)
	proof, err := localTree.GetProof(leafIndex)
	if err != nil {
		return nil, err
	}

	proofSteps := make([]api.ProofStep, 0, len(proof))
	for _, step := range proof {
		proofSteps = append(proofSteps, api.ProofStep{
			Hash:    hex.EncodeToString(step.Hash),
			IsRight: step.IsRight,
		})
	}

	req := api.VerifyProofRequest{
		TreeID:    treeID,
		LeafIndex: leafIndex,
		LeafHash:  leafHash,
		Proof:     proofSteps,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		c.serverAddr+"/api/verify-proof",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.VerifyProofResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return &result, nil
}

func (c *Client) FindDifferences(sourceTreeID string, targetLeafHashes []string) (*api.FindDifferencesResponse, error) {
	targetLeafHashesBytes := make([][]byte, 0, len(targetLeafHashes))
	for _, h := range targetLeafHashes {
		b, _ := hex.DecodeString(h)
		targetLeafHashesBytes = append(targetLeafHashesBytes, b)
	}

	targetTree := merkle.NewMerkleTree(targetLeafHashesBytes)

	req := api.FindDifferencesRequest{
		SourceTreeID:     sourceTreeID,
		TargetRootHash:   targetTree.RootHex(),
		TargetLeafHashes: targetLeafHashes,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(
		c.serverAddr+"/api/find-differences",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result api.FindDifferencesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("invalid response: %s", string(respBody))
	}

	if !result.Success {
		return nil, fmt.Errorf("server error: %s", result.Error)
	}

	return &result, nil
}
