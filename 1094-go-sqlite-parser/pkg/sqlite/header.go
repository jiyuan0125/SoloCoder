package sqlite

import (
	"encoding/binary"
	"errors"
)

const magicString = "SQLite format 3\x00"

func parseHeader(data []byte) (*Header, error) {
	if len(data) < 100 {
		return nil, errors.New("header too short")
	}

	magic := string(data[0:16])
	if magic != magicString {
		return nil, errors.New("invalid SQLite magic string")
	}

	pageSizeVal := binary.BigEndian.Uint16(data[16:18])
	actualPageSize := uint32(pageSizeVal)
	if pageSizeVal == 1 {
		actualPageSize = 65536
	}

	if !isValidPageSize(actualPageSize) {
		return nil, errors.New("invalid page size")
	}

	header := &Header{
		MagicString:        magic,
		PageSize:           pageSizeVal,
		FileFormatWrite:    data[18],
		FileFormatRead:     data[19],
		ReservedSpace:      data[20],
		MaxEmbeddedPayload: data[21],
		MinEmbeddedPayload: data[22],
		LeafPayloadFraction: data[23],
		FileChangeCounter:  binary.BigEndian.Uint32(data[24:28]),
		PageCount:          binary.BigEndian.Uint32(data[28:32]),
		FirstFreelistPage:  binary.BigEndian.Uint32(data[32:36]),
		FreelistPageCount:  binary.BigEndian.Uint32(data[36:40]),
		SchemaCookie:       binary.BigEndian.Uint32(data[40:44]),
		SchemaFormat:       binary.BigEndian.Uint32(data[44:48]),
		DefaultPageCache:   binary.BigEndian.Uint32(data[48:52]),
		AutoVacuumTop:      binary.BigEndian.Uint32(data[52:56]),
		IncrementalVacuum:  binary.BigEndian.Uint32(data[56:60]),
		TextEncoding:       binary.BigEndian.Uint32(data[60:64]),
		UserVersion:        binary.BigEndian.Uint32(data[64:68]),
		ApplicationID:      binary.BigEndian.Uint32(data[68:72]),
		VersionValidFor:    binary.BigEndian.Uint32(data[92:96]),
		SQLiteVersion:      binary.BigEndian.Uint32(data[96:100]),
	}

	return header, nil
}

func isValidPageSize(size uint32) bool {
	if size < 512 || size > 65536 {
		return false
	}
	return (size & (size - 1)) == 0
}
