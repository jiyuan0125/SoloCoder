package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/solocoder/openpgp-parser/api"
	"github.com/solocoder/openpgp-parser/openpgp"
)

func main() {
	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/keyinfo", handleKeyInfo)

	port := ":8080"
	log.Printf("OpenPGP Parser Server listening on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed, use POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.ParseRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, "failed to parse JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Armor == "" {
		writeError(w, "armor field is required", http.StatusBadRequest)
		return
	}

	armor, packets, err := openpgp.ParseArmorAndPackets([]byte(req.Armor))
	if err != nil {
		writeError(w, "failed to parse: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := api.ParseResponse{
		Success: true,
		Armor: &api.ArmorInfo{
			Type:           string(armor.Type),
			Headers:        armor.Headers,
			ChecksumExists: armor.ChecksumExists,
			ChecksumValid:  armor.ChecksumValid,
			PayloadSize:    len(armor.Payload),
		},
		Packets: make([]api.PacketInfo, len(packets)),
	}

	for i, packet := range packets {
		packetInfo := api.PacketInfo{
			Tag:          packet.Tag.String(),
			TagValue:     packet.TagValue,
			IsNewFormat:  packet.IsNewFormat,
			Length:       packet.Length,
			IsIndefinite: packet.IsIndefinite,
			HasBody:      len(packet.Body) > 0,
		}

		if len(packet.Subpackets) > 0 {
			packetInfo.Subpackets = make([]api.Subpacket, len(packet.Subpackets))
			for j, sp := range packet.Subpackets {
				packetInfo.Subpackets[j] = api.Subpacket{
					Type:   sp.Type,
					Length: sp.Length,
					Data:   sp.Data,
				}
			}
		}

		if len(packet.Body) > 0 {
			previewLen := 64
			if len(packet.Body) < previewLen {
				previewLen = len(packet.Body)
			}
			packetInfo.BodyPreview = hex.EncodeToString(packet.Body[:previewLen])
			if len(packet.Body) > previewLen {
				packetInfo.BodyPreview += "..."
			}
		}

		response.Packets[i] = packetInfo
	}

	writeJSON(w, response, http.StatusOK)
}

func handleKeyInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed, use POST", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, "failed to read request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req api.KeyInfoRequest
	if len(body) > 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, "failed to parse JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	if req.Armor == "" {
		writeError(w, "armor field is required", http.StatusBadRequest)
		return
	}

	armor, packets, err := openpgp.ParseArmorAndPackets([]byte(req.Armor))
	if err != nil {
		writeError(w, "failed to parse armor: "+err.Error(), http.StatusBadRequest)
		return
	}

	if armor.Type != openpgp.ArmorTypePublicKey && armor.Type != openpgp.ArmorTypePrivateKey {
		writeError(w, fmt.Sprintf("expected PUBLIC KEY BLOCK or PRIVATE KEY BLOCK, got %s", armor.Type), http.StatusBadRequest)
		return
	}

	keyInfo, err := openpgp.ParseKey(packets)
	if err != nil {
		writeError(w, "failed to parse key: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := api.KeyInfoResponse{
		Success: true,
		UserIDs: keyInfo.UserIDs,
	}

	if keyInfo.PrimaryKey != nil {
		response.PrimaryKey = &api.PublicKeyInfo{
			Version:      keyInfo.PrimaryKey.Version,
			CreationTime: keyInfo.PrimaryKey.CreationTime,
			Algorithm:    keyInfo.PrimaryKey.Algorithm.String(),
			AlgorithmID:  keyInfo.PrimaryKey.AlgorithmID,
			KeyID:        keyInfo.PrimaryKey.KeyID,
			Fingerprint:  keyInfo.PrimaryKey.Fingerprint,
		}
	}

	if len(keyInfo.Subkeys) > 0 {
		response.Subkeys = make([]*api.PublicKeyInfo, len(keyInfo.Subkeys))
		for i, sk := range keyInfo.Subkeys {
			response.Subkeys[i] = &api.PublicKeyInfo{
				Version:      sk.Version,
				CreationTime: sk.CreationTime,
				Algorithm:    sk.Algorithm.String(),
				AlgorithmID:  sk.AlgorithmID,
				KeyID:        sk.KeyID,
				Fingerprint:  sk.Fingerprint,
			}
		}
	}

	writeJSON(w, response, http.StatusOK)
}

func writeError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.ErrorResponse{
		Success: false,
		Error:   message,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
