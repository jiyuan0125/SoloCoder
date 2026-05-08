package cursorpaginator

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"cursor-paginator/pkg/api"
)

const (
	DefaultLimit = 10
	MaxLimit     = 100
)

type Cursor struct {
	LastID    int
	LastValue interface{}
	SortField string
	SortOrder api.SortOrder
	Direction Direction
}

type Direction string

const (
	DirectionNext Direction = "next"
	DirectionPrev Direction = "prev"
)

func NewCursor(lastID int, lastValue interface{}, sortField string, sortOrder api.SortOrder, direction Direction) *Cursor {
	return &Cursor{
		LastID:    lastID,
		LastValue: lastValue,
		SortField: sortField,
		SortOrder: sortOrder,
		Direction: direction,
	}
}

func EncodeCursor(cursor *Cursor) (string, error) {
	if cursor == nil {
		return "", nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(cursor)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var cursor Cursor
	buf := bytes.NewBuffer(decoded)
	dec := gob.NewDecoder(buf)
	err = dec.Decode(&cursor)
	if err != nil {
		return nil, err
	}

	return &cursor, nil
}

type Comparable interface {
	GetID() int
	GetFieldValue(field string) interface{}
}

func sanitizeLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

func sortUsers(users []*api.User, sortField string, sortOrder api.SortOrder) {
	sort.Slice(users, func(i, j int) bool {
		iVal := users[i].GetFieldValue(sortField)
		jVal := users[j].GetFieldValue(sortField)

		result := compareValues(iVal, jVal)
		if sortOrder == api.SortOrderDesc {
			result = -result
		}

		if result == 0 {
			return users[i].ID < users[j].ID
		}
		return result < 0
	})
}

func compareValues(a, b interface{}) int {
	switch av := a.(type) {
	case int:
		bv := b.(int)
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		default:
			return 0
		}
	case int64:
		bv := b.(int64)
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		default:
			return 0
		}
	case string:
		bv := b.(string)
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		default:
			return 0
		}
	default:
		return 0
	}
}

func shouldInclude(item *api.User, cursor *Cursor, sortField string, sortOrder api.SortOrder, isNext bool) bool {
	if cursor == nil {
		return true
	}

	itemVal := item.GetFieldValue(sortField)
	result := compareValues(itemVal, cursor.LastValue)

	if isNext {
		if sortOrder == api.SortOrderAsc {
			if result > 0 {
				return true
			}
			if result == 0 {
				return item.ID > cursor.LastID
			}
			return false
		} else {
			if result < 0 {
				return true
			}
			if result == 0 {
				return item.ID > cursor.LastID
			}
			return false
		}
	} else {
		if sortOrder == api.SortOrderAsc {
			if result < 0 {
				return true
			}
			if result == 0 {
				return item.ID < cursor.LastID
			}
			return false
		} else {
			if result > 0 {
				return true
			}
			if result == 0 {
				return item.ID < cursor.LastID
			}
			return false
		}
	}
}

func PaginateUsers(allUsers []*api.User, req *api.PageRequest) (*api.PageResponse, error) {
	limit := sanitizeLimit(req.Limit)
	sortField := req.SortField
	if sortField == "" {
		sortField = "id"
	}
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = api.SortOrderAsc
	}

	sortedUsers := make([]*api.User, len(allUsers))
	copy(sortedUsers, allUsers)
	sortUsers(sortedUsers, sortField, sortOrder)

	total := len(sortedUsers)

	var startCursor *Cursor
	var isNext bool

	if req.Cursor != "" {
		decoded, err := DecodeCursor(req.Cursor)
		if err != nil {
			return nil, err
		}
		startCursor = decoded
		isNext = decoded.Direction == DirectionNext
	} else if req.Previous != "" {
		decoded, err := DecodeCursor(req.Previous)
		if err != nil {
			return nil, err
		}
		startCursor = decoded
		isNext = false
	}

	var filtered []*api.User
	for _, user := range sortedUsers {
		if shouldInclude(user, startCursor, sortField, sortOrder, isNext) {
			filtered = append(filtered, user)
		}
	}

	if !isNext && startCursor != nil {
		reverseSlice(filtered)
	}

	actualLimit := limit + 1
	var hasNext, hasPrev bool
	var resultData []*api.User

	if len(filtered) > actualLimit {
		resultData = filtered[:actualLimit]
		if isNext {
			hasNext = true
			hasPrev = startCursor != nil
		} else {
			hasPrev = true
			hasNext = len(filtered) > actualLimit
		}
	} else {
		resultData = filtered
		hasNext = false
		hasPrev = startCursor != nil
	}

	if len(resultData) > limit {
		resultData = resultData[:limit]
	}

	if !isNext && startCursor != nil {
		reverseSlice(resultData)
	}

	dataJSON := make([]json.RawMessage, 0, len(resultData))
	for _, user := range resultData {
		jsonBytes, err := json.Marshal(user)
		if err != nil {
			return nil, err
		}
		dataJSON = append(dataJSON, jsonBytes)
	}

	response := &api.PageResponse{
		Data:  dataJSON,
		Total: total,
		Limit: limit,
	}

	if len(resultData) > 0 {
		lastItem := resultData[len(resultData)-1]
		nextCursor := NewCursor(
			lastItem.ID,
			lastItem.GetFieldValue(sortField),
			sortField,
			sortOrder,
			DirectionNext,
		)
		nextCursorStr, err := EncodeCursor(nextCursor)
		if err != nil {
			return nil, err
		}
		response.NextCursor = nextCursorStr
		response.HasNext = hasNext

		firstItem := resultData[0]
		prevCursor := NewCursor(
			firstItem.ID,
			firstItem.GetFieldValue(sortField),
			sortField,
			sortOrder,
			DirectionPrev,
		)
		prevCursorStr, err := EncodeCursor(prevCursor)
		if err != nil {
			return nil, err
		}
		response.PrevCursor = prevCursorStr
		response.HasPrev = hasPrev
	}

	return response, nil
}

func reverseSlice(s []*api.User) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func GenerateSampleUsers(count int) []*api.User {
	users := make([]*api.User, 0, count)
	for i := 1; i <= count; i++ {
		users = append(users, &api.User{
			ID:        i,
			Name:      fmt.Sprintf("User %d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			CreatedAt: int64(1000000000 + i*1000),
		})
	}
	return users
}

func init() {
	gob.Register(&api.User{})
	gob.Register(int(0))
	gob.Register(int64(0))
	gob.Register(string(""))
	gob.Register(&Cursor{})
	gob.Register(strconv.IntSize)
}
