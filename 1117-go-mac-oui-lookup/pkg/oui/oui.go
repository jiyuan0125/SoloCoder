package oui

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
)

type OUIEntry struct {
	Prefix  []byte
	Vendor  string
	PrefixLen int
}

type Database struct {
	entries []OUIEntry
}

func NewDatabase() *Database {
	return &Database{
		entries: make([]OUIEntry, 0),
	}
}

func normalizePrefix(prefix string) ([]byte, error) {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.ToLower(prefix)
	prefix = strings.ReplaceAll(prefix, ":", "")
	prefix = strings.ReplaceAll(prefix, "-", "")
	prefix = strings.ReplaceAll(prefix, ".", "")

	if len(prefix)%2 != 0 {
		return nil, errors.New("invalid prefix length")
	}

	return hex.DecodeString(prefix)
}

func (db *Database) LoadFromReader(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue
		}

		prefixStr := strings.TrimSpace(parts[0])
		vendor := strings.TrimSpace(parts[1])

		if vendor == "" {
			continue
		}

		prefix, err := normalizePrefix(prefixStr)
		if err != nil {
			continue
		}

		db.AddEntry(prefix, vendor)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	db.sortEntries()
	return nil
}

func (db *Database) AddEntry(prefix []byte, vendor string) {
	existingIdx := -1
	for i, entry := range db.entries {
		if bytesEqual(entry.Prefix, prefix) {
			existingIdx = i
			break
		}
	}

	entry := OUIEntry{
		Prefix:    append([]byte(nil), prefix...),
		Vendor:    vendor,
		PrefixLen: len(prefix) * 8,
	}

	if existingIdx >= 0 {
		db.entries[existingIdx] = entry
	} else {
		db.entries = append(db.entries, entry)
	}
}

func (db *Database) sortEntries() {
	sort.Slice(db.entries, func(i, j int) bool {
		if db.entries[i].PrefixLen != db.entries[j].PrefixLen {
			return db.entries[i].PrefixLen > db.entries[j].PrefixLen
		}
		return bytesCompare(db.entries[i].Prefix, db.entries[j].Prefix) < 0
	})
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func bytesCompare(a, b []byte) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] < b[i] {
			return -1
		} else if a[i] > b[i] {
			return 1
		}
	}
	if len(a) < len(b) {
		return -1
	} else if len(a) > len(b) {
		return 1
	}
	return 0
}

func prefixMatches(mac []byte, prefix []byte) bool {
	if len(prefix) > len(mac) {
		return false
	}
	for i := range prefix {
		if mac[i] != prefix[i] {
			return false
		}
	}
	return true
}

func (db *Database) Lookup(mac []byte) (string, error) {
	if len(mac) < 3 {
		return "", errors.New("invalid MAC address")
	}

	for _, entry := range db.entries {
		if prefixMatches(mac, entry.Prefix) {
			return entry.Vendor, nil
		}
	}

	return "", fmt.Errorf("vendor not found")
}

func (db *Database) Size() int {
	return len(db.entries)
}
