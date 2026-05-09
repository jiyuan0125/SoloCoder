package api

import "time"

type ParseRequest struct {
	Armor string `json:"armor"`
}

type Subpacket struct {
	Type   uint8  `json:"type"`
	Length uint32 `json:"length"`
	Data   []byte `json:"data"`
}

type PacketInfo struct {
	Tag          string      `json:"tag"`
	TagValue     uint8       `json:"tag_value"`
	IsNewFormat  bool        `json:"is_new_format"`
	Length       uint64      `json:"length"`
	IsIndefinite bool        `json:"is_indefinite"`
	Subpackets   []Subpacket `json:"subpackets,omitempty"`
	HasBody      bool        `json:"has_body"`
	BodyPreview  string      `json:"body_preview,omitempty"`
}

type ArmorInfo struct {
	Type           string            `json:"type"`
	Headers        map[string]string `json:"headers"`
	ChecksumExists bool              `json:"checksum_exists"`
	ChecksumValid  bool              `json:"checksum_valid"`
	PayloadSize    int               `json:"payload_size"`
}

type ParseResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Armor   *ArmorInfo  `json:"armor,omitempty"`
	Packets []PacketInfo `json:"packets,omitempty"`
}

type KeyInfoRequest struct {
	Armor string `json:"armor"`
}

type PublicKeyInfo struct {
	Version      uint8     `json:"version"`
	CreationTime time.Time `json:"creation_time"`
	Algorithm    string    `json:"algorithm"`
	AlgorithmID  uint8     `json:"algorithm_id"`
	KeyID        string    `json:"key_id"`
	Fingerprint  string    `json:"fingerprint"`
}

type KeyInfoResponse struct {
	Success    bool             `json:"success"`
	Error      string           `json:"error,omitempty"`
	PrimaryKey *PublicKeyInfo   `json:"primary_key,omitempty"`
	Subkeys    []*PublicKeyInfo `json:"subkeys,omitempty"`
	UserIDs    []string         `json:"user_ids,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
