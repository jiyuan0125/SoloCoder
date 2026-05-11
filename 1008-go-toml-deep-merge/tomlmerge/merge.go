package tomlmerge

import (
"fmt"
"io"
"os"
"regexp"
"strconv"
"strings"
"time"

"github.com/BurntSushi/toml"
)

type MergeStrategy int

const (
StrategyReplace MergeStrategy = iota
StrategyAppend
)

type ParsedConfig struct {
Data         map[string]interface{}
AppendPaths  map[string]bool
IndexedEdits map[string]*IndexedEdit
}

type IndexedEdit struct {
ArrayKey string
Index    int
Data     map[string]interface{}
}

var (
boolTrueRegex  = regexp.MustCompile(`\b(True|TRUE)\b`)
boolFalseRegex = regexp.MustCompile(`\b(False|FALSE)\b`)
)

func MergeFiles(basePath, envPath string) (map[string]interface{}, error) {
baseContent, err := readFileRaw(basePath)
if err != nil {
return nil, fmt.Errorf("failed to read base file: %w", err)
}
envContent, err := readFileRaw(envPath)
if err != nil {
return nil, fmt.Errorf("failed to read env file: %w", err)
}
return MergeStrings(baseContent, envContent)
}

func MergeStrings(baseStr, envStr string) (map[string]interface{}, error) {
baseParsed, err := parseWithPreprocessing(baseStr)
if err != nil {
return nil, fmt.Errorf("failed to parse base: %w", err)
}
envParsed, err := parseWithPreprocessing(envStr)
if err != nil {
return nil, fmt.Errorf("failed to parse env: %w", err)
}
return mergeParsed(baseParsed, envParsed)
}

func Merge(base, env map[string]interface{}) (map[string]interface{}, error) {
result := deepCopy(base).(map[string]interface{})
if err := mergeMapsWithAppend(result, env, map[string]bool{}, ""); err != nil {
return nil, err
}
return result, nil
}

func mergeParsed(base, env *ParsedConfig) (map[string]interface{}, error) {
result := deepCopy(base.Data).(map[string]interface{})
if err := applyIndexedEdits(result, env.IndexedEdits); err != nil {
return nil, err
}
if err := mergeMapsWithAppend(result, env.Data, env.AppendPaths, ""); err != nil {
return nil, err
}
return result, nil
}

func applyIndexedEdits(data map[string]interface{}, edits map[string]*IndexedEdit) error {
for _, edit := range edits {
sliceVal, exists := data[edit.ArrayKey]
if !exists {
return fmt.Errorf("array table %s does not exist", edit.ArrayKey)
}
switch typedSlice := sliceVal.(type) {
case []map[string]interface{}:
if edit.Index < 0 || edit.Index >= len(typedSlice) {
return fmt.Errorf("index %d out of bounds for array table %s (length %d)", edit.Index, edit.ArrayKey, len(typedSlice))
}
if err := mergeMapsSimple(typedSlice[edit.Index], edit.Data); err != nil {
return err
}
case []interface{}:
if edit.Index < 0 || edit.Index >= len(typedSlice) {
return fmt.Errorf("index %d out of bounds for array %s (length %d)", edit.Index, edit.ArrayKey, len(typedSlice))
}
item := typedSlice[edit.Index]
if itemMap, ok := item.(map[string]interface{}); ok {
if err := mergeMapsSimple(itemMap, edit.Data); err != nil {
return err
}
} else {
typedSlice[edit.Index] = edit.Data
}
default:
return fmt.Errorf("%s is not an array table", edit.ArrayKey)
}
}
return nil
}

func mergeMapsSimple(dst, src map[string]interface{}) error {
for k, v := range src {
dst[k] = normalizeValue(v)
}
return nil
}

func mergeMapsWithAppend(dst, src map[string]interface{}, appendPaths map[string]bool, currentPath string) error {
for key, srcVal := range src {
var path string
if currentPath == "" {
path = key
} else {
path = currentPath + "." + key
}
isAppend := isInAppendMode(path, appendPaths)
srcVal = normalizeValue(srcVal)
dstVal, exists := dst[key]
if !exists {
dst[key] = srcVal
continue
}
dstVal = normalizeValue(dstVal)
if isAppend {
if err := handleAppendMerge(dst, key, srcVal, dstVal); err != nil {
return err
}
continue
}
if !sameValueType(dstVal, srcVal) {
dst[key] = srcVal
continue
}
switch typedSrc := srcVal.(type) {
case map[string]interface{}:
typedDst, ok := dstVal.(map[string]interface{})
if !ok {
dst[key] = typedSrc
continue
}
if err := mergeMapsWithAppend(typedDst, typedSrc, appendPaths, path); err != nil {
return err
}
case []interface{}:
dst[key] = typedSrc
default:
dst[key] = srcVal
}
}
return nil
}

func isInAppendMode(path string, appendPaths map[string]bool) bool {
if appendPaths[path] {
return true
}
for ap := range appendPaths {
if strings.HasPrefix(path, ap+".") {
return true
}
}
return false
}

func handleAppendMerge(dst map[string]interface{}, key string, srcVal, dstVal interface{}) error {
switch typedSrc := srcVal.(type) {
case map[string]interface{}:
typedDst, ok := dstVal.(map[string]interface{})
if !ok {
dst[key] = typedSrc
return nil
}
for k, v := range typedSrc {
v = normalizeValue(v)
dstV, exists := typedDst[k]
if !exists {
typedDst[k] = v
continue
}
dstV = normalizeValue(dstV)
if typedSrcSlice, ok := v.([]interface{}); ok {
if typedDstSlice, ok := dstV.([]interface{}); ok {
typedDst[k] = append(typedDstSlice, typedSrcSlice...)
continue
}
}
if typedSrcSlice, ok := v.([]map[string]interface{}); ok {
if typedDstSlice, ok := dstV.([]map[string]interface{}); ok {
typedDst[k] = appendArrayOfMaps(typedDstSlice, typedSrcSlice)
continue
}
}
typedDst[k] = v
}
case []interface{}:
dstSlice, ok := dstVal.([]interface{})
if !ok {
dst[key] = typedSrc
return nil
}
if len(typedSrc) == 0 {
return nil
}
firstSrc := normalizeValue(typedSrc[0])
if _, ok := firstSrc.(map[string]interface{}); ok {
for _, srcItem := range typedSrc {
typedSrcItem, ok := normalizeValue(srcItem).(map[string]interface{})
if !ok {
dstSlice = append(dstSlice, srcItem)
continue
}
found := false
for i, dstItem := range dstSlice {
typedDstItem, ok := normalizeValue(dstItem).(map[string]interface{})
if !ok {
continue
}
if canMergeTableItems(typedDstItem, typedSrcItem) {
if err := mergeMapsSimple(typedDstItem, typedSrcItem); err != nil {
return err
}
dstSlice[i] = typedDstItem
found = true
break
}
}
if !found {
dstSlice = append(dstSlice, typedSrcItem)
}
}
dst[key] = dstSlice
return nil
}
dst[key] = append(dstSlice, typedSrc...)
case []map[string]interface{}:
dstSlice, ok := dstVal.([]map[string]interface{})
if !ok {
dst[key] = typedSrc
return nil
}
dst[key] = appendArrayOfMaps(dstSlice, typedSrc)
default:
dst[key] = srcVal
}
return nil
}

func appendArrayOfMaps(dst, src []map[string]interface{}) []map[string]interface{} {
for _, srcItem := range src {
cp := make(map[string]interface{})
for k, v := range srcItem {
cp[k] = normalizeValue(v)
}
dst = append(dst, cp)
}
return dst
}

func canMergeTableItems(dst, src map[string]interface{}) bool {
for k := range src {
if _, exists := dst[k]; exists {
return true
}
}
return false
}

func normalizeValue(v interface{}) interface{} {
switch typed := v.(type) {
case time.Time:
return typed.UTC()
case bool:
return typed
default:
return v
}
}

func sameValueType(a, b interface{}) bool {
if a == nil || b == nil {
return false
}
_, aMap := a.(map[string]interface{})
_, bMap := b.(map[string]interface{})
if aMap && bMap {
return true
}
if aMap || bMap {
return false
}
_, aSlice := a.([]interface{})
_, bSlice := b.([]interface{})
if aSlice && bSlice {
return true
}
if aSlice || bSlice {
return false
}
_, aTime := a.(time.Time)
_, bTime := b.(time.Time)
if aTime && bTime {
return true
}
if aTime || bTime {
return false
}
_, aBool := a.(bool)
_, bBool := b.(bool)
if aBool && bBool {
return true
}
if aBool || bBool {
return false
}
_, aStr := a.(string)
_, bStr := b.(string)
if aStr && bStr {
return true
}
if aStr || bStr {
return false
}
_, aInt := a.(int64)
_, aFloat := a.(float64)
_, bInt := b.(int64)
_, bFloat := b.(float64)
if (aInt || aFloat) && (bInt || bFloat) {
return true
}
return false
}

func deepCopy(v interface{}) interface{} {
switch typed := v.(type) {
case map[string]interface{}:
cp := make(map[string]interface{}, len(typed))
for k, val := range typed {
cp[k] = deepCopy(val)
}
return cp
case []interface{}:
cp := make([]interface{}, len(typed))
for i, val := range typed {
cp[i] = deepCopy(val)
}
return cp
case []map[string]interface{}:
cp := make([]map[string]interface{}, len(typed))
for i, val := range typed {
cp[i] = deepCopyMap(val)
}
return cp
case time.Time:
return typed.UTC()
default:
return v
}
}

func deepCopyMap(m map[string]interface{}) map[string]interface{} {
cp := make(map[string]interface{}, len(m))
for k, v := range m {
cp[k] = deepCopy(v)
}
return cp
}

func Encode(w io.Writer, data map[string]interface{}) error {
enc := toml.NewEncoder(w)
enc.Indent = ""
return enc.Encode(data)
}

func DecodeString(s string) (map[string]interface{}, error) {
parsed, err := parseWithPreprocessing(s)
if err != nil {
return nil, err
}
return parsed.Data, nil
}

func DecodeFile(path string) (map[string]interface{}, error) {
content, err := readFileRaw(path)
if err != nil {
return nil, err
}
return DecodeString(content)
}

func readFileRaw(path string) (string, error) {
data, err := os.ReadFile(path)
if err != nil {
return "", err
}
return string(data), nil
}

func parseWithPreprocessing(input string) (*ParsedConfig, error) {
lines := strings.Split(input, "\n")
appendPaths := make(map[string]bool)
indexedEdits := make(map[string]*IndexedEdit)
var processedLines []string
var currentPath string
var inAppendMode bool
var indexedEdit *IndexedEdit

for _, line := range lines {
trimmed := strings.TrimSpace(line)
if strings.HasPrefix(trimmed, "#") || trimmed == "" {
processedLines = append(processedLines, line)
continue
}
line = boolTrueRegex.ReplaceAllString(line, "true")
line = boolFalseRegex.ReplaceAllString(line, "false")
trimmed = strings.TrimSpace(line)

if strings.HasPrefix(trimmed, "[[") && strings.HasSuffix(trimmed, "]]") {
inner := trimmed[2 : len(trimmed)-2]
if strings.HasPrefix(inner, "+") {
realKey := strings.TrimPrefix(inner, "+")
currentPath = realKey
inAppendMode = true
appendPaths[realKey] = true
indent := line[:strings.Index(line, "[")]
processedLines = append(processedLines, indent+"[["+realKey+"]]")
indexedEdit = nil
continue
}
if hasIndexSuffix(inner) {
arrayKey, idxStr := parseIndex(inner)
if arrayKey != "" {
idx, _ := strconv.Atoi(idxStr)
indexedEdit = &IndexedEdit{
ArrayKey: arrayKey,
Index:    idx,
Data:     make(map[string]interface{}),
}
editKey := fmt.Sprintf("%s__%d", arrayKey, idx)
indexedEdits[editKey] = indexedEdit
newLine := strings.Replace(line, inner, editKey, 1)
processedLines = append(processedLines, newLine)
currentPath = editKey
inAppendMode = false
continue
}
}
indexedEdit = nil
currentPath = inner
inAppendMode = false
processedLines = append(processedLines, line)
continue
}

if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") &&
!strings.HasPrefix(trimmed, "[[") {
inner := trimmed[1 : len(trimmed)-1]
if strings.HasPrefix(inner, "+") {
realKey := strings.TrimPrefix(inner, "+")
currentPath = realKey
inAppendMode = true
appendPaths[realKey] = true
indent := line[:strings.Index(line, "[")]
processedLines = append(processedLines, indent+"["+realKey+"]")
} else {
currentPath = inner
inAppendMode = false
processedLines = append(processedLines, line)
}
indexedEdit = nil
continue
}

if inAppendMode && currentPath != "" {
if idx := strings.Index(trimmed, "="); idx > 0 {
keyPart := strings.TrimSpace(trimmed[:idx])
fullPath := currentPath + "." + keyPart
appendPaths[fullPath] = true
}
}

if indexedEdit != nil {
if idx := strings.Index(trimmed, "="); idx > 0 {
keyPart := strings.TrimSpace(trimmed[:idx])
valPart := strings.TrimSpace(trimmed[idx+1:])
if parsedVal, ok := parseSimpleValue(valPart); ok {
indexedEdit.Data[keyPart] = parsedVal
}
}
}

processedLines = append(processedLines, line)
}

processedTOML := strings.Join(processedLines, "\n")
var data map[string]interface{}
if _, err := toml.Decode(processedTOML, &data); err != nil {
return nil, err
}

for _, edit := range indexedEdits {
editKey := fmt.Sprintf("%s__%d", edit.ArrayKey, edit.Index)
if tempData, exists := data[editKey]; exists {
if tempMap, ok := tempData.(map[string]interface{}); ok {
for k, v := range tempMap {
edit.Data[k] = v
}
}
delete(data, editKey)
}
}

return &ParsedConfig{
Data:         data,
AppendPaths:  appendPaths,
IndexedEdits: indexedEdits,
}, nil
}

func parseSimpleValue(s string) (interface{}, bool) {
s = strings.TrimSpace(s)
if s == "true" {
return true, true
}
if s == "false" {
return false, true
}
if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
return s[1 : len(s)-1], true
}
if strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'") {
return s[1 : len(s)-1], true
}
if i, err := strconv.ParseInt(s, 10, 64); err == nil {
return i, true
}
if f, err := strconv.ParseFloat(s, 64); err == nil {
return f, true
}
return nil, false
}

func hasIndexSuffix(key string) bool {
parts := strings.Split(key, ".")
if len(parts) < 2 {
return false
}
last := parts[len(parts)-1]
_, err := strconv.Atoi(last)
return err == nil
}

func parseIndex(key string) (string, string) {
parts := strings.Split(key, ".")
if len(parts) < 2 {
return "", ""
}
last := parts[len(parts)-1]
_, err := strconv.Atoi(last)
if err != nil {
return "", ""
}
return strings.Join(parts[:len(parts)-1], "."), last
}
