package api

import "yaml-anchor-resolver/pkg/anchor"

type ResolveRequest struct {
	YAMLText string `json:"yaml_text"`
}

type ResolveResponse struct {
	Success  bool                      `json:"success"`
	Data     interface{}               `json:"data,omitempty"`
	Anchors  []anchor.AnchorInfo       `json:"anchors,omitempty"`
	Refs     []anchor.ReferenceInfo    `json:"references,omitempty"`
	Relations []anchor.ReferenceRelation `json:"relations,omitempty"`
	Warnings []string                  `json:"warnings,omitempty"`
	Error    string                    `json:"error,omitempty"`
}

type ListAnchorsRequest struct {
	YAMLText string `json:"yaml_text"`
}

type ListAnchorsResponse struct {
	Success   bool                      `json:"success"`
	Anchors   []anchor.AnchorInfo       `json:"anchors,omitempty"`
	Refs      []anchor.ReferenceInfo    `json:"references,omitempty"`
	Relations []anchor.ReferenceRelation `json:"relations,omitempty"`
	Error     string                    `json:"error,omitempty"`
}
