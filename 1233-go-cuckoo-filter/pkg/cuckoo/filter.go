package cuckoo

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/rand"
)

const (
	FingerprintSize = 1
	BucketSize      = 4
	MaxKicks        = 500
	FPZero          = 0
)

type Fingerprint byte
type Bucket [BucketSize]Fingerprint

type Filter struct {
	buckets []Bucket
	size    uint64
	count   uint64
}

type Result struct {
	Success bool
	Error   error
}

type Stats struct {
	UsedSlots  uint64
	TotalSlots uint64
	LoadRate   float64
}

func New(capacity uint64) *Filter {
	if capacity == 0 {
		capacity = 1
	}
	bucketCount := capacity / BucketSize
	if bucketCount == 0 {
		bucketCount = 1
	}
	if capacity%BucketSize != 0 {
		bucketCount++
	}
	return &Filter{
		buckets: make([]Bucket, bucketCount),
		size:    bucketCount,
		count:   0,
	}
}

func (f *Filter) Insert(data []byte) Result {
	hashVal := f.hash(data)
	fp := Fingerprint(hashVal[0])
	idx1 := f.hashToIndex(hashVal)
	idx2 := f.alternateIndex(idx1, fp)

	if fp == FPZero {
		fp++
	}

	if f.tryInsert(idx1, fp) {
		f.count++
		return Result{Success: true}
	}
	if f.tryInsert(idx2, fp) {
		f.count++
		return Result{Success: true}
	}

	currentIdx := idx1
	if rand.Intn(2) == 1 {
		currentIdx = idx2
	}
	currentFp := fp

	for kick := 0; kick < MaxKicks; kick++ {
		entryIdx := rand.Intn(BucketSize)
		oldFp := f.buckets[currentIdx][entryIdx]
		f.buckets[currentIdx][entryIdx] = currentFp
		currentFp = oldFp
		currentIdx = f.alternateIndex(currentIdx, currentFp)

		if f.tryInsert(currentIdx, currentFp) {
			f.count++
			return Result{Success: true}
		}
	}

	return Result{Success: false, Error: errors.New("filter is full")}
}

func (f *Filter) Lookup(data []byte) bool {
	hashVal := f.hash(data)
	fp := Fingerprint(hashVal[0])
	idx1 := f.hashToIndex(hashVal)
	idx2 := f.alternateIndex(idx1, fp)

	if fp == FPZero {
		fp++
	}

	return f.findInBucket(idx1, fp) || f.findInBucket(idx2, fp)
}

func (f *Filter) Delete(data []byte) Result {
	hashVal := f.hash(data)
	fp := Fingerprint(hashVal[0])
	idx1 := f.hashToIndex(hashVal)
	idx2 := f.alternateIndex(idx1, fp)

	if fp == FPZero {
		fp++
	}

	if f.removeFromBucket(idx1, fp) || f.removeFromBucket(idx2, fp) {
		f.count--
		return Result{Success: true}
	}
	return Result{Success: false, Error: errors.New("element not found")}
}

func (f *Filter) Stats() Stats {
	total := f.size * BucketSize
	loadRate := 0.0
	if total > 0 {
		loadRate = float64(f.count) / float64(total)
	}
	return Stats{
		UsedSlots:  f.count,
		TotalSlots: total,
		LoadRate:   loadRate,
	}
}

func (f *Filter) Expand(newCapacity uint64) *Filter {
	newFilter := New(newCapacity)
	for _, bucket := range f.buckets {
		for _, fp := range bucket {
			if fp == FPZero {
				continue
			}
			newFilter.insertFP(fp, 0)
		}
	}
	return newFilter
}

func (f *Filter) hash(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func (f *Filter) hashToIndex(hashVal []byte) uint64 {
	n := binary.BigEndian.Uint64(hashVal[0:8])
	return n % f.size
}

func (f *Filter) alternateIndex(idx uint64, fp Fingerprint) uint64 {
	h := sha256.Sum256([]byte{byte(fp)})
	alt := binary.BigEndian.Uint64(h[0:8])
	return (idx ^ alt) % f.size
}

func (f *Filter) tryInsert(idx uint64, fp Fingerprint) bool {
	for i := range f.buckets[idx] {
		if f.buckets[idx][i] == FPZero {
			f.buckets[idx][i] = fp
			return true
		}
	}
	return false
}

func (f *Filter) findInBucket(idx uint64, fp Fingerprint) bool {
	for _, fIn := range f.buckets[idx] {
		if fIn == fp {
			return true
		}
	}
	return false
}

func (f *Filter) removeFromBucket(idx uint64, fp Fingerprint) bool {
	for i, fIn := range f.buckets[idx] {
		if fIn == fp {
			f.buckets[idx][i] = FPZero
			return true
		}
	}
	return false
}

func (f *Filter) insertFP(fp Fingerprint, originalIdx uint64) {
	if fp == FPZero {
		return
	}

	idx1 := uint64(0)
	if f.size > 0 {
		idx1 = uint64(fp) % f.size
	}
	idx2 := f.alternateIndex(idx1, fp)

	if f.tryInsert(idx1, fp) || f.tryInsert(idx2, fp) {
		f.count++
		return
	}

	currentIdx := idx1
	if rand.Intn(2) == 1 {
		currentIdx = idx2
	}
	currentFp := fp

	for kick := 0; kick < MaxKicks; kick++ {
		entryIdx := rand.Intn(BucketSize)
		oldFp := f.buckets[currentIdx][entryIdx]
		f.buckets[currentIdx][entryIdx] = currentFp
		currentFp = oldFp
		currentIdx = f.alternateIndex(currentIdx, currentFp)

		if f.tryInsert(currentIdx, currentFp) {
			f.count++
			return
		}
	}
}
