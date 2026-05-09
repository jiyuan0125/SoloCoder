package sqlite

import (
	"encoding/binary"
	"errors"
	"math"
)

const (
	BTREE_INTERIOR_INDEX    = 0x02
	BTREE_INTERIOR_TABLE    = 0x05
	BTREE_LEAF_INDEX        = 0x0A
	BTREE_LEAF_TABLE        = 0x0D
)

type btreePageHeader struct {
	pageType        uint8
	firstFreeblock  uint16
	cellCount       uint16
	cellContentArea uint16
	fragmentedFree  uint8
	rightPointer    uint32
}

func parseBtreeHeader(data []byte, pageStart int) (*btreePageHeader, error) {
	if pageStart+8 > len(data) {
		return nil, errors.New("page too short for btree header")
	}

	h := &btreePageHeader{
		pageType:        data[pageStart],
		firstFreeblock:  binary.BigEndian.Uint16(data[pageStart+1 : pageStart+3]),
		cellCount:       binary.BigEndian.Uint16(data[pageStart+3 : pageStart+5]),
		cellContentArea: binary.BigEndian.Uint16(data[pageStart+5 : pageStart+7]),
		fragmentedFree:  data[pageStart+7],
	}

	if h.pageType == BTREE_INTERIOR_INDEX || h.pageType == BTREE_INTERIOR_TABLE {
		if pageStart+12 > len(data) {
			return nil, errors.New("page too short for interior btree header")
		}
		h.rightPointer = binary.BigEndian.Uint32(data[pageStart+8 : pageStart+12])
	}

	return h, nil
}

func headerSize(pageType uint8) int {
	if pageType == BTREE_INTERIOR_INDEX || pageType == BTREE_INTERIOR_TABLE {
		return 12
	}
	return 8
}

func (db *Database) getPage(pageNum uint32) ([]byte, error) {
	page, ok := db.pages[pageNum]
	if !ok {
		return nil, errors.New("page not found")
	}
	return page, nil
}

func (db *Database) scanLeafTablePage(pageNum uint32, rows *[]*Row) error {
	page, err := db.getPage(pageNum)
	if err != nil {
		return err
	}

	var startOffset int
	if pageNum == 1 {
		startOffset = 100
	} else {
		startOffset = 0
	}

	header, err := parseBtreeHeader(page, startOffset)
	if err != nil {
		return nil
	}

	if header.pageType != BTREE_LEAF_TABLE {
		return nil
	}

	headerSizeVal := headerSize(header.pageType)
	cellPtrOffset := startOffset + headerSizeVal

	for i := 0; i < int(header.cellCount); i++ {
		ptrPos := cellPtrOffset + i*2
		if ptrPos+2 > len(page) {
			continue
		}
		cellOffset := int(binary.BigEndian.Uint16(page[ptrPos:ptrPos+2]))
		if cellOffset >= len(page) {
			continue
		}

		row, err := parseLeafTableCell(page, cellOffset)
		if err == nil && row != nil {
			*rows = append(*rows, row)
		}
	}

	return nil
}

func (db *Database) scanInteriorTablePage(pageNum uint32, rows *[]*Row) error {
	page, err := db.getPage(pageNum)
	if err != nil {
		return err
	}

	var startOffset int
	if pageNum == 1 {
		startOffset = 100
	} else {
		startOffset = 0
	}

	header, err := parseBtreeHeader(page, startOffset)
	if err != nil {
		return nil
	}

	if header.pageType != BTREE_INTERIOR_TABLE {
		return nil
	}

	headerSizeVal := headerSize(header.pageType)
	cellPtrOffset := startOffset + headerSizeVal

	for i := 0; i < int(header.cellCount); i++ {
		ptrPos := cellPtrOffset + i*2
		if ptrPos+2 > len(page) {
			continue
		}
		cellOffset := int(binary.BigEndian.Uint16(page[ptrPos:ptrPos+2]))
		if cellOffset+4 > len(page) {
			continue
		}

		childPage := binary.BigEndian.Uint32(page[cellOffset : cellOffset+4])
		db.scanTablePage(childPage, rows)
	}

	if header.rightPointer > 0 {
		db.scanTablePage(header.rightPointer, rows)
	}

	return nil
}

func (db *Database) scanTablePage(pageNum uint32, rows *[]*Row) error {
	if pageNum == 0 {
		return nil
	}

	page, err := db.getPage(pageNum)
	if err != nil {
		return nil
	}

	var startOffset int
	if pageNum == 1 {
		startOffset = 100
	} else {
		startOffset = 0
	}

	if startOffset >= len(page) {
		return nil
	}

	pageType := page[startOffset]

	switch pageType {
	case BTREE_INTERIOR_TABLE:
		return db.scanInteriorTablePage(pageNum, rows)
	case BTREE_LEAF_TABLE:
		return db.scanLeafTablePage(pageNum, rows)
	default:
		return nil
	}
}

func parseLeafTableCell(page []byte, offset int) (*Row, error) {
	if offset >= len(page) {
		return nil, errors.New("cell offset out of bounds")
	}

	payloadLen, n1 := readVarint(page, offset)
	offset += n1
	if offset >= len(page) {
		return nil, errors.New("invalid payload")
	}

	rowID, n2 := readVarint(page, offset)
	offset += n2

	payloadStart := offset
	payloadEnd := offset + int(payloadLen)
	if payloadEnd > len(page) {
		payloadEnd = len(page)
	}

	if payloadStart >= payloadEnd {
		return nil, errors.New("empty payload")
	}

	return parseRecord(page[payloadStart:payloadEnd], rowID)
}

func parseRecord(payload []byte, rowID uint64) (*Row, error) {
	if len(payload) == 0 {
		return nil, errors.New("empty payload")
	}

	headerLen, n := readVarint(payload, 0)
	if n == 0 {
		return nil, errors.New("invalid header")
	}

	headerEnd := int(headerLen)
	if headerEnd > len(payload) {
		headerEnd = len(payload)
	}

	headerOffset := n
	var serialTypes []uint64

	for headerOffset < headerEnd {
		st, m := readVarint(payload, headerOffset)
		serialTypes = append(serialTypes, st)
		headerOffset += m
	}

	valuesOffset := headerEnd
	row := &Row{
		values: make([]interface{}, 0, len(serialTypes)),
	}

	for i, st := range serialTypes {
		var value interface{}
		var size int

		switch {
		case st == 0:
			value = nil
			size = 0
		case st == 1:
			size = 1
			if valuesOffset+size <= len(payload) {
				value = int64(int8(payload[valuesOffset]))
			}
		case st == 2:
			size = 2
			if valuesOffset+size <= len(payload) {
				value = int64(int16(binary.BigEndian.Uint16(payload[valuesOffset : valuesOffset+size])))
			}
		case st == 3:
			size = 3
			if valuesOffset+size <= len(payload) {
				b := payload[valuesOffset : valuesOffset+size]
				val := int32(int8(b[0]))<<16 | int32(b[1])<<8 | int32(b[2])
				value = int64(val)
			}
		case st == 4:
			size = 4
			if valuesOffset+size <= len(payload) {
				value = int64(int32(binary.BigEndian.Uint32(payload[valuesOffset : valuesOffset+size])))
			}
		case st == 5:
			size = 6
			if valuesOffset+size <= len(payload) {
				b := payload[valuesOffset : valuesOffset+size]
				val := int64(int8(b[0]))<<40 |
					int64(b[1])<<32 |
					int64(b[2])<<24 |
					int64(b[3])<<16 |
					int64(b[4])<<8 |
					int64(b[5])
				value = val
			}
		case st == 6:
			size = 8
			if valuesOffset+size <= len(payload) {
				value = int64(binary.BigEndian.Uint64(payload[valuesOffset : valuesOffset+size]))
			}
		case st == 7:
			size = 8
			if valuesOffset+size <= len(payload) {
				bits := binary.BigEndian.Uint64(payload[valuesOffset : valuesOffset+size])
				value = math.Float64frombits(bits)
			}
		case st == 8:
			value = int64(0)
			size = 0
		case st == 9:
			value = int64(1)
			size = 0
		case st >= 12 && st%2 == 0:
			size = int((st - 12) / 2)
			if valuesOffset+size <= len(payload) {
				value = string(payload[valuesOffset : valuesOffset+size])
			}
		case st >= 13:
			size = int((st - 13) / 2)
			if valuesOffset+size <= len(payload) {
				blob := make([]byte, size)
				copy(blob, payload[valuesOffset:valuesOffset+size])
				value = blob
			}
		default:
			size = 0
			value = nil
		}

		if i == 0 && (st == 0 || (st >= 8 && st <= 9)) {
			row.values = append(row.values, int64(rowID))
		} else {
			row.values = append(row.values, value)
		}

		valuesOffset += size
	}

	return row, nil
}
