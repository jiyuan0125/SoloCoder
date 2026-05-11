package metadata

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf16"
)

type ID3v2Header struct {
	Version      uint8
	MinorVersion uint8
	Flags        uint8
	Size         uint32
}

type ID3v2Frame struct {
	ID          string
	Data        []byte
	TextEncoding uint8
}

var (
	bitrateIndexV1 = []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 0}
	bitrateIndexV2 = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, 0}
	sampleRateV1 = []int{44100, 48000, 32000, 0}
	sampleRateV2 = []int{22050, 24000, 16000, 0}
	sampleRateV25 = []int{11025, 12000, 8000, 0}
)

func parseMP3(filePath string) (*AudioMetadata, error) {
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
		Format:   "MP3",
		FileSize: fileSize,
	}

	tags := make(map[string]string)
	var dataOffset int64 = 0
	var totalFrames int64 = 0
	var totalDuration float64 = 0

	header := make([]byte, 10)
	_, err = file.Read(header)
	if err != nil {
		return meta, nil
	}

	if bytes.HasPrefix(header, []byte("ID3")) {
		id3Header, _, err := parseID3v2Header(header)
		if err == nil {
			dataOffset = int64(10 + id3Header.Size)
			frameData := make([]byte, id3Header.Size)
			_, err = file.Read(frameData)
			if err == nil {
				parsedTags, err := parseID3v2Frames(id3Header, frameData)
				if err == nil {
					for k, v := range parsedTags {
						tags[k] = v
					}
				}
			}
		}
	} else {
		file.Seek(0, io.SeekStart)
	}

	_, err = file.Seek(dataOffset, io.SeekStart)
	if err != nil {
		return nil, fmt.Errorf("定位文件失败: %v", err)
	}

	frameCount := 0
	maxFrames := 1000
	
	for frameCount < maxFrames {
		frameHeader := make([]byte, 4)
		_, err := file.Read(frameHeader)
		if err != nil {
			break
		}

		if !isValidMP3FrameHeader(frameHeader) {
			file.Seek(-3, io.SeekCurrent)
			continue
		}

		frameInfo, err := parseMP3FrameHeader(frameHeader)
		if err != nil {
			file.Seek(-3, io.SeekCurrent)
			continue
		}

		if frameInfo.SampleRate > 0 {
			frameDuration := float64(frameInfo.Samples) / float64(frameInfo.SampleRate)
			totalDuration += frameDuration
			totalFrames++
		}

		if frameInfo.FrameSize > 0 {
			file.Seek(int64(frameInfo.FrameSize-4), io.SeekCurrent)
		} else {
			break
		}
		frameCount++
	}

	if totalDuration > 0 {
		meta.Duration = totalDuration
		audioSize := fileSize - dataOffset
		if audioSize > 0 && totalDuration > 0 {
			meta.Bitrate = int(float64(audioSize*8) / totalDuration)
		}
	}

	for k, v := range tags {
		if meta.AudioTags == nil {
			meta.AudioTags = make(map[string]string)
		}
		meta.AudioTags[k] = v
	}

	return meta, nil
}

func parseID3v2Header(header []byte) (*ID3v2Header, uint32, error) {
	if len(header) < 10 {
		return nil, 0, fmt.Errorf("ID3头太短")
	}

	if header[0] != 'I' || header[1] != 'D' || header[2] != '3' {
		return nil, 0, fmt.Errorf("不是ID3v2标签")
	}

	version := header[3]
	if version < 2 || version > 4 {
		return nil, 0, fmt.Errorf("不支持的ID3v2版本: %d", version)
	}

	minorVersion := header[4]
	flags := header[5]

	size := uint32(header[6]&0x7f)<<21 |
		uint32(header[7]&0x7f)<<14 |
		uint32(header[8]&0x7f)<<7 |
		uint32(header[9]&0x7f)

	return &ID3v2Header{
		Version:      version,
		MinorVersion: minorVersion,
		Flags:        flags,
		Size:         size,
	}, 10, nil
}

func parseID3v2Frames(header *ID3v2Header, data []byte) (map[string]string, error) {
	tags := make(map[string]string)
	offset := 0
	dataLen := len(data)

	for offset < dataLen-10 {
		var frameID string
		var frameSize uint32
		var frameFlags uint16

		if header.Version == 2 {
			if offset+6 > dataLen {
				break
			}
			frameID = string(data[offset : offset+3])
			frameSize = uint32(data[offset+3])<<16 |
				uint32(data[offset+4])<<8 |
				uint32(data[offset+5])
			offset += 6
		} else if header.Version == 3 {
			if offset+10 > dataLen {
				break
			}
			frameID = string(data[offset : offset+4])
			frameSize = binary.BigEndian.Uint32(data[offset+4 : offset+8])
			frameFlags = binary.BigEndian.Uint16(data[offset+8 : offset+10])
			offset += 10
		} else {
			if offset+10 > dataLen {
				break
			}
			frameID = string(data[offset : offset+4])
			frameSize = uint32(data[offset+4]&0x7f)<<21 |
				uint32(data[offset+5]&0x7f)<<14 |
				uint32(data[offset+6]&0x7f)<<7 |
				uint32(data[offset+7]&0x7f)
			frameFlags = binary.BigEndian.Uint16(data[offset+8 : offset+10])
			offset += 10
		}

		if frameSize == 0 || frameID[0] == 0 {
			break
		}

		if int(offset)+int(frameSize) > dataLen {
			break
		}

		frameData := data[offset : offset+int(frameSize)]
		offset += int(frameSize)

		tagKey, tagValue, err := parseID3v2FrameData(frameID, frameData)
		if err == nil && tagKey != "" {
			tags[tagKey] = tagValue
		}
		_ = frameFlags
	}

	return tags, nil
}

func parseID3v2FrameData(frameID string, frameData []byte) (string, string, error) {
	if len(frameData) == 0 {
		return "", "", fmt.Errorf("空帧数据")
	}

	textEncoding := frameData[0]
	var text string
	var err error

	switch textEncoding {
	case 0:
		text, err = decodeISO88591(frameData[1:])
	case 1:
		text, err = decodeUTF16WithBOM(frameData[1:])
	case 2:
		text, err = decodeUTF16BE(frameData[1:])
	case 3:
		text, err = decodeUTF8(frameData[1:])
	default:
		return "", "", fmt.Errorf("未知的文本编码: %d", textEncoding)
	}

	if err != nil {
		return "", "", err
	}

	tagKey := id3FrameToTagKey(frameID)
	if tagKey == "" {
		return "", "", nil
	}

	return tagKey, strings.TrimSpace(text), nil
}

func id3FrameToTagKey(frameID string) string {
	switch frameID {
	case "TIT2", "TT2":
		return "title"
	case "TPE1", "TP1":
		return "artist"
	case "TALB", "TAL":
		return "album"
	case "TYER", "TYE":
		return "year"
	case "TRCK", "TRK":
		return "track"
	case "TCON", "TCO":
		return "genre"
	case "TPUB", "TPB":
		return "publisher"
	case "COMM", "COM":
		return "comment"
	default:
		return ""
	}
}

func decodeISO88591(data []byte) (string, error) {
	runes := make([]rune, len(data))
	for i, b := range data {
		runes[i] = rune(b)
	}
	return string(runes), nil
}

func decodeUTF16WithBOM(data []byte) (string, error) {
	if len(data) < 2 {
		return "", nil
	}

	isBE := true
	if data[0] == 0xFF && data[1] == 0xFE {
		isBE = false
		data = data[2:]
	} else if data[0] == 0xFE && data[1] == 0xFF {
		data = data[2:]
	}

	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	uint16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		var u uint16
		if isBE {
			u = uint16(data[i])<<8 | uint16(data[i+1])
		} else {
			u = uint16(data[i+1])<<8 | uint16(data[i])
		}
		if u == 0 {
			break
		}
		uint16s = append(uint16s, u)
	}

	return string(utf16.Decode(uint16s)), nil
}

func decodeUTF16BE(data []byte) (string, error) {
	if len(data)%2 != 0 {
		data = data[:len(data)-1]
	}

	uint16s := make([]uint16, 0, len(data)/2)
	for i := 0; i < len(data); i += 2 {
		u := uint16(data[i])<<8 | uint16(data[i+1])
		if u == 0 {
			break
		}
		uint16s = append(uint16s, u)
	}

	return string(utf16.Decode(uint16s)), nil
}

func decodeUTF8(data []byte) (string, error) {
	for i, b := range data {
		if b == 0 {
			data = data[:i]
			break
		}
	}
	return string(data), nil
}

func isValidMP3FrameHeader(header []byte) bool {
	if len(header) < 4 {
		return false
	}

	if header[0] != 0xFF || (header[1]&0xE0) != 0xE0 {
		return false
	}

	layer := (header[1] & 0x06) >> 1
	if layer == 0 {
		return false
	}

	bitrateIdx := (header[2] & 0xF0) >> 4
	if bitrateIdx == 0 || bitrateIdx == 15 {
		return false
	}

	sampleRateIdx := (header[2] & 0x0C) >> 2
	if sampleRateIdx == 3 {
		return false
	}

	return true
}

type MP3FrameInfo struct {
	Version     int
	Layer       int
	SampleRate  int
	Bitrate     int
	FrameSize   int
	Samples     int
	Channels    int
}

func parseMP3FrameHeader(header []byte) (*MP3FrameInfo, error) {
	if !isValidMP3FrameHeader(header) {
		return nil, fmt.Errorf("无效的MP3帧头")
	}

	versionIdx := (header[1] & 0x18) >> 3
	layerIdx := (header[1] & 0x06) >> 1
	protection := header[1] & 0x01
	bitrateIdx := (header[2] & 0xF0) >> 4
	sampleRateIdx := (header[2] & 0x0C) >> 2
	padding := (header[2] & 0x02) >> 1
	channelIdx := (header[3] & 0xC0) >> 6

	_ = protection

	info := &MP3FrameInfo{}

	switch versionIdx {
	case 0:
		info.Version = 25
	case 2:
		info.Version = 2
	case 3:
		info.Version = 1
	default:
		info.Version = 2
	}

	info.Layer = 4 - int(layerIdx)

	switch info.Version {
	case 1:
		info.SampleRate = sampleRateV1[sampleRateIdx]
		if info.Layer == 1 {
			info.Bitrate = bitrateIndexV1[bitrateIdx] * 1000
			info.Samples = 384
		} else {
			info.Bitrate = bitrateIndexV1[bitrateIdx] * 1000
			info.Samples = 1152
		}
	case 2, 25:
		info.SampleRate = sampleRateV2[sampleRateIdx]
		if info.Version == 25 {
			info.SampleRate = sampleRateV25[sampleRateIdx]
		}
		if info.Layer == 1 {
			info.Bitrate = bitrateIndexV2[bitrateIdx] * 1000
			info.Samples = 384
		} else {
			info.Bitrate = bitrateIndexV2[bitrateIdx] * 1000
			info.Samples = 576
		}
	}

	channels := 2
	if channelIdx == 3 {
		channels = 1
	}
	info.Channels = channels

	if info.Bitrate > 0 && info.SampleRate > 0 {
		if info.Layer == 1 {
			info.FrameSize = (12*info.Bitrate/info.SampleRate + int(padding)) * 4
		} else {
			info.FrameSize = 144*info.Bitrate/info.SampleRate + int(padding)
		}
	}

	return info, nil
}
