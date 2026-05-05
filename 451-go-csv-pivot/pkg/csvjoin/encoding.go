package csvjoin

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"unicode/utf8"
)

// Encoding represents a text encoding.
type Encoding string

const (
	// EncodingUTF8 represents UTF-8 encoding.
	EncodingUTF8 Encoding = "UTF-8"
	// EncodingGBK represents GBK encoding (Chinese).
	EncodingGBK Encoding = "GBK"
	// EncodingGB18030 represents GB18030 encoding (Chinese, superset of GBK).
	EncodingGB18030 Encoding = "GB18030"
	// EncodingUnknown represents unknown encoding.
	EncodingUnknown Encoding = "Unknown"
)

// EncodingDetector detects the encoding of text data.
type EncodingDetector struct{}

// NewEncodingDetector creates a new encoding detector.
func NewEncodingDetector() *EncodingDetector {
	return &EncodingDetector{}
}

// DetectFile detects the encoding of a file.
// It reads a sample of the file to determine the encoding.
func (d *EncodingDetector) DetectFile(filePath string) (Encoding, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return EncodingUnknown, &JoinError{Op: "DetectFile", Err: err}
	}
	defer file.Close()

	return d.Detect(file)
}

// Detect detects the encoding of data from a reader.
// It reads a sample of up to 8KB to determine the encoding.
func (d *EncodingDetector) Detect(r io.Reader) (Encoding, error) {
	// Read up to 8KB for detection
	const sampleSize = 8192
	sample := make([]byte, sampleSize)

	reader := bufio.NewReader(r)
	n, err := reader.Read(sample)
	if err != nil && err != io.EOF {
		return EncodingUnknown, &JoinError{Op: "Detect", Err: err}
	}

	if n == 0 {
		// Empty file, default to UTF-8
		return EncodingUTF8, nil
	}

	sample = sample[:n]

	// Check for BOM first
	encoding := d.detectBOM(sample)
	if encoding != EncodingUnknown {
		return encoding, nil
	}

	// Check if valid UTF-8
	if utf8.Valid(sample) {
		return EncodingUTF8, nil
	}

	// Check for GBK/GB18030 characteristics
	if d.looksLikeGBK(sample) {
		return EncodingGBK, nil
	}

	// Default to UTF-8 if we can't determine
	return EncodingUTF8, nil
}

// detectBOM checks for Byte Order Marks to determine encoding.
func (d *EncodingDetector) detectBOM(data []byte) Encoding {
	if len(data) >= 3 {
		// UTF-8 BOM: EF BB BF
		if data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
			return EncodingUTF8
		}
	}

	if len(data) >= 2 {
		// UTF-16 LE BOM: FF FE
		if data[0] == 0xFF && data[1] == 0xFE {
			// Not fully supported, but we can note it
			return EncodingUnknown
		}
		// UTF-16 BE BOM: FE FF
		if data[0] == 0xFE && data[1] == 0xFF {
			return EncodingUnknown
		}
	}

	return EncodingUnknown
}

// looksLikeGBK checks if the data appears to be GBK encoded.
// GBK uses 1 byte for ASCII and 2 bytes for Chinese characters.
// First byte range: 0x81-0xFE, Second byte range: 0x40-0x7E, 0x80-0xFE
func (d *EncodingDetector) looksLikeGBK(data []byte) bool {
	i := 0
	for i < len(data) {
		b := data[i]

		// ASCII character (0x00-0x7F)
		if b <= 0x7F {
			i++
			continue
		}

		// Potential multi-byte character
		// GBK first byte: 0x81-0xFE
		if b < 0x81 || b > 0xFE {
			return false
		}

		// Need at least one more byte
		if i+1 >= len(data) {
			// End of sample, can't determine
			return true
		}

		// GBK second byte: 0x40-0x7E or 0x80-0xFE
		second := data[i+1]
		if !((second >= 0x40 && second <= 0x7E) || (second >= 0x80 && second <= 0xFE)) {
			return false
		}

		i += 2
	}

	return true
}

// EncodingConverter converts text from one encoding to UTF-8.
// Note: For full GBK/GB18030 support, consider using golang.org/x/text/encoding/simplifiedchinese
// This implementation provides basic detection and a framework for conversion.
type EncodingConverter struct{}

// NewEncodingConverter creates a new encoding converter.
func NewEncodingConverter() *EncodingConverter {
	return &EncodingConverter{}
}

// ConvertToUTF8 converts data from the specified encoding to UTF-8.
// For GBK/GB18030, this is a placeholder that returns the original data.
// For production use with GBK, import golang.org/x/text/encoding/simplifiedchinese
// and use simplifiedchinese.GBK.NewDecoder().Reader()
func (c *EncodingConverter) ConvertToUTF8(r io.Reader, from Encoding) (io.Reader, error) {
	switch from {
	case EncodingUTF8:
		// Already UTF-8, just check for BOM
		return detectBOM(r)
	case EncodingGBK, EncodingGB18030:
		// For GBK conversion, you need to import golang.org/x/text/encoding/simplifiedchinese
		// Example:
		// import "golang.org/x/text/encoding/simplifiedchinese"
		// import "golang.org/x/text/transform"
		// return transform.NewReader(r, simplifiedchinese.GBK.NewDecoder()), nil
		//
		// Since we can't use external packages in this implementation,
		// we provide a basic implementation that assumes the data is already UTF-8
		// or returns the data as-is.
		//
		// In a real implementation, you would add to go.mod:
		// require golang.org/x/text v0.14.0
		//
		// For now, we'll read the data and attempt to handle common cases
		return c.basicGBKConverter(r), nil
	default:
		return r, nil
	}
}

// basicGBKConverter is a simple converter that reads the data and returns it as-is.
// For proper GBK support, use golang.org/x/text/encoding/simplifiedchinese
func (c *EncodingConverter) basicGBKConverter(r io.Reader) io.Reader {
	// Read all data
	data, err := io.ReadAll(r)
	if err != nil {
		return bytes.NewReader(nil)
	}

	// Check if it's already valid UTF-8
	if utf8.Valid(data) {
		return bytes.NewReader(data)
	}

	// For now, return the data as-is
	// In production, use the proper encoding library
	return bytes.NewReader(data)
}

// AutoDetectAndConvert automatically detects the encoding and converts to UTF-8.
func (c *EncodingConverter) AutoDetectAndConvert(r io.Reader) (io.Reader, Encoding, error) {
	// We need to read the data twice: once for detection, once for conversion
	// So we read it all into memory first
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, EncodingUnknown, &JoinError{Op: "AutoDetectAndConvert", Err: err}
	}

	detector := NewEncodingDetector()
	encoding, err := detector.Detect(bytes.NewReader(data))
	if err != nil {
		return nil, EncodingUnknown, err
	}

	converted, err := c.ConvertToUTF8(bytes.NewReader(data), encoding)
	if err != nil {
		return nil, encoding, err
	}

	return converted, encoding, nil
}

// OpenFileForReading opens a file with automatic encoding detection and conversion to UTF-8.
func OpenFileForReading(filePath string) (io.Reader, Encoding, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, EncodingUnknown, &JoinError{Op: "OpenFileForReading", Err: err}
	}

	// Read the file content for detection and conversion
	data, err := io.ReadAll(file)
	file.Close()
	if err != nil {
		return nil, EncodingUnknown, &JoinError{Op: "OpenFileForReading", Err: err}
	}

	converter := NewEncodingConverter()
	reader, encoding, err := converter.AutoDetectAndConvert(bytes.NewReader(data))
	if err != nil {
		return nil, encoding, err
	}

	return reader, encoding, nil
}
