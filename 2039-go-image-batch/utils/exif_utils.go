package utils

import (
	"bytes"
	"io"
	"os"
)

type ExifData struct {
	RawData []byte
	Format  string
}

func ExtractExifData(file *os.File, format string) (*ExifData, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	exifData := extractExifFromBuffer(data, format)

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	if exifData != nil && len(exifData) > 0 {
		return &ExifData{
			RawData: exifData,
			Format:  format,
		}, nil
	}

	return nil, nil
}

func extractExifFromBuffer(data []byte, format string) []byte {
	switch format {
	case "jpeg":
		return extractExifFromJPEG(data)
	case "webp":
		return extractExifFromWebP(data)
	case "png":
		return extractExifFromPNG(data)
	}
	return nil
}

func extractExifFromJPEG(data []byte) []byte {
	if len(data) < 4 {
		return nil
	}

	if data[0] != 0xFF || data[1] != 0xD8 {
		return nil
	}

	offset := 2
	for offset < len(data)-1 {
		if data[offset] != 0xFF {
			return nil
		}

		marker := data[offset+1]

		if marker == 0xD9 {
			return nil
		}

		if marker == 0xE1 {
			if offset+4 >= len(data) {
				return nil
			}

			segmentLen := int(data[offset+2])<<8 | int(data[offset+3])
			if offset+segmentLen > len(data) {
				return nil
			}

			exifHeader := []byte("Exif\x00\x00")
			exifStart := offset + 4
			if exifStart+6 <= offset+segmentLen &&
				bytes.Equal(data[exifStart:exifStart+6], exifHeader) {
				return data[offset : offset+segmentLen]
			}
		}

		if offset+4 >= len(data) {
			return nil
		}

		segmentLen := int(data[offset+2])<<8 | int(data[offset+3])
		offset += segmentLen
	}

	return nil
}

func extractExifFromWebP(data []byte) []byte {
	if len(data) < 12 {
		return nil
	}

	if !bytes.Equal(data[0:4], []byte("RIFF")) {
		return nil
	}
	if !bytes.Equal(data[8:12], []byte("WEBP")) {
		return nil
	}

	offset := 12
	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := int(data[offset+4]) | int(data[offset+5])<<8 |
			int(data[offset+6])<<16 | int(data[offset+7])<<24
		chunkDataStart := offset + 8
		chunkDataEnd := chunkDataStart + chunkSize

		if chunkDataEnd > len(data) {
			return nil
		}

		if bytes.Equal(chunkID, []byte("EXIF")) {
			chunk := make([]byte, 8+chunkSize)
			copy(chunk[0:4], chunkID)
			chunk[4] = byte(chunkSize & 0xFF)
			chunk[5] = byte((chunkSize >> 8) & 0xFF)
			chunk[6] = byte((chunkSize >> 16) & 0xFF)
			chunk[7] = byte((chunkSize >> 24) & 0xFF)
			copy(chunk[8:], data[chunkDataStart:chunkDataEnd])
			return chunk
		}

		offset = chunkDataEnd
		if chunkSize%2 == 1 {
			offset++
		}
	}

	return nil
}

func extractExifFromPNG(data []byte) []byte {
	if len(data) < 8 {
		return nil
	}

	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(data[0:8], pngSig) {
		return nil
	}

	offset := 8
	for offset < len(data)-12 {
		if offset+8 > len(data) {
			return nil
		}

		length := int(data[offset])<<24 | int(data[offset+1])<<16 |
			int(data[offset+2])<<8 | int(data[offset+3])
		chunkType := data[offset+4 : offset+8]
		chunkDataStart := offset + 8
		chunkDataEnd := chunkDataStart + length

		if chunkDataEnd+4 > len(data) {
			return nil
		}

		if string(chunkType) == "eXIf" {
			chunk := make([]byte, 12+length)
			copy(chunk[0:4], data[offset:offset+4])
			copy(chunk[4:8], chunkType)
			copy(chunk[8:8+length], data[chunkDataStart:chunkDataEnd])
			copy(chunk[8+length:], data[chunkDataEnd:chunkDataEnd+4])
			return chunk
		}

		offset = chunkDataEnd + 4
	}

	return nil
}

func EmbedExifIntoImage(outputPath string, exifData *ExifData, outputFormat string) error {
	if exifData == nil || len(exifData.RawData) == 0 {
		return nil
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		return err
	}

	var newData []byte

	switch outputFormat {
	case "jpeg":
		newData = embedExifIntoJPEG(data, exifData.RawData)
	case "webp":
		newData = embedExifIntoWebP(data, exifData.RawData)
	case "png":
		newData = embedExifIntoPNG(data, exifData.RawData)
	default:
		return nil
	}

	if newData == nil {
		return nil
	}

	return os.WriteFile(outputPath, newData, 0644)
}

func embedExifIntoJPEG(data []byte, exifData []byte) []byte {
	if len(data) < 4 {
		return nil
	}

	if data[0] != 0xFF || data[1] != 0xD8 {
		return nil
	}

	var buffer bytes.Buffer

	buffer.WriteByte(0xFF)
	buffer.WriteByte(0xD8)

	if len(exifData) > 0 {
		buffer.Write(exifData)
	}

	offset := 2
	for offset < len(data)-1 {
		if data[offset] != 0xFF {
			return nil
		}

		marker := data[offset+1]

		if marker == 0xD9 {
			buffer.WriteByte(0xFF)
			buffer.WriteByte(0xD9)
			break
		}

		if offset+4 >= len(data) {
			return nil
		}

		segmentLen := int(data[offset+2])<<8 | int(data[offset+3])

		if marker == 0xE1 && len(exifData) > 0 {
			exifHeader := []byte("Exif\x00\x00")
			exifStart := offset + 4
			if exifStart+6 <= offset+segmentLen &&
				bytes.Equal(data[exifStart:exifStart+6], exifHeader) {
				offset += segmentLen
				continue
			}
		}

		if offset+segmentLen > len(data) {
			return nil
		}

		buffer.Write(data[offset : offset+segmentLen])
		offset += segmentLen
	}

	return buffer.Bytes()
}

func embedExifIntoWebP(data []byte, exifData []byte) []byte {
	if len(data) < 12 {
		return nil
	}

	if !bytes.Equal(data[0:4], []byte("RIFF")) {
		return nil
	}
	if !bytes.Equal(data[8:12], []byte("WEBP")) {
		return nil
	}

	var buffer bytes.Buffer

	buffer.Write(data[0:12])

	offset := 12
	hasExif := false

	for offset < len(data)-8 {
		chunkID := data[offset : offset+4]
		chunkSize := int(data[offset+4]) | int(data[offset+5])<<8 |
			int(data[offset+6])<<16 | int(data[offset+7])<<24
		chunkDataStart := offset + 8
		chunkDataEnd := chunkDataStart + chunkSize

		if chunkDataEnd > len(data) {
			return nil
		}

		if bytes.Equal(chunkID, []byte("EXIF")) {
			hasExif = true
			if len(exifData) > 0 {
				buffer.Write(exifData)
			}
		} else {
			buffer.Write(data[offset : chunkDataEnd+4])
		}

		offset = chunkDataEnd + 4
		if chunkSize%2 == 1 {
			offset++
		}
	}

	if !hasExif && len(exifData) > 0 {
		buffer.Write(exifData)
	}

	result := buffer.Bytes()
	fileSize := len(result) - 8
	result[4] = byte(fileSize & 0xFF)
	result[5] = byte((fileSize >> 8) & 0xFF)
	result[6] = byte((fileSize >> 16) & 0xFF)
	result[7] = byte((fileSize >> 24) & 0xFF)

	return result
}

func embedExifIntoPNG(data []byte, exifData []byte) []byte {
	if len(data) < 8 {
		return nil
	}

	pngSig := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	if !bytes.Equal(data[0:8], pngSig) {
		return nil
	}

	var buffer bytes.Buffer
	buffer.Write(data[0:8])

	offset := 8
	hasExif := false
	firstChunk := true

	for offset < len(data)-12 {
		if offset+8 > len(data) {
			return nil
		}

		length := int(data[offset])<<24 | int(data[offset+1])<<16 |
			int(data[offset+2])<<8 | int(data[offset+3])
		chunkType := data[offset+4 : offset+8]
		chunkDataStart := offset + 8
		chunkDataEnd := chunkDataStart + length

		if chunkDataEnd+4 > len(data) {
			return nil
		}

		chunkName := string(chunkType)

		if chunkName == "eXIf" {
			hasExif = true
			if len(exifData) > 0 {
				buffer.Write(exifData)
			}
		} else {
			if firstChunk && chunkName != "IHDR" {
				return nil
			}

			if !hasExif && chunkName == "IDAT" && len(exifData) > 0 {
				buffer.Write(exifData)
				hasExif = true
			}

			buffer.Write(data[offset : chunkDataEnd+4])
		}

		firstChunk = false
		offset = chunkDataEnd + 4
	}

	return buffer.Bytes()
}
