package punycode

import (
	"strings"
)

func EncodeDomain(domain string) (string, error) {
	labels := strings.Split(domain, ".")
	result := make([]string, 0, len(labels))

	for _, label := range labels {
		encoded, err := Encode(label)
		if err != nil {
			return "", err
		}
		result = append(result, encoded)
	}

	return strings.Join(result, "."), nil
}

func DecodeDomain(domain string) (string, error) {
	labels := strings.Split(domain, ".")
	result := make([]string, 0, len(labels))

	for _, label := range labels {
		decoded, err := Decode(label)
		if err != nil {
			return "", err
		}
		result = append(result, decoded)
	}

	return strings.Join(result, "."), nil
}
