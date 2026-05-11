package asn1

import (
	"encoding/hex"
	"fmt"
)

func EncodeDER(jt JSONTLV) (string, error) {
	tlv, err := FromJSON(jt)
	if err != nil {
		return "", err
	}

	bytes, err := Encode(tlv, DefaultEncodeOptions(ModeDER))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func EncodeBER(jt JSONTLV) (string, error) {
	tlv, err := FromJSON(jt)
	if err != nil {
		return "", err
	}

	bytes, err := Encode(tlv, DefaultEncodeOptions(ModeBER))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func DecodeHex(hexStr string) ([]JSONTLV, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}

	tlvs, err := DecodeAll(data)
	if err != nil {
		return nil, err
	}

	result := make([]JSONTLV, 0, len(tlvs))
	for _, tlv := range tlvs {
		jt, err := tlv.ToJSON()
		if err != nil {
			return nil, err
		}
		result = append(result, jt)
	}

	return result, nil
}
