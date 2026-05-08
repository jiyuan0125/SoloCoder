package exif

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"
)

type ByteOrder interface {
	binary.ByteOrder
}

type EXIFData struct {
	Tags      map[uint16]interface{}
	ExifTags  map[uint16]interface{}
	GPSTags   map[uint16]interface{}
	Thumbnail []byte
}

type Rational struct {
	Numerator   uint32
	Denominator uint32
}

type SRational struct {
	Numerator   int32
	Denominator int32
}

const (
	TagDateTimeOriginal = 0x9003
	TagDateTime         = 0x0132
	TagFNumber          = 0x829D
	TagExposureTime     = 0x829A
	TagISOSpeedRatings  = 0x8827
	TagExifIFDPointer   = 0x8769
	TagGPSIFDPointer    = 0x8825
	TagJPEGInterchangeFormat = 0x0201
	TagJPEGInterchangeFormatLength = 0x0202
)

const (
	GPSTagLatitudeRef     = 0x0001
	GPSTagLatitude        = 0x0002
	GPSTagLongitudeRef    = 0x0003
	GPSTagLongitude       = 0x0004
	GPSTagAltitudeRef     = 0x0005
	GPSTagAltitude        = 0x0006
	GPSTagTimestamp       = 0x0007
	GPSTagDateStamp       = 0x001D
)

const (
	TypeByte       = 1
	TypeASCII      = 2
	TypeShort      = 3
	TypeLong       = 4
	TypeRational   = 5
	TypeSByte      = 6
	TypeUndefined  = 7
	TypeSShort     = 8
	TypeSLong      = 9
	TypeSRational  = 10
	TypeFloat      = 11
	TypeDouble     = 12
)

var ErrInvalidFormat = errors.New("invalid image format")
var ErrNoEXIF = errors.New("no EXIF data found")

func Parse(reader io.ReadSeeker) (*EXIFData, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return nil, err
	}

	if header[0] == 0xFF && header[1] == 0xD8 {
		return parseJPEG(reader)
	}

	if (header[0] == 'I' && header[1] == 'I' && header[2] == 0x2A && header[3] == 0x00) ||
		(header[0] == 'M' && header[1] == 'M' && header[2] == 0x00 && header[3] == 0x2A) {
		_, err := reader.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return parseTIFF(reader)
	}

	return nil, ErrInvalidFormat
}

func parseJPEG(reader io.ReadSeeker) (*EXIFData, error) {
	for {
		marker := make([]byte, 2)
		_, err := io.ReadFull(reader, marker)
		if err != nil {
			if err == io.EOF {
				return nil, ErrNoEXIF
			}
			return nil, err
		}

		if marker[0] != 0xFF {
			return nil, ErrInvalidFormat
		}

		if marker[1] == 0xDA {
			return nil, ErrNoEXIF
		}

		lengthBuf := make([]byte, 2)
		_, err = io.ReadFull(reader, lengthBuf)
		if err != nil {
			return nil, err
		}
		length := binary.BigEndian.Uint16(lengthBuf)

		if marker[1] == 0xE1 {
			app1Data := make([]byte, length-2)
			_, err := io.ReadFull(reader, app1Data)
			if err != nil {
				return nil, err
			}

			if len(app1Data) >= 6 && string(app1Data[0:6]) == "Exif\000\000" {
				tiffData := app1Data[6:]
				return parseTIFFBytes(tiffData)
			}
			continue
		}

		_, err = reader.Seek(int64(length-2), io.SeekCurrent)
		if err != nil {
			return nil, err
		}
	}
}

func parseTIFF(reader io.ReadSeeker) (*EXIFData, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return parseTIFFBytes(data)
}

func parseTIFFBytes(data []byte) (*EXIFData, error) {
	if len(data) < 8 {
		return nil, ErrInvalidFormat
	}

	var byteOrder binary.ByteOrder
	if data[0] == 'I' && data[1] == 'I' {
		byteOrder = binary.LittleEndian
	} else if data[0] == 'M' && data[1] == 'M' {
		byteOrder = binary.BigEndian
	} else {
		return nil, ErrInvalidFormat
	}

	magic := byteOrder.Uint16(data[2:4])
	if magic != 0x002A {
		return nil, ErrInvalidFormat
	}

	ifd0Offset := byteOrder.Uint32(data[4:8])
	if int(ifd0Offset) >= len(data) {
		return nil, ErrInvalidFormat
	}

	result := &EXIFData{
		Tags:     make(map[uint16]interface{}),
		ExifTags: make(map[uint16]interface{}),
		GPSTags:  make(map[uint16]interface{}),
	}

	ifd0Tags, err := parseIFD(data, ifd0Offset, byteOrder)
	if err != nil {
		return nil, err
	}

	for k, v := range ifd0Tags {
		result.Tags[k] = v
	}

	if exifPtr, ok := ifd0Tags[TagExifIFDPointer]; ok {
		if exifOffset, ok := exifPtr.(uint32); ok {
			exifTags, err := parseIFD(data, exifOffset, byteOrder)
			if err == nil {
				for k, v := range exifTags {
					result.ExifTags[k] = v
				}
			}
		}
	}

	if gpsPtr, ok := ifd0Tags[TagGPSIFDPointer]; ok {
		if gpsOffset, ok := gpsPtr.(uint32); ok {
			gpsTags, err := parseIFD(data, gpsOffset, byteOrder)
			if err == nil {
				for k, v := range gpsTags {
					result.GPSTags[k] = v
				}
			}
		}
	}

	result.Thumbnail = extractThumbnail(data, ifd0Tags, byteOrder)

	return result, nil
}

func parseIFD(data []byte, offset uint32, byteOrder binary.ByteOrder) (map[uint16]interface{}, error) {
	if int(offset) >= len(data) {
		return nil, ErrInvalidFormat
	}

	if int(offset+2) > len(data) {
		return nil, ErrInvalidFormat
	}

	count := byteOrder.Uint16(data[offset : offset+2])
	tags := make(map[uint16]interface{})

	for i := 0; i < int(count); i++ {
		entryOffset := offset + 2 + uint32(i*12)
		if int(entryOffset+12) > len(data) {
			break
		}

		tagID := byteOrder.Uint16(data[entryOffset : entryOffset+2])
		dataType := byteOrder.Uint16(data[entryOffset+2 : entryOffset+4])
		dataCount := byteOrder.Uint32(data[entryOffset+4 : entryOffset+8])
		valueOffset := data[entryOffset+8 : entryOffset+12]

		value, err := parseTagValue(data, dataType, dataCount, valueOffset, byteOrder)
		if err == nil {
			tags[tagID] = value
		}
	}

	return tags, nil
}

func parseTagValue(data []byte, dataType uint16, count uint32, valueData []byte, byteOrder binary.ByteOrder) (interface{}, error) {
	dataSize := getTypeSize(dataType)
	totalSize := dataSize * int(count)

	var rawData []byte
	if totalSize <= 4 {
		rawData = valueData[:totalSize]
	} else {
		offset := byteOrder.Uint32(valueData)
		if int(offset)+totalSize > len(data) {
			return nil, ErrInvalidFormat
		}
		rawData = data[offset : offset+uint32(totalSize)]
	}

	return decodeValue(rawData, dataType, count, byteOrder)
}

func getTypeSize(dataType uint16) int {
	switch dataType {
	case TypeByte, TypeSByte, TypeASCII, TypeUndefined:
		return 1
	case TypeShort, TypeSShort:
		return 2
	case TypeLong, TypeSLong, TypeFloat:
		return 4
	case TypeRational, TypeSRational, TypeDouble:
		return 8
	default:
		return 1
	}
}

func decodeValue(data []byte, dataType uint16, count uint32, byteOrder binary.ByteOrder) (interface{}, error) {
	switch dataType {
	case TypeByte:
		if count == 1 {
			return data[0], nil
		}
		result := make([]byte, count)
		copy(result, data)
		return result, nil

	case TypeASCII:
		if count == 0 {
			return "", nil
		}
		end := count - 1
		if end > uint32(len(data)) {
			end = uint32(len(data))
		}
		return strings.TrimRight(string(data[:end]), "\x00"), nil

	case TypeShort:
		if count == 1 {
			return byteOrder.Uint16(data[0:2]), nil
		}
		result := make([]uint16, count)
		for i := uint32(0); i < count; i++ {
			result[i] = byteOrder.Uint16(data[i*2 : (i+1)*2])
		}
		return result, nil

	case TypeLong:
		if count == 1 {
			return byteOrder.Uint32(data[0:4]), nil
		}
		result := make([]uint32, count)
		for i := uint32(0); i < count; i++ {
			result[i] = byteOrder.Uint32(data[i*4 : (i+1)*4])
		}
		return result, nil

	case TypeRational:
		if count == 1 {
			return Rational{
				Numerator:   byteOrder.Uint32(data[0:4]),
				Denominator: byteOrder.Uint32(data[4:8]),
			}, nil
		}
		result := make([]Rational, count)
		for i := uint32(0); i < count; i++ {
			offset := i * 8
			result[i] = Rational{
				Numerator:   byteOrder.Uint32(data[offset : offset+4]),
				Denominator: byteOrder.Uint32(data[offset+4 : offset+8]),
			}
		}
		return result, nil

	case TypeUndefined:
		result := make([]byte, count)
		copy(result, data)
		return result, nil

	default:
		return nil, errors.New("unsupported data type")
	}
}

func extractThumbnail(data []byte, ifdTags map[uint16]interface{}, byteOrder binary.ByteOrder) []byte {
	offsetVal, ok := ifdTags[TagJPEGInterchangeFormat]
	if !ok {
		return nil
	}
	lengthVal, ok := ifdTags[TagJPEGInterchangeFormatLength]
	if !ok {
		return nil
	}

	offset, ok := offsetVal.(uint32)
	if !ok {
		return nil
	}
	length, ok := lengthVal.(uint32)
	if !ok {
		return nil
	}

	if int(offset)+int(length) > len(data) {
		return nil
	}

	result := make([]byte, length)
	copy(result, data[offset:offset+length])
	return result
}

func (r *Rational) Float64() float64 {
	if r.Denominator == 0 {
		return 0
	}
	return float64(r.Numerator) / float64(r.Denominator)
}

func (r *SRational) Float64() float64 {
	if r.Denominator == 0 {
		return 0
	}
	return float64(r.Numerator) / float64(r.Denominator)
}

func (e *EXIFData) GetDateTime() (time.Time, bool) {
	var dateStr string
	if v, ok := e.ExifTags[TagDateTimeOriginal]; ok {
		dateStr = v.(string)
	} else if v, ok := e.Tags[TagDateTime]; ok {
		dateStr = v.(string)
	} else {
		return time.Time{}, false
	}

	t, err := time.Parse("2006:01:02 15:04:05", dateStr)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (e *EXIFData) GetExposureTime() string {
	if v, ok := e.ExifTags[TagExposureTime]; ok {
		if rat, ok := v.(Rational); ok {
			val := rat.Float64()
			if val < 1 {
				return fmt.Sprintf("1/%d", int(1/val+0.5))
			}
			return fmt.Sprintf("%.1f", val)
		}
	}
	return ""
}

func (e *EXIFData) GetFNumber() string {
	if v, ok := e.ExifTags[TagFNumber]; ok {
		if rat, ok := v.(Rational); ok {
			return fmt.Sprintf("f/%.1f", rat.Float64())
		}
	}
	return ""
}

func (e *EXIFData) GetISO() int {
	if v, ok := e.ExifTags[TagISOSpeedRatings]; ok {
		switch val := v.(type) {
		case uint16:
			return int(val)
		case []uint16:
			if len(val) > 0 {
				return int(val[0])
			}
		}
	}
	return 0
}

func (e *EXIFData) GetGPSLatitude() (float64, bool) {
	latRef, refOk := e.GPSTags[GPSTagLatitudeRef]
	latVal, valOk := e.GPSTags[GPSTagLatitude]
	if !refOk || !valOk {
		return 0, false
	}

	coords, ok := latVal.([]Rational)
	if !ok || len(coords) != 3 {
		return 0, false
	}

	degrees := coords[0].Float64()
	minutes := coords[1].Float64()
	seconds := coords[2].Float64()

	lat := degrees + minutes/60 + seconds/3600

	refStr, ok := latRef.(string)
	if ok && (refStr == "S" || refStr == "s") {
		lat = -lat
	}

	return lat, true
}

func (e *EXIFData) GetGPSLongitude() (float64, bool) {
	lonRef, refOk := e.GPSTags[GPSTagLongitudeRef]
	lonVal, valOk := e.GPSTags[GPSTagLongitude]
	if !refOk || !valOk {
		return 0, false
	}

	coords, ok := lonVal.([]Rational)
	if !ok || len(coords) != 3 {
		return 0, false
	}

	degrees := coords[0].Float64()
	minutes := coords[1].Float64()
	seconds := coords[2].Float64()

	lon := degrees + minutes/60 + seconds/3600

	refStr, ok := lonRef.(string)
	if ok && (refStr == "W" || refStr == "w") {
		lon = -lon
	}

	return lon, true
}

func (e *EXIFData) GetGPSAltitude() (float64, bool) {
	altRef, refOk := e.GPSTags[GPSTagAltitudeRef]
	altVal, valOk := e.GPSTags[GPSTagAltitude]
	if !valOk {
		return 0, false
	}

	rat, ok := altVal.(Rational)
	if !ok {
		return 0, false
	}

	alt := rat.Float64()

	if refOk {
		if refByte, ok := altRef.(byte); ok && refByte == 1 {
			alt = -alt
		}
	}

	return alt, true
}

func (e *EXIFData) FormatAll() map[string]interface{} {
	result := make(map[string]interface{})

	if dt, ok := e.GetDateTime(); ok {
		result["DateTime"] = dt.Format(time.RFC3339)
	}

	if expTime := e.GetExposureTime(); expTime != "" {
		result["ExposureTime"] = expTime
	}

	if fNum := e.GetFNumber(); fNum != "" {
		result["FNumber"] = fNum
	}

	if iso := e.GetISO(); iso > 0 {
		result["ISO"] = iso
	}

	if lat, ok := e.GetGPSLatitude(); ok {
		result["GPSLatitude"] = lat
	}

	if lon, ok := e.GetGPSLongitude(); ok {
		result["GPSLongitude"] = lon
	}

	if alt, ok := e.GetGPSAltitude(); ok {
		result["GPSAltitude"] = alt
	}

	return result
}

func (e *EXIFData) RoundFloat(f float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(f*pow) / pow
}
