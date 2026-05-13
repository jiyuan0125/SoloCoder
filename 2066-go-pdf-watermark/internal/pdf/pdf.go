package pdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"pdf-watermark/internal/types"
)

type PDF struct {
	raw       []byte
	pageCount int
	encrypted bool
	pages     []pageRef
	xrefs     map[int]int
	trailer   map[string]interface{}
	root      int
	size      int
}

type pageRef struct {
	offset int
	refNum int
}

func NewPDF(path string) (*PDF, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &PDF{raw: data, xrefs: make(map[int]int)}

	if len(data) < 8 || !bytes.HasPrefix(data, []byte("%PDF-")) {
		return nil, fmt.Errorf("不是有效的PDF文件")
	}

	if err := p.parse(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PDF) parse() error {
	startxref, err := p.findStartXref()
	if err != nil {
		return fmt.Errorf("解析PDF失败: %v (位置: 无法找到交叉引用)", err)
	}

	xref, size, trailer, err := p.readXref(startxref)
	if err != nil {
		return fmt.Errorf("解析PDF失败: %v (位置: 交叉引用表)", err)
	}

	p.xrefs = xref
	p.trailer = trailer
	p.size = size

	rootObj, ok := trailer["Root"]
	if !ok {
		return fmt.Errorf("解析PDF失败: 缺少Root (位置: trailer字典)")
	}

	rootRef, ok := rootObj.([]interface{})
	if !ok || len(rootRef) != 2 {
		return fmt.Errorf("解析PDF失败: 无效的Root引用")
	}
	p.root = rootRef[0].(int)

	if enc, ok := trailer["Encrypt"]; ok && enc != nil {
		p.encrypted = true
	}

	pageTree, err := p.getObject(p.root)
	if err != nil {
		return fmt.Errorf("解析PDF失败: %v (位置: Catalog)", err)
	}

	catalogDict, ok := pageTree.(map[string]interface{})
	if !ok {
		return fmt.Errorf("解析PDF失败: 无效的Catalog")
	}

	pagesRef, ok := catalogDict["Pages"]
	if !ok {
		return fmt.Errorf("解析PDF失败: Catalog缺少Pages")
	}

	var pagesObj interface{}
	if ref, ok := pagesRef.([]interface{}); ok && len(ref) == 2 {
		pagesObj, err = p.getObject(ref[0].(int))
		if err != nil {
			return fmt.Errorf("解析PDF失败: %v (位置: Pages节点)", err)
		}
	} else {
		return fmt.Errorf("解析PDF失败: 无效的Pages引用")
	}

	treeDict, ok := pagesObj.(map[string]interface{})
	if !ok {
		return fmt.Errorf("解析PDF失败: 无效的Pages对象")
	}

	pages, err := p.collectPages(treeDict)
	if err != nil {
		return fmt.Errorf("解析PDF失败: %v (位置: 页面树)", err)
	}
	p.pages = pages
	p.pageCount = len(pages)

	return nil
}

func (p *PDF) findStartXref() (int, error) {
	idx := bytes.LastIndex(p.raw, []byte("startxref"))
	if idx == -1 {
		return 0, fmt.Errorf("找不到startxref")
	}
	rest := p.raw[idx:]
	parts := strings.Fields(string(rest))
	if len(parts) < 2 {
		return 0, fmt.Errorf("无效的startxref")
	}
	offset, _ := strconv.Atoi(parts[1])
	return offset, nil
}

func (p *PDF) readXref(start int) (map[int]int, int, map[string]interface{}, error) {
	xrefs := make(map[int]int)

	if !bytes.HasPrefix(p.raw[start:], []byte("xref")) {
		xrefObj, err := p.readObjectAt(start)
		if err != nil {
			return nil, 0, nil, err
		}
		xrefDict, ok := xrefObj.(map[string]interface{})
		if !ok {
			return nil, 0, nil, fmt.Errorf("无效的交叉引用流")
		}

		size, _ := xrefDict["Size"].(int)

		if wArr, ok := xrefDict["W"].([]interface{}); ok {
			if stream, ok := xrefDict["_stream"].([]byte); ok {
				w1, _ := wArr[0].(int)
				w2, _ := wArr[1].(int)
				w3, _ := wArr[2].(int)
				entrySize := w1 + w2 + w3

				if dec, err := decodeStream(stream, xrefDict); err == nil {
					stream = dec
				}

				idxArr, ok := xrefDict["Index"].([]interface{})
				var firstObj, count int
				if ok && len(idxArr) >= 2 {
					firstObj, _ = idxArr[0].(int)
					count, _ = idxArr[1].(int)
				} else {
					firstObj = 0
					count = size
				}

				for i := 0; i < count && i*entrySize+entrySize <= len(stream); i++ {
					entry := stream[i*entrySize : (i+1)*entrySize]
					typ := 0
					if w1 > 0 {
						typ = int(readBytesAsInt(entry[:w1]))
					}
					field2 := readBytesAsInt(entry[w1 : w1+w2])
					objNum := firstObj + i

					if typ == 1 {
						xrefs[objNum] = int(field2)
					}
				}
			}
		}

		return xrefs, size, xrefDict, nil
	}

	lines := strings.Split(string(p.raw[start:]), "\n")
	i := 1
	totalSize := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}
		if line == "trailer" {
			break
		}

		parts := strings.Fields(line)
		if len(parts) == 2 {
			first, _ := strconv.Atoi(parts[0])
			count, _ := strconv.Atoi(parts[1])
			totalSize = first + count

			i++
			for j := 0; j < count && i < len(lines); j++ {
				line = strings.TrimSpace(lines[i])
				parts := strings.Fields(line)
				if len(parts) == 3 {
					offset, _ := strconv.Atoi(parts[0])
					if parts[2] == "n" {
						xrefs[first+j] = offset
					}
				}
				i++
			}
		} else {
			i++
		}
	}

	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "<<" {
			break
		}
	}

	var trailer map[string]interface{}
	if i < len(lines) {
		trailerStart := bytes.Index(p.raw[start:], []byte("<<"))
		if trailerStart != -1 {
			if dict, _, err := p.parseDict(p.raw[start+trailerStart:]); err == nil {
				trailer, _ = dict.(map[string]interface{})
			}
		}
	}
	if trailer == nil {
		trailer = make(map[string]interface{})
	}

	return xrefs, totalSize, trailer, nil
}

func readBytesAsInt(b []byte) int64 {
	var v int64
	for _, bb := range b {
		v = v<<8 | int64(bb)
	}
	return v
}

func (p *PDF) collectPages(tree map[string]interface{}) ([]pageRef, error) {
	typ, _ := tree["Type"].(string)
	kids, _ := tree["Kids"].([]interface{})

	if typ == "Pages" || (typ == "" && len(kids) > 0) {
		var pages []pageRef
		for _, kid := range kids {
			if ref, ok := kid.([]interface{}); ok && len(ref) == 2 {
				objNum := ref[0].(int)
				obj, err := p.getObject(objNum)
				if err != nil {
					continue
				}
				dict, ok := obj.(map[string]interface{})
				if !ok {
					continue
				}
				kidType, _ := dict["Type"].(string)
				if kidType == "Page" {
					pages = append(pages, pageRef{refNum: objNum})
				} else {
					subPages, err := p.collectPages(dict)
					if err != nil {
						continue
					}
					pages = append(pages, subPages...)
				}
			}
		}
		return pages, nil
	}

	return []pageRef{}, nil
}

func (p *PDF) getObject(num int) (interface{}, error) {
	offset, ok := p.xrefs[num]
	if !ok {
		return nil, fmt.Errorf("对象 %d 未找到", num)
	}
	return p.readObjectAt(offset)
}

func (p *PDF) readObjectAt(offset int) (interface{}, error) {
	if offset < 0 || offset >= len(p.raw) {
		return nil, fmt.Errorf("无效的偏移量")
	}

	idx := offset
	for idx < len(p.raw) && (p.raw[idx] == '\n' || p.raw[idx] == '\r' || p.raw[idx] == ' ') {
		idx++
	}

	rest := p.raw[idx:]
	parts := strings.Fields(string(rest))
	if len(parts) < 3 || parts[2] != "obj" {
		return nil, fmt.Errorf("无效的对象偏移")
	}

	startIdx := idx
	for i := 0; i < 3 && startIdx < len(p.raw); i++ {
		for startIdx < len(p.raw) && p.raw[startIdx] != ' ' && p.raw[startIdx] != '\n' && p.raw[startIdx] != '\r' {
			startIdx++
		}
		for startIdx < len(p.raw) && (p.raw[startIdx] == ' ' || p.raw[startIdx] == '\n' || p.raw[startIdx] == '\r') {
			startIdx++
		}
	}

	obj, consumed, err := p.parseValue(p.raw[startIdx:])
	if err != nil {
		return nil, err
	}

	if dict, ok := obj.(map[string]interface{}); ok {
		streamStart := startIdx + consumed
		streamIdx := bytes.Index(p.raw[streamStart:], []byte("stream"))
		if streamIdx != -1 {
			streamIdx += streamStart
			dataStart := streamIdx + 6
			for dataStart < len(p.raw) && (p.raw[dataStart] == '\n' || p.raw[dataStart] == '\r') {
				dataStart++
			}

			endstream := bytes.Index(p.raw[dataStart:], []byte("endstream"))
			if endstream != -1 {
				endIdx := dataStart + endstream
				for endIdx > dataStart && (p.raw[endIdx-1] == '\n' || p.raw[endIdx-1] == '\r') {
					endIdx--
				}
				dict["_stream"] = p.raw[dataStart:endIdx]
			}
		}
	}

	return obj, nil
}

func (p *PDF) parseValue(data []byte) (interface{}, int, error) {
	i := 0
	for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
		i++
	}

	if i >= len(data) {
		return nil, i, fmt.Errorf("空值")
	}

	switch data[i] {
	case '(':
		return p.parseString(data[i:])
	case '/':
		return p.parseName(data[i:])
	case '[':
		return p.parseArray(data[i:])
	case '<':
		if i+1 < len(data) && data[i+1] == '<' {
			return p.parseDict(data[i:])
		}
		return p.parseHexString(data[i:])
	case 'n':
		if bytes.HasPrefix(data[i:], []byte("null")) {
			return nil, i + 4, nil
		}
	case 't':
		if bytes.HasPrefix(data[i:], []byte("true")) {
			return true, i + 4, nil
		}
	case 'f':
		if bytes.HasPrefix(data[i:], []byte("false")) {
			return false, i + 5, nil
		}
	}

	if (data[i] >= '0' && data[i] <= '9') || data[i] == '-' || data[i] == '+' || data[i] == '.' {
		return p.parseNumberOrRef(data[i:])
	}

	return nil, i, fmt.Errorf("未知类型: %c", data[i])
}

func (p *PDF) parseString(data []byte) (interface{}, int, error) {
	i := 1
	depth := 1
	var result bytes.Buffer

	for i < len(data) && depth > 0 {
		switch data[i] {
		case '\\':
			if i+1 < len(data) {
				result.WriteByte(data[i])
				result.WriteByte(data[i+1])
				i += 2
			} else {
				i++
			}
		case '(':
			depth++
			result.WriteByte(data[i])
			i++
		case ')':
			depth--
			if depth > 0 {
				result.WriteByte(data[i])
			}
			i++
		default:
			result.WriteByte(data[i])
			i++
		}
	}

	return result.String(), i, nil
}

func (p *PDF) parseHexString(data []byte) (interface{}, int, error) {
	i := 1
	var result bytes.Buffer

	for i < len(data) && data[i] != '>' {
		if data[i] != ' ' && data[i] != '\n' && data[i] != '\r' && data[i] != '\t' {
			result.WriteByte(data[i])
		}
		i++
	}
	i++

	return result.String(), i, nil
}

func (p *PDF) parseName(data []byte) (interface{}, int, error) {
	i := 1
	var result bytes.Buffer

	for i < len(data) {
		c := data[i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '/' || c == '[' || c == ']' || c == '(' || c == ')' || c == '<' || c == '>' {
			break
		}
		result.WriteByte(c)
		i++
	}

	return result.String(), i, nil
}

func (p *PDF) parseArray(data []byte) (interface{}, int, error) {
	i := 1
	var result []interface{}

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i >= len(data) || data[i] == ']' {
			i++
			break
		}

		val, consumed, err := p.parseValue(data[i:])
		if err != nil {
			return nil, i, err
		}
		result = append(result, val)
		i += consumed
	}

	return result, i, nil
}

func (p *PDF) parseDict(data []byte) (interface{}, int, error) {
	i := 2
	result := make(map[string]interface{})

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i+1 < len(data) && data[i] == '>' && data[i+1] == '>' {
			i += 2
			break
		}
		if i >= len(data) {
			break
		}

		if data[i] != '/' {
			i++
			continue
		}

		key, keyConsumed, err := p.parseName(data[i:])
		if err != nil {
			return nil, i, err
		}
		i += keyConsumed

		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}

		val, valConsumed, err := p.parseValue(data[i:])
		if err != nil {
			return nil, i, err
		}
		i += valConsumed

		result[key.(string)] = val
	}

	return result, i, nil
}

func (p *PDF) parseNumberOrRef(data []byte) (interface{}, int, error) {
	i := 0
	var numStr bytes.Buffer

	for i < len(data) && ((data[i] >= '0' && data[i] <= '9') || data[i] == '-' || data[i] == '+' || data[i] == '.' || data[i] == 'e' || data[i] == 'E') {
		numStr.WriteByte(data[i])
		i++
	}

	if numStr.Len() == 0 {
		return nil, 0, fmt.Errorf("无效的数字")
	}

	num1Str := numStr.String()

	for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
		i++
	}

	j := i
	var num2Str bytes.Buffer
	for j < len(data) && data[j] >= '0' && data[j] <= '9' {
		num2Str.WriteByte(data[j])
		j++
	}

	if num2Str.Len() > 0 {
		for j < len(data) && (data[j] == ' ' || data[j] == '\n' || data[j] == '\r' || data[j] == '\t') {
			j++
		}
		if bytes.HasPrefix(data[j:], []byte("R")) {
			num1, _ := strconv.Atoi(num1Str)
			num2, _ := strconv.Atoi(num2Str.String())
			return []interface{}{num1, num2}, j + 1, nil
		}
	}

	if strings.Contains(num1Str, ".") {
		f, _ := strconv.ParseFloat(num1Str, 64)
		return f, i, nil
	}
	n, _ := strconv.Atoi(num1Str)
	return n, i, nil
}

func (p *PDF) parseDictFromLines(lines []string, startIdx int) (map[string]interface{}, int, error) {
	result := make(map[string]interface{})
	i := startIdx

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == ">>" || line == "trailer" {
			break
		}

		if strings.HasPrefix(line, "/") {
			parts := strings.SplitN(line[1:], " ", 2)
			key := parts[0]
			if len(parts) > 1 {
				val := strings.TrimSpace(parts[1])
				if strings.HasPrefix(val, "[") {
					result[key] = []interface{}{val}
				} else if strings.HasPrefix(val, "<<") {
					depth := 0
					buf := new(bytes.Buffer)
					j := i
					for j < len(lines) {
						l := lines[j]
						if strings.Contains(l, "<<") {
							depth += strings.Count(l, "<<")
						}
						if strings.Contains(l, ">>") {
							depth -= strings.Count(l, ">>")
						}
						buf.WriteString(l)
						buf.WriteString("\n")
						if depth == 0 {
							break
						}
						j++
					}
					result[key] = map[string]interface{}{"_raw": buf.String()}
					i = j
				} else {
					result[key] = val
				}
			}
		}
		i++
	}

	return result, i, nil
}

func decodeStream(data []byte, dict map[string]interface{}) ([]byte, error) {
	filter, _ := dict["Filter"].(string)
	if filter == "FlateDecode" || filter == "Flate" {
		r, err := zlib.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}
	return nil, fmt.Errorf("不支持的压缩格式: %s", filter)
}

func (p *PDF) IsEncrypted() bool {
	return p.encrypted
}

func (p *PDF) PageCount() int {
	return p.pageCount
}

func (p *PDF) AddWatermark(outputPath string, config *types.WatermarkConfig) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	xrefOffsets := make(map[int]int)

	header := []byte("%PDF-1.7\n%\xE2\xE3\xCF\xD3\n")
	out.Write(header)
	offset := len(header)

	newObjNum := p.size
	newObjects := make(map[int][]byte)

	watermarkStream, err := p.createWatermarkStream(config)
	if err != nil {
		return err
	}
	wmStreamObj := fmt.Sprintf("%d 0 obj\n<<\n  /Length %d\n>>\nstream\n%s\nendstream\nendobj\n",
		newObjNum, len(watermarkStream), watermarkStream)
	newObjects[newObjNum] = []byte(wmStreamObj)
	newObjNum++

	wmXObjDict := fmt.Sprintf("%d 0 obj\n<<\n  /Type /XObject\n  /Subtype /Form\n  /BBox [0 0 612 792]\n  /Matrix [1 0 0 1 0 0]\n  /Resources <<\n    /ProcSet [/PDF /Text]\n  >>\n  /Length %d 0 R\n>>\nendobj\n",
		newObjNum, newObjNum-1)
	newObjects[newObjNum] = []byte(wmXObjDict)
	newObjNum++

	for i := 0; i < len(p.raw); {
		objStart := bytes.Index(p.raw[i:], []byte(" obj\n"))
		if objStart == -1 {
			break
		}

		lineStart := bytes.LastIndex(p.raw[:i+objStart], []byte("\n"))
		if lineStart == -1 {
			lineStart = 0
		} else {
			lineStart++
		}

		headerLine := string(p.raw[lineStart : i+objStart])
		parts := strings.Fields(headerLine)
		if len(parts) < 2 {
			i = i + objStart + 5
			continue
		}

		objNum, _ := strconv.Atoi(parts[0])
		xrefOffsets[objNum] = offset

		endObj := bytes.Index(p.raw[i+objStart:], []byte("endobj"))
		if endObj == -1 {
			break
		}

		objData := p.raw[lineStart : i+objStart+endObj+6]

		isPage := false
		for _, page := range p.pages {
			if page.refNum == objNum {
				isPage = true
				break
			}
		}

		if isPage {
			modData := p.modifyPageContent(objData, newObjNum-1)
			out.Write(modData)
			offset += len(modData)
		} else {
			out.Write(objData)
			offset += len(objData)
		}

		i = i + objStart + endObj + 6
	}

	for num, objData := range newObjects {
		xrefOffsets[num] = offset
		out.Write(objData)
		offset += len(objData)
	}

	xrefStart := offset

	out.WriteString("xref\n")
	maxNum := newObjNum
	for num := range xrefOffsets {
		if num > maxNum {
			maxNum = num
		}
	}
	out.WriteString(fmt.Sprintf("0 %d\n", maxNum+1))
	out.WriteString("0000000000 65535 f \n")
	for i := 1; i <= maxNum; i++ {
		if off, ok := xrefOffsets[i]; ok {
			out.WriteString(fmt.Sprintf("%010d 00000 n \n", off))
		} else {
			out.WriteString("0000000000 00000 f \n")
		}
	}

	out.WriteString("trailer\n")
	out.WriteString("<<\n")
	out.WriteString(fmt.Sprintf("  /Size %d\n", maxNum+1))
	out.WriteString(fmt.Sprintf("  /Root %d 0 R\n", p.root))
	if info, ok := p.trailer["Info"]; ok {
		if ref, ok := info.([]interface{}); ok && len(ref) == 2 {
			out.WriteString(fmt.Sprintf("  /Info %d 0 R\n", ref[0]))
		}
	}
	if id, ok := p.trailer["ID"]; ok {
		if arr, ok := id.([]interface{}); ok && len(arr) == 2 {
			out.WriteString(fmt.Sprintf("  /ID [%s %s]\n", arr[0], arr[1]))
		}
	}
	out.WriteString(">>\n")
	out.WriteString("startxref\n")
	out.WriteString(fmt.Sprintf("%d\n", xrefStart))
	out.WriteString("%%EOF\n")

	return nil
}

func (p *PDF) modifyPageContent(data []byte, wmObjNum int) []byte {
	pageStr := string(data)

	contentsRe := regexp.MustCompile(`/Contents\s+(\d+\s+\d+\s+R)`)
	matches := contentsRe.FindStringSubmatch(pageStr)

	if len(matches) > 0 {
		origContents := matches[1]
		newContents := fmt.Sprintf("[ %s %d 0 R ]", origContents, wmObjNum)
		pageStr = strings.Replace(pageStr, matches[0], "/Contents "+newContents, 1)
	} else {
		dictEnd := strings.Index(pageStr, ">>")
		if dictEnd != -1 {
			pageStr = pageStr[:dictEnd] + fmt.Sprintf("\n  /Contents %d 0 R\n", wmObjNum) + pageStr[dictEnd:]
		}
	}

	return []byte(pageStr)
}

func (p *PDF) createWatermarkStream(config *types.WatermarkConfig) (string, error) {
	if config.Mode == "image" {
		return p.createImageWatermarkStream(config)
	}
	return p.createTextWatermarkStream(config)
}

func (p *PDF) createTextWatermarkStream(config *types.WatermarkConfig) (string, error) {
	var buf bytes.Buffer

	pageWidth := 612.0
	pageHeight := 792.0
	text := config.Text
	fontSize := config.FontSize
	if fontSize <= 0 {
		fontSize = 24
	}

	rgb := parseRGB(config.Color)
	opacity := config.Opacity
	if opacity <= 0 || opacity > 1 {
		opacity = 0.5
	}
	rotation := config.Rotation * math.Pi / 180

	textWidth := fontSize * float64(len(text)) * 0.6
	textHeight := fontSize

	var x, y float64

	switch config.Position {
	case "top-left":
		x = 50
		y = pageHeight - 50 - textHeight
	case "top-right":
		x = pageWidth - 50 - textWidth
		y = pageHeight - 50 - textHeight
	case "bottom-left":
		x = 50
		y = 50
	case "bottom-right":
		x = pageWidth - 50 - textWidth
		y = 50
	case "tile":
		return p.createTileTextWatermark(config, text, fontSize, rgb, opacity, rotation)
	default:
		x = (pageWidth - textWidth) / 2
		y = (pageHeight - textHeight) / 2
	}

	buf.WriteString("q\n")
	buf.WriteString(fmt.Sprintf("/DeviceRGB cs\n"))
	buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f scn\n", rgb[0], rgb[1], rgb[2]))
	buf.WriteString(fmt.Sprintf("%.3f gs\n", opacity))
	buf.WriteString("BT\n")
	buf.WriteString(fmt.Sprintf("/F1 %.0f Tf\n", fontSize))

	if rotation != 0 {
		cos := math.Cos(rotation)
		sin := math.Sin(rotation)
		cx := x + textWidth/2
		cy := y + textHeight/2
		buf.WriteString(fmt.Sprintf("1 0 0 1 %.3f %.3f Tm\n", cx, cy))
		buf.WriteString(fmt.Sprintf("%.5f %.5f %.5f %.5f %.3f %.3f Tm\n",
			cos, sin, -sin, cos, -cx, -cy))
	}
	buf.WriteString(fmt.Sprintf("%.3f %.3f Td\n", x, y))

	escaped := escapePDFString(text)
	buf.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
	buf.WriteString("ET\n")
	buf.WriteString("Q\n")

	return buf.String(), nil
}

func (p *PDF) createTileTextWatermark(config *types.WatermarkConfig, text string, fontSize float64, rgb []float64, opacity float64, rotation float64) (string, error) {
	var buf bytes.Buffer

	pageWidth := 612.0
	pageHeight := 792.0
	textWidth := fontSize * float64(len(text)) * 0.6
	textHeight := fontSize * 1.5

	gapX := textWidth + 100
	gapY := textHeight + 100

	buf.WriteString("q\n")
	buf.WriteString(fmt.Sprintf("/DeviceRGB cs\n"))
	buf.WriteString(fmt.Sprintf("%.3f %.3f %.3f scn\n", rgb[0], rgb[1], rgb[2]))
	buf.WriteString(fmt.Sprintf("%.3f gs\n", opacity))

	if rotation != 0 {
		cos := math.Cos(rotation)
		sin := math.Sin(rotation)
		buf.WriteString(fmt.Sprintf("%.5f %.5f %.5f %.5f 0 0 cm\n", cos, sin, -sin, cos))
	}

	for y := -pageHeight; y < pageHeight*2; y += gapY {
		offset := (y / gapY) * (gapX / 2)
		for x := -pageWidth + offset; x < pageWidth*2; x += gapX {
			buf.WriteString("BT\n")
			buf.WriteString(fmt.Sprintf("/F1 %.0f Tf\n", fontSize))
			buf.WriteString(fmt.Sprintf("%.3f %.3f Td\n", x, y))
			escaped := escapePDFString(text)
			buf.WriteString(fmt.Sprintf("(%s) Tj\n", escaped))
			buf.WriteString("ET\n")
		}
	}

	buf.WriteString("Q\n")
	return buf.String(), nil
}

func (p *PDF) createImageWatermarkStream(config *types.WatermarkConfig) (string, error) {
	f, err := os.Open(config.ImagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var img image.Image
	imgName := strings.ToLower(config.ImagePath)
	if strings.HasSuffix(imgName, ".png") {
		img, err = png.Decode(f)
	} else {
		img, err = jpeg.Decode(f)
	}
	if err != nil {
		return "", fmt.Errorf("无法解码图片: %v", err)
	}

	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	pageWidth := 612.0
	pageHeight := 792.0

	scale := 0.3
	wmWidth := float64(imgWidth) * scale
	wmHeight := float64(imgHeight) * scale

	var x, y float64

	switch config.Position {
	case "top-left":
		x = 50
		y = pageHeight - 50 - wmHeight
	case "top-right":
		x = pageWidth - 50 - wmWidth
		y = pageHeight - 50 - wmHeight
	case "bottom-left":
		x = 50
		y = 50
	case "bottom-right":
		x = pageWidth - 50 - wmWidth
		y = 50
	case "tile":
		return p.createTileImageWatermark(config, imgWidth, imgHeight, scale)
	default:
		x = (pageWidth - wmWidth) / 2
		y = (pageHeight - wmHeight) / 2
	}

	var buf bytes.Buffer
	buf.WriteString("q\n")
	buf.WriteString(fmt.Sprintf("%.2f 0 0 %.2f %.2f %.2f cm\n", wmWidth, wmHeight, x, y))
	buf.WriteString("/Img0 Do\n")
	buf.WriteString("Q\n")

	return buf.String(), nil
}

func (p *PDF) createTileImageWatermark(config *types.WatermarkConfig, imgWidth, imgHeight int, scale float64) (string, error) {
	var buf bytes.Buffer

	pageWidth := 612.0
	pageHeight := 792.0
	wmWidth := float64(imgWidth) * scale
	wmHeight := float64(imgHeight) * scale

	gap := 50.0

	buf.WriteString("q\n")

	for y := 0.0; y < pageHeight+wmHeight; y += wmHeight + gap {
		for x := 0.0; x < pageWidth+wmWidth; x += wmWidth + gap {
			buf.WriteString("q\n")
			buf.WriteString(fmt.Sprintf("%.2f 0 0 %.2f %.2f %.2f cm\n", wmWidth, wmHeight, x, y))
			buf.WriteString("/Img0 Do\n")
			buf.WriteString("Q\n")
		}
	}

	buf.WriteString("Q\n")
	return buf.String(), nil
}

func parseRGB(color string) []float64 {
	parts := strings.Split(color, ",")
	result := []float64{0, 0, 0}
	for i, p := range parts {
		if i >= 3 {
			break
		}
		if v, err := strconv.ParseFloat(strings.TrimSpace(p), 64); err == nil {
			result[i] = v / 255.0
		}
	}
	return result
}

func escapePDFString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	s = strings.ReplaceAll(s, ")", "\\)")
	return s
}
