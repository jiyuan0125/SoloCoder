package internal

import (
	"fmt"
	"os"
	"sort"
)

type Statistics struct {
	OriginalTotal   int
	VerifiedTotal   int
	BucketCounts    map[string]int
	AdjustedCounts  map[string]int
	DiffDetails     []string
	IsConsistent    bool
}

func NewStatistics() *Statistics {
	return &Statistics{
		BucketCounts:   make(map[string]int),
		AdjustedCounts: make(map[string]int),
		DiffDetails:    []string{},
		IsConsistent:   true,
	}
}

func (s *Statistics) AdjustQuotas(newTotal int) {
	keys := make([]string, 0, len(s.BucketCounts))
	for k := range s.BucketCounts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		return
	}

	originalTotal := s.OriginalTotal
	if originalTotal == 0 {
		perBucket := newTotal / len(keys)
		remainder := newTotal % len(keys)

		for i, k := range keys {
			if i == 0 {
				s.AdjustedCounts[k] = perBucket + remainder
			} else {
				s.AdjustedCounts[k] = perBucket
			}
		}
		return
	}

	adjTotal := 0
	adjCounts := make(map[string]float64)
	ratio := float64(newTotal) / float64(originalTotal)

	var maxKey string
	maxCount := 0

	for k, v := range s.BucketCounts {
		adjusted := float64(v) * ratio
		floored := int(adjusted)
		s.AdjustedCounts[k] = floored
		adjTotal += floored
		adjCounts[k] = adjusted - float64(floored)

		if v > maxCount {
			maxCount = v
			maxKey = k
		}
	}

	deficit := newTotal - adjTotal
	if deficit > 0 {
		if maxKey != "" {
			s.AdjustedCounts[maxKey] += deficit
		}
	}
}

func (s *Statistics) PrintReport() {
	fmt.Println("========== 统计报表 ==========")
	fmt.Printf("原始数据总行数: %d\n", s.OriginalTotal)
	fmt.Printf("验证数据总行数: %d\n", s.VerifiedTotal)

	if s.OriginalTotal != s.VerifiedTotal {
		fmt.Println("⚠ 数据不一致!")
		fmt.Println("差异详情:")
		for _, diff := range s.DiffDetails {
			fmt.Printf("  %s\n", diff)
		}
	} else {
		fmt.Println("✓ 数据一致")
	}

	fmt.Printf("聚合时间段数: %d\n", len(s.BucketCounts))

	if len(s.AdjustedCounts) > 0 {
		fmt.Println("\n调整后的配额:")
		keys := make([]string, 0, len(s.AdjustedCounts))
		for k := range s.AdjustedCounts {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			fmt.Printf("  %s: %d\n", k, s.AdjustedCounts[k])
		}
	}
	fmt.Println("==============================")
}

type Verifier struct {
	config *Config
}

func NewVerifier(config *Config) *Verifier {
	return &Verifier{config: config}
}

func (v *Verifier) VerifyAndGenerateReport(
	filename string,
	originalTotal int,
	bucketCounts map[string]int,
	loc interface{},
) (*Statistics, error) {
	stats := NewStatistics()
	stats.OriginalTotal = originalTotal

	for k, v := range bucketCounts {
		stats.BucketCounts[k] = v
	}

	reader, err := NewCSVReader(filename)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	_, err = reader.ReadHeader()
	if err != nil {
		return nil, err
	}

	verifiedTotal := 0
	timeColIdx := v.config.TimeColumnIndex
	valueColIdxs := make([]int, 0, len(v.config.Columns))
	for _, col := range v.config.Columns {
		valueColIdxs = append(valueColIdxs, col.Index)
	}

	reader.ReadAll(timeColIdx, valueColIdxs, func(row DataRow, valid bool, warning string) {
		if !valid {
			fmt.Fprintln(os.Stderr, "警告:", warning)
			return
		}
		verifiedTotal++
	})

	stats.VerifiedTotal = verifiedTotal

	if originalTotal != verifiedTotal {
		stats.IsConsistent = false
		stats.DiffDetails = append(stats.DiffDetails,
			fmt.Sprintf("行数不一致: 原始 %d 行 vs 验证 %d 行", originalTotal, verifiedTotal))
		fmt.Fprintf(os.Stderr, "⚠ 核对不一致: 原始 %d 行 vs 验证 %d 行\n", originalTotal, verifiedTotal)
	}

	stats.AdjustQuotas(verifiedTotal)

	return stats, nil
}
