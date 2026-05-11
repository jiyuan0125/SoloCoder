package metadata

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type MP4Atom struct {
	Size   uint64
	Type   [4]byte
	Data   []byte
	Offset int64
}

type MP4Box struct {
	Type   string
	Offset int64
	Size   uint64
	Data   []byte
}

func parseMP4(filePath string) (*AudioMetadata, error) {
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
		Format:   "MP4",
		FileSize: fileSize,
		AudioTags: make(map[string]string),
	}

	atoms, err := parseMP4Atoms(file)
	if err != nil {
		return meta, nil
	}

	if moov, ok := atoms["moov"]; ok {
		extractMP4Metadata(moov, meta)
		extractMP4Duration(moov, meta)
	}

	extractMP4Bitrate(meta)

	return meta, nil
}

func parseMP4Atoms(file *os.File) (map[string]*MP4Box, error) {
	atoms := make(map[string]*MP4Box)

	info, err := file.Stat()
	if err != nil {
		return atoms, err
	}
	fileSize := info.Size()

	offset := int64(0)
	for offset < fileSize {
		box, err := readMP4Box(file, offset)
		if err != nil {
			break
		}
		if box == nil {
			break
		}

		atoms[box.Type] = box
		offset = box.Offset + int64(box.Size)
	}

	return atoms, nil
}

func readMP4Box(file *os.File, offset int64) (*MP4Box, error) {
	_, err := file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	header := make([]byte, 8)
	_, err = file.Read(header)
	if err != nil {
		return nil, err
	}

	size := uint64(binary.BigEndian.Uint32(header[0:4]))
	boxType := string(header[4:8])

	if size == 1 {
		extSize := make([]byte, 8)
		_, err = file.Read(extSize)
		if err != nil {
			return nil, err
		}
		size = binary.BigEndian.Uint64(extSize)
		_ = int64(16)
	}

	if size < 8 {
		return nil, fmt.Errorf("无效的MP4 box 大小")
	}

	dataSize := int64(size) - 8
	if dataSize <= 0 {
		dataSize = 0
	}

	data := make([]byte, dataSize)
	if dataSize > 0 {
		_, err = file.Read(data)
	}

	return &MP4Box{
		Type:   boxType,
		Offset: offset,
		Size:   size,
		Data:   data,
	}, nil
}

func parseNestedBoxes(data []byte) ([]*MP4Box, error) {
	var boxes []*MP4Box
	offset := 0
	dataLen := len(data)

	for offset < dataLen {
		if offset+8 > dataLen {
			break
		}

		size := uint64(binary.BigEndian.Uint32(data[offset:offset+4]))
		boxType := string(data[offset+4 : offset+8])

		if size == 1 {
			if offset+16 > dataLen {
				break
			}
			size = binary.BigEndian.Uint64(data[offset+8 : offset+16])
			offset += 16
		} else {
			offset += 8
		}

		if size == 0 {
			break
		}

		dataSize := int(size) - 8
		if dataSize < 0 {
			dataSize = 0
		}

		if offset+dataSize > dataLen {
			dataSize = dataLen - offset
		}

		boxData := make([]byte, dataSize)
		if dataSize > 0 {
			copy(boxData, data[offset:offset+dataSize])
		}

		boxes = append(boxes, &MP4Box{
			Type:   boxType,
			Offset: int64(offset),
			Size:   size,
			Data:   boxData,
		})

		offset += dataSize
	}

	return boxes, nil
}

func extractMP4Metadata(moov *MP4Box, meta *AudioMetadata) {
	moovBoxes, err := parseNestedBoxes(moov.Data)
	if err != nil {
		return
	}

	for _, box := range moovBoxes {
		if box.Type == "udta" {
			extractMP4UDTA(box, meta)
		}
	}
}

func extractMP4UDTA(udta *MP4Box, meta *AudioMetadata) {
	udtaBoxes, err := parseNestedBoxes(udta.Data)
	if err != nil {
		return
	}

	for _, box := range udtaBoxes {
		if box.Type == "meta" {
			extractMP4Meta(box, meta)
		}
	}
}

func extractMP4Meta(metaBox *MP4Box, meta *AudioMetadata) {
	metaBoxes, err := parseNestedBoxes(metaBox.Data)
	if err != nil {
		return
	}

	for _, box := range metaBoxes {
		if box.Type == "ilst" {
			extractMP4ILST(box, meta)
		}
	}
}

func extractMP4ILST(ilst *MP4Box, meta *AudioMetadata) {
	ilstBoxes, err := parseNestedBoxes(ilst.Data)
	if err != nil {
		return
	}

	for _, box := range ilstBoxes {
		extractMP4TagValue(box.Type, box, meta)
	}
}

func extractMP4TagValue(atomName string, atomBox *MP4Box, meta *AudioMetadata) {
	atomBoxes, err := parseNestedBoxes(atomBox.Data)
	if err != nil {
		return
	}

	for _, box := range atomBoxes {
		if box.Type == "data" && len(box.Data) >= 8 {
			dataType := binary.BigEndian.Uint32(box.Data[0:4])
			data := box.Data[8:]

			switch dataType {
			case 1:
				tagName := iTunestoTagName(atomName)
				if tagName != "" {
					meta.AudioTags[tagName] = string(data)
				}
			case 21:
				tagName := iTunestoTagName(atomName)
				if tagName != "" {
					meta.AudioTags[tagName] = string(data)
				}
			case 0:
				if len(data) >= 4 {
					tagName := iTunestoTagName(atomName)
					if tagName != "" {
						value := fmt.Sprintf("%d", binary.BigEndian.Uint32(data[0:4]))
						meta.AudioTags[tagName] = value
					}
				}
			}
		}
	}
}

func iTunestoTagName(name string) string {
	switch name {
	case "name":
		return "title"
	case "artist":
		return "artist"
	case "album":
		return "album"
	case "date":
		return "year"
	case "trkn":
		return "track"
	case "genre":
		return "genre"
	case "wrt":
		return "composer"
	case "day":
		return "year"
	default:
		return ""
	}
}

func extractMP4Duration(moov *MP4Box, meta *AudioMetadata) {
	moovBoxes, err := parseNestedBoxes(moov.Data)
	if err != nil {
		return
	}

	for _, box := range moovBoxes {
		if box.Type == "mvhd" {
			parseMP4MVHD(box, meta)
			break
		}
	}
}

func parseMP4MVHD(mvhd *MP4Box, meta *AudioMetadata) {
	if len(mvhd.Data) < 20 {
		return
	}

	version := mvhd.Data[0]
	var timescale uint32
	var duration uint64

	if version == 0 {
		if len(mvhd.Data) < 20 {
			return
		}
		timescale = binary.BigEndian.Uint32(mvhd.Data[12:16])
		duration = uint64(binary.BigEndian.Uint32(mvhd.Data[16:20]))
	} else {
		if len(mvhd.Data) < 28 {
			return
		}
		timescale = binary.BigEndian.Uint32(mvhd.Data[20:24])
		duration = binary.BigEndian.Uint64(mvhd.Data[24:32])
	}

	if timescale > 0 {
		meta.Duration = float64(duration) / float64(timescale)
	}
}

func extractMP4Bitrate(meta *AudioMetadata) {
	if meta.Duration > 0 {
		meta.Bitrate = int(float64(meta.FileSize*8) / meta.Duration)
	}
}
