package iplocation

import (
	"sort"
	"sync"
)

type LocationDB struct {
	ranges []IPRange
	mu     sync.RWMutex
}

func NewLocationDB() *LocationDB {
	return &LocationDB{
		ranges: make([]IPRange, 0),
	}
}

func (db *LocationDB) AddRange(startIP, endIP string, loc Location) error {
	start, err := IPToUint32(startIP)
	if err != nil {
		return err
	}

	end, err := IPToUint32(endIP)
	if err != nil {
		return err
	}

	if start > end {
		return ErrInvalidIPRange
	}

	if start == 0 || end == 0xFFFFFFFF {
		return ErrSpecialAddress
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	for _, r := range db.ranges {
		if rangesOverlap(start, end, r.StartIP, r.EndIP) {
			return ErrRangeOverlap
		}
	}

	db.ranges = append(db.ranges, IPRange{
		StartIP:  start,
		EndIP:    end,
		Location: loc,
	})

	db.sortRanges()

	return nil
}

func (db *LocationDB) sortRanges() {
	sort.Slice(db.ranges, func(i, j int) bool {
		if db.ranges[i].StartIP == db.ranges[j].StartIP {
			return db.ranges[i].Size() < db.ranges[j].Size()
		}
		return db.ranges[i].StartIP < db.ranges[j].StartIP
	})
}

func rangesOverlap(s1, e1, s2, e2 uint32) bool {
	return !(e1 < s2 || s1 > e2)
}

func (db *LocationDB) Query(ipStr string) (Location, error) {
	ip, err := IPToUint32(ipStr)
	if err != nil {
		return Location{}, err
	}

	if ip == 0 || ip == 0xFFFFFFFF {
		return Location{}, ErrSpecialAddress
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	var bestMatch *IPRange
	for i := range db.ranges {
		r := &db.ranges[i]
		if ip >= r.StartIP && ip <= r.EndIP {
			if bestMatch == nil || r.Size() < bestMatch.Size() {
				bestMatch = r
			}
		}
	}

	if bestMatch == nil {
		return Location{}, ErrNotFound
	}

	return bestMatch.Location, nil
}

func (db *LocationDB) BatchQuery(ipStrs []string) ([]BatchResult, error) {
	results := make([]BatchResult, len(ipStrs))

	for i, ipStr := range ipStrs {
		loc, err := db.Query(ipStr)
		results[i] = BatchResult{
			IP:       ipStr,
			Location: loc,
			Error:    err,
		}
	}

	return results, nil
}

type BatchResult struct {
	IP       string
	Location Location
	Error    error
}

func (db *LocationDB) SameProvince(ipStr1, ipStr2 string) (bool, error) {
	isPrivate1, err := IsPrivateIP(ipStr1)
	if err != nil {
		return false, err
	}

	isPrivate2, err := IsPrivateIP(ipStr2)
	if err != nil {
		return false, err
	}

	if isPrivate1 || isPrivate2 {
		return isPrivate1 && isPrivate2, nil
	}

	loc1, err := db.Query(ipStr1)
	if err != nil {
		return false, err
	}

	loc2, err := db.Query(ipStr2)
	if err != nil {
		return false, err
	}

	return loc1.Province == loc2.Province, nil
}

func (db *LocationDB) RangeCount() int {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return len(db.ranges)
}

func (db *LocationDB) Clear() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.ranges = make([]IPRange, 0)
}
