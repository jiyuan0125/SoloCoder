package openpgp

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"io"
	"strings"
)

type ArmorType string

const (
	ArmorTypeMessage        ArmorType = "MESSAGE"
	ArmorTypePublicKey      ArmorType = "PUBLIC KEY BLOCK"
	ArmorTypePrivateKey     ArmorType = "PRIVATE KEY BLOCK"
	ArmorTypeSignedMessage  ArmorType = "SIGNED MESSAGE"
	ArmorTypeSignature      ArmorType = "SIGNATURE"
)

const (
	beginPrefix    = "-----BEGIN PGP "
	endPrefix      = "-----END PGP "
	armorSuffix    = "-----"
	checksumPrefix = "="
)

type Armor struct {
	Type           ArmorType
	Headers        map[string]string
	Payload        []byte
	Checksum       []byte
	ChecksumExists bool
	ChecksumValid  bool
}

func ParseArmor(data []byte) (*Armor, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var (
		beginFound   bool
		headers      = make(map[string]string)
		payloadParts []string
		checksumPart string
		result       Armor
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, beginPrefix) && strings.HasSuffix(line, armorSuffix) {
			beginFound = true
			typeStr := line[len(beginPrefix) : len(line)-len(armorSuffix)]
			result.Type = ArmorType(typeStr)
			continue
		}

		if !beginFound {
			continue
		}

		if strings.HasPrefix(line, endPrefix) && strings.HasSuffix(line, armorSuffix) {
			break
		}

		if strings.HasPrefix(line, checksumPrefix) && len(line) == 5 {
			checksumPart = line[1:]
			continue
		}

		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				headers[key] = value
				continue
			}
		}

		if strings.ContainsAny(line, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=") {
			payloadParts = append(payloadParts, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !beginFound {
		return nil, errors.New("no PGP armor begin marker found")
	}

	if len(payloadParts) == 0 {
		return nil, errors.New("no payload data found in armor")
	}

	result.Headers = headers

	payloadBase64 := strings.Join(payloadParts, "")
	payload, err := base64.StdEncoding.DecodeString(payloadBase64)
	if err != nil {
		return nil, errors.New("failed to decode base64 payload: " + err.Error())
	}
	result.Payload = payload

	if checksumPart != "" {
		result.ChecksumExists = true
		decodedChecksum, err := base64.StdEncoding.DecodeString(checksumPart)
		if err != nil {
			return nil, errors.New("failed to decode checksum: " + err.Error())
		}
		result.Checksum = decodedChecksum

		expected := crc24(payload)
		result.ChecksumValid = bytes.Equal(result.Checksum, expected)
	}

	return &result, nil
}

func ParseArmorReader(r io.Reader) (*Armor, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseArmor(data)
}

func crc24(data []byte) []byte {
	crc := uint32(0xB704CE)
	for _, b := range data {
		crc ^= uint32(b) << 16
		for i := 0; i < 8; i++ {
			crc <<= 1
			if crc&0x1000000 != 0 {
				crc ^= 0x1864CFB
			}
		}
	}
	crc &= 0xFFFFFF
	return []byte{
		byte(crc >> 16),
		byte(crc >> 8),
		byte(crc),
	}
}
