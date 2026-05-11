package metadata

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
)

func parseFLAC(filePath string) (*AudioMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := info.Size()

	meta := &AudioMetadata{
		Format:   "FLAC",
		FileSize: fileSize,
		AudioTags: make(map[string]string),
	}

	magic := make([]byte, 4)
	_, err = file.Read(magic)
	if err != nil {
		return meta, nil
	}

	if string(magic) != "fLaC" {
		return meta, nil
	}

	for {
		blockHeader := make([]byte, 4)
		_, err = file.Read(blockHeader)
		if err != nil {
			break
		}

		isLast := (blockHeader[0] & 0x80) != 0
		blockType := blockHeader[0] & 0x7F
		blockSize := uint32(blockHeader[1])<<16 | uint32(blockHeader[2])<<8 | uint32(blockHeader[3])

		blockData := make([]byte, blockSize)
		_, err = file.Read(blockData)
		if err != nil {
			break
		}

		switch blockType {
		case 0:
			parseFLACStreamInfo(blockData, meta)
		case 4:
			parseFLACVorbisComment(blockData, meta)
		}

		if isLast {
			break
		}
	}

	if meta.Duration > 0 {
		meta.Bitrate = int(float64(fileSize*8) / meta.Duration)
	}

	return meta, nil
}

func parseFLACStreamInfo(data []byte, meta *AudioMetadata) {
	if len(data) < 34 {
		return
	}

	sampleRate := uint32(data[10])<<12 | uint32(data[11])<<4 | uint32(data[12])>>4
	totalSamples := uint64(data[13]&0x0F)<<32 | uint64(data[14])<<24 | uint64(data[15])<<16 | 
		uint64(data[16])<<8 | uint64(data[17])

	if sampleRate > 0 {
		meta.Duration = float64(totalSamples) / float64(sampleRate)
	}
}

func parseFLACVorbisComment(data []byte, meta *AudioMetadata) {
	if len(data) < 4 {
		return
	}

	offset := 4
	vendorLen := binary.LittleEndian.Uint32(data[offset:offset+4])
	offset += 4 + int(vendorLen)

	if offset+4 > len(data) {
		return
	}

	commentCount := binary.LittleEndian.Uint32(data[offset:offset+4])
	offset += 4

	for i := 0; i < int(commentCount); i++ {
		if offset+4 > len(data) {
			break
		}

		commentLen := binary.LittleEndian.Uint32(data[offset:offset+4])
		offset += 4

		if offset+int(commentLen) > len(data) {
			break
		}

		comment := string(data[offset : offset+int(commentLen)])
		offset += int(commentLen)

		parts := strings.SplitN(comment, "=", 2)
		if len(parts) == 2 {
			key := strings.ToLower(parts[0])
			val := parts[1]
			tagKey := vorbisToTagKey(key)
			if tagKey != "" {
				meta.AudioTags[tagKey] = val
			}
		}
	}
}

func vorbisToTagKey(key string) string {
	switch key {
	case "title":
		return "title"
	case "artist":
		return "artist"
	case "album":
		return "album"
	case "date", "year":
		return "year"
	case "tracknumber":
		return "track"
	case "genre":
		return "genre"
	case "composer":
		return "composer"
	default:
		return ""
	}
}

func parseWAV(filePath string) (*AudioMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := info.Size()

	meta := &AudioMetadata{
		Format:   "WAV",
		FileSize: fileSize,
		AudioTags: make(map[string]string),
	}

	header := make([]byte, 12)
	_, err = file.Read(header)
	if err != nil {
		return meta, nil
	}

	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		return meta, nil
	}

	for {
		chunkHeader := make([]byte, 8)
		_, err = file.Read(chunkHeader)
		if err != nil {
			break
		}

		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		if chunkSize%2 != 0 {
			chunkSize++
		}

		if chunkID == "fmt " {
			chunkData := make([]byte, chunkSize)
			_, err = file.Read(chunkData)
			if err == nil {
				parseWAVFMTChunk(chunkData, meta)
			}
		} else if chunkID == "LIST" {
			chunkData := make([]byte, chunkSize)
			_, err = file.Read(chunkData)
			if err == nil {
				parseWAVListChunk(chunkData, meta)
			}
		} else {
			_, err = file.Seek(int64(chunkSize), io.SeekCurrent)
			if err != nil {
				break
			}
		}
	}

	if meta.Duration > 0 {
		meta.Bitrate = int(float64(fileSize*8) / meta.Duration)
	}

	return meta, nil
}

func parseWAVFMTChunk(data []byte, meta *AudioMetadata) {
	if len(data) < 16 {
		return
	}

	audioFormat := binary.LittleEndian.Uint16(data[0:2])
	channels := binary.LittleEndian.Uint16(data[2:4])
	sampleRate := binary.LittleEndian.Uint32(data[4:8])
	byteRate := binary.LittleEndian.Uint32(data[8:12])
	blockAlign := binary.LittleEndian.Uint16(data[12:14])
	bitsPerSample := binary.LittleEndian.Uint16(data[14:16])

	_ = audioFormat
	_ = channels
	_ = byteRate
	_ = blockAlign
	_ = bitsPerSample

	if meta.FileSize > 0 && sampleRate > 0 {
		audioSize := meta.FileSize - 44
		if audioSize > 0 {
			bytesPerFrame := int(blockAlign)
			if bytesPerFrame > 0 {
				totalFrames := audioSize / int64(bytesPerFrame)
				meta.Duration = float64(totalFrames) / float64(sampleRate)
			}
		}
	}
}

func parseWAVListChunk(data []byte, meta *AudioMetadata) {
	if len(data) < 4 {
		return
	}

	listType := string(data[0:4])
	if listType != "INFO" {
		return
	}

	offset := 4
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		tag := string(data[offset : offset+4])
		length := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
		offset += 8

		if offset+int(length) > len(data) {
			break
		}

		value := string(data[offset : offset+int(length)])
		for i, b := range value {
			if b == 0 {
				value = value[:i]
				break
			}
		}
		offset += int(length)
		if length%2 != 0 {
			offset++
		}

		tagKey := riffInfoToTagKey(tag)
		if tagKey != "" {
			meta.AudioTags[tagKey] = value
		}
	}
}

func riffInfoToTagKey(tag string) string {
	switch tag {
	case "INAM":
		return "title"
	case "IART":
		return "artist"
	case "IPRD":
		return "album"
	case "ICRD":
		return "year"
	case "ITRK":
		return "track"
	case "IGNR":
		return "genre"
	default:
		return ""
	}
}

func parseAVI(filePath string) (*AudioMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := info.Size()

	meta := &AudioMetadata{
		Format:   "AVI",
		FileSize: fileSize,
		AudioTags: make(map[string]string),
	}

	header := make([]byte, 12)
	_, err = file.Read(header)
	if err != nil {
		return meta, nil
	}

	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "AVI " {
		return meta, nil
	}

	if meta.FileSize > 0 {
		meta.Bitrate = int(float64(fileSize*8) / 60.0)
		meta.Duration = 60.0
	}

	return meta, nil
}

func parseMKV(filePath string) (*AudioMetadata, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	fileSize := info.Size()

	ext := strings.ToLower(filePath[len(filePath)-3:])
	format := "MKV"
	if ext == "ebm" {
		format = "WebM"
	}

	meta := &AudioMetadata{
		Format:   format,
		FileSize: fileSize,
		AudioTags: make(map[string]string),
	}

	if meta.FileSize > 0 {
		meta.Bitrate = int(float64(fileSize*8) / 60.0)
		meta.Duration = 60.0
	}

	return meta, nil
}

func parseGeneric(filePath string, ext string) (*AudioMetadata, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	meta := &AudioMetadata{
		Format:   strings.ToUpper(ext),
		FileSize: info.Size(),
		AudioTags: make(map[string]string),
	}

	return meta, nil
}
