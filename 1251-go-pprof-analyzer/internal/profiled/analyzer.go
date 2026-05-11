package profiled

import (
	"compress/gzip"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/google/pprof/profile"
)

type ProfileType string

const (
	ProfileTypeHeap ProfileType = "heap"
	ProfileTypeCPU  ProfileType = "cpu"
	ProfileTypeOther ProfileType = "other"
)

type FunctionStat struct {
	FullName    string
	PackageName string
	FuncName    string
	Bytes       int64
	Objects     int64
	Percentage  float64
}

type AnalysisResult struct {
	ProfileType ProfileType
	TotalBytes  int64
	TotalObjects int64
	Functions   []FunctionStat
}

var runtimePackages = []string{
	"runtime.",
	"runtime/internal/",
	"sync/",
	"sync/atomic/",
	"internal/",
	"reflect.",
	"syscall.",
}

func isRuntimeFunction(fullName string) bool {
	for _, pkg := range runtimePackages {
		if strings.HasPrefix(fullName, pkg) {
			return true
		}
	}
	return false
}

func parseFunctionName(fullName string) (packageName, funcName string) {
	lastSlash := strings.LastIndex(fullName, "/")
	if lastSlash == -1 {
		dotIdx := strings.Index(fullName, ".")
		if dotIdx == -1 {
			return "", fullName
		}
		return fullName[:dotIdx], fullName[dotIdx+1:]
	}
	
	remain := fullName[lastSlash+1:]
	dotIdx := strings.Index(remain, ".")
	if dotIdx == -1 {
		return fullName, ""
	}
	return fullName[:lastSlash+dotIdx], remain[dotIdx+1:]
}

func findTopUserFunction(locations []*profile.Location) string {
	for i := len(locations) - 1; i >= 0; i-- {
		loc := locations[i]
		for _, line := range loc.Line {
			if line.Function == nil {
				continue
			}
			fullName := line.Function.Name
			if !isRuntimeFunction(fullName) {
				return fullName
			}
		}
	}
	
	if len(locations) > 0 {
		loc := locations[len(locations)-1]
		if len(loc.Line) > 0 && loc.Line[0].Function != nil {
			return loc.Line[0].Function.Name
		}
	}
	return "unknown"
}

func AnalyzeFile(filePath string) (*AnalysisResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	
	return Analyze(f)
}

func Analyze(r io.Reader) (*AnalysisResult, error) {
	var reader io.Reader = r
	
	buf := make([]byte, 2)
	_, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	
	isGzip := buf[0] == 0x1f && buf[1] == 0x8b
	
	peekedReader := io.MultiReader(strings.NewReader(string(buf)), r)
	
	if isGzip {
		gzr, err := gzip.NewReader(peekedReader)
		if err != nil {
			return nil, err
		}
		defer gzr.Close()
		reader = gzr
	} else {
		reader = peekedReader
	}
	
	prof, err := profile.Parse(reader)
	if err != nil {
		return nil, err
	}
	
	profileType := detectProfileType(prof)
	
	functionStats := make(map[string]*FunctionStat)
	var totalBytes, totalObjects int64
	
	bytesIdx := -1
	objectsIdx := -1
	
	for i, st := range prof.SampleType {
		switch st.Type {
		case "inuse_space":
			bytesIdx = i
		case "inuse_objects":
			objectsIdx = i
		case "alloc_space":
			if bytesIdx == -1 {
				bytesIdx = i
			}
		case "alloc_objects":
			if objectsIdx == -1 {
				objectsIdx = i
			}
		case "samples":
			if bytesIdx == -1 {
				bytesIdx = i
			}
		case "time":
			if bytesIdx == -1 {
				bytesIdx = i
			}
		}
	}
	
	if bytesIdx == -1 && len(prof.SampleType) > 0 {
		bytesIdx = 0
	}
	
	for _, sample := range prof.Sample {
		fullName := findTopUserFunction(sample.Location)
		if fullName == "" {
			continue
		}
		
		var bytes, objects int64
		
		if bytesIdx >= 0 && bytesIdx < len(sample.Value) {
			bytes = sample.Value[bytesIdx]
		}
		if objectsIdx >= 0 && objectsIdx < len(sample.Value) {
			objects = sample.Value[objectsIdx]
		} else if profileType == ProfileTypeCPU {
			objects = bytes
		}
		
		totalBytes += bytes
		totalObjects += objects
		
		if stat, exists := functionStats[fullName]; exists {
			stat.Bytes += bytes
			stat.Objects += objects
		} else {
			pkgName, funcName := parseFunctionName(fullName)
			functionStats[fullName] = &FunctionStat{
				FullName:    fullName,
				PackageName: pkgName,
				FuncName:    funcName,
				Bytes:       bytes,
				Objects:     objects,
			}
		}
	}
	
	result := &AnalysisResult{
		ProfileType:  profileType,
		TotalBytes:   totalBytes,
		TotalObjects: totalObjects,
		Functions:    make([]FunctionStat, 0, len(functionStats)),
	}
	
	for _, stat := range functionStats {
		if totalBytes > 0 {
			stat.Percentage = float64(stat.Bytes) / float64(totalBytes) * 100.0
		}
		result.Functions = append(result.Functions, *stat)
	}
	
	sort.Slice(result.Functions, func(i, j int) bool {
		return result.Functions[i].Bytes > result.Functions[j].Bytes
	})
	
	return result, nil
}

func detectProfileType(prof *profile.Profile) ProfileType {
	for _, st := range prof.SampleType {
		switch st.Type {
		case "inuse_space", "inuse_objects", "alloc_space", "alloc_objects":
			return ProfileTypeHeap
		}
	}
	
	for _, st := range prof.SampleType {
		switch st.Type {
		case "samples", "time", "cpu":
			return ProfileTypeCPU
		}
	}
	
	return ProfileTypeOther
}

func (r *AnalysisResult) GetTop(n int) ([]FunctionStat, *FunctionStat) {
	if n <= 0 || n >= len(r.Functions) {
		return r.Functions, nil
	}
	
	top := r.Functions[:n]
	
	if n >= len(r.Functions) {
		return top, nil
	}
	
	other := &FunctionStat{
		FullName: "其他",
		FuncName: "其他",
	}
	
	for _, stat := range r.Functions[n:] {
		other.Bytes += stat.Bytes
		other.Objects += stat.Objects
	}
	
	if r.TotalBytes > 0 {
		other.Percentage = float64(other.Bytes) / float64(r.TotalBytes) * 100.0
	}
	
	return top, other
}
