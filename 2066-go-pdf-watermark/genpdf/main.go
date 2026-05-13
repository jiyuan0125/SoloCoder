package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	obj1 = `1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
`
	obj2 = `2 0 obj
<<
/Type /Pages
/Count 1
/Kids [3 0 R]
>>
endobj
`
	obj3 = `3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Resources <<
/Font <<
/F1 4 0 R
>>
>>
/Contents 5 0 R
>>
endobj
`
	obj4 = `4 0 obj
<<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
endobj
`
	obj5 = `5 0 obj
<<
/Length 68
>>
stream
BT
/F1 24 Tf
100 700 Td
(Hello) Tj
ET

endstream
endobj
`
)

func main() {
	outDir := "."
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}

	header := "%PDF-1.4\n"

	offset1 := len(header)
	offset2 := offset1 + len(obj1)
	offset3 := offset2 + len(obj2)
	offset4 := offset3 + len(obj3)
	offset5 := offset4 + len(obj4)

	var body strings.Builder
	body.WriteString(header)
	body.WriteString(obj1)
	body.WriteString(obj2)
	body.WriteString(obj3)
	body.WriteString(obj4)
	body.WriteString(obj5)

	xrefOffset := body.Len()

	var xref strings.Builder
	xref.WriteString("xref\n")
	xref.WriteString("0 6\n")
	xref.WriteString(fmt.Sprintf("%010d 65535 f \n", 0))
	xref.WriteString(fmt.Sprintf("%010d 00000 n \n", offset1))
	xref.WriteString(fmt.Sprintf("%010d 00000 n \n", offset2))
	xref.WriteString(fmt.Sprintf("%010d 00000 n \n", offset3))
	xref.WriteString(fmt.Sprintf("%010d 00000 n \n", offset4))
	xref.WriteString(fmt.Sprintf("%010d 00000 n \n", offset5))

	trailer := `trailer
<<
/Size 6
/Root 1 0 R
>>
`

	var final strings.Builder
	final.WriteString(body.String())
	final.WriteString(xref.String())
	final.WriteString(trailer)
	final.WriteString("startxref\n")
	final.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	final.WriteString("%%EOF\n")

	outPath := filepath.Join(outDir, "generated.pdf")
	if err := os.WriteFile(outPath, []byte(final.String()), 0644); err != nil {
		fmt.Printf("写入失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("生成PDF: %s\n", outPath)
	fmt.Printf("Header len: %d\n", len(header))
	fmt.Printf("Obj1 offset: %d, len: %d\n", offset1, len(obj1))
	fmt.Printf("Obj2 offset: %d, len: %d\n", offset2, len(obj2))
	fmt.Printf("Obj3 offset: %d, len: %d\n", offset3, len(obj3))
	fmt.Printf("Obj4 offset: %d, len: %d\n", offset4, len(obj4))
	fmt.Printf("Obj5 offset: %d, len: %d\n", offset5, len(obj5))
	fmt.Printf("Xref offset: %d\n", xrefOffset)
}
