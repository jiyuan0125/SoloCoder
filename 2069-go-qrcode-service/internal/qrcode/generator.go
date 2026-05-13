package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"strings"

	qrcodelib "github.com/skip2/go-qrcode"
	"qrcode-service/internal/types"
	"qrcode-service/internal/utils"
)

func GenerateQR(req *types.QRCodeRequest) ([]byte, error) {
	level := utils.ParseErrorLevel(req.ErrorLevel)
	
	if req.LogoFile != nil && len(req.LogoFile) > 0 {
		if level < 3 {
			level = 3
		}
	}

	content := req.Content
	
	q, err := qrcodelib.New(content, qrcodelib.RecoveryLevel(level))
	if err != nil {
		return nil, err
	}

	foreground, err := utils.ParseColor(req.Foreground)
	if err != nil {
		return nil, fmt.Errorf("invalid foreground color: %v", err)
	}
	background, err := utils.ParseColor(req.Background)
	if err != nil {
		return nil, fmt.Errorf("invalid background color: %v", err)
	}
	q.ForegroundColor = foreground
	q.BackgroundColor = background

	baseSize := req.Size
	if baseSize <= 0 {
		baseSize = 256
	}

	format := strings.ToUpper(req.Format)
	
	if format == "SVG" {
		return generateSVG(q, baseSize, req.Border, req.Title, foreground, background)
	}

	img := q.Image(baseSize)
	
	if req.LogoFile != nil && len(req.LogoFile) > 0 {
		logoImg, err := decodeImage(req.LogoFile, req.LogoFilename)
		if err != nil {
			return nil, err
		}
		img = overlayLogo(img, logoImg)
	}

	if req.Title != "" || req.Border > 0 {
		img = addTitleAndBorder(img, req.Title, req.Border, foreground, background)
	}

	var buf bytes.Buffer
	switch format {
	case "PNG":
		err = png.Encode(&buf, img)
	case "JPEG", "JPG":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decodeImage(data []byte, filename string) (image.Image, error) {
	reader := bytes.NewReader(data)
	if strings.HasSuffix(strings.ToLower(filename), ".png") {
		return png.Decode(reader)
	}
	return jpeg.Decode(reader)
}

func overlayLogo(qrImg image.Image, logoImg image.Image) image.Image {
	qrBounds := qrImg.Bounds()
	qrSize := qrBounds.Dx()
	
	logoMaxSize := int(float64(qrSize) * 0.25)
	
	logoBounds := logoImg.Bounds()
	logoW := logoBounds.Dx()
	logoH := logoBounds.Dy()
	
	var newLogoW, newLogoH int
	if logoW > logoH {
		newLogoW = logoMaxSize
		newLogoH = int(float64(logoH) * float64(logoMaxSize) / float64(logoW))
	} else {
		newLogoH = logoMaxSize
		newLogoW = int(float64(logoW) * float64(logoMaxSize) / float64(logoH))
	}
	
	resizedLogo := resizeImage(logoImg, newLogoW, newLogoH)
	
	result := image.NewRGBA(qrBounds)
	draw.Draw(result, qrBounds, qrImg, image.Point{}, draw.Src)
	
	offsetX := (qrSize - newLogoW) / 2
	offsetY := (qrSize - newLogoH) / 2
	logoPos := image.Rect(offsetX, offsetY, offsetX+newLogoW, offsetY+newLogoH)
	draw.Draw(result, logoPos, resizedLogo, image.Point{}, draw.Over)
	
	return result
}

func resizeImage(img image.Image, w, h int) image.Image {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := int(math.Round(float64(x) * float64(srcW) / float64(w)))
			srcY := int(math.Round(float64(y) * float64(srcH) / float64(h)))
			if srcX >= srcW {
				srcX = srcW - 1
			}
			if srcY >= srcH {
				srcY = srcH - 1
			}
			dst.Set(x, y, img.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}
	
	return dst
}

func addTitleAndBorder(img image.Image, title string, border int, fg, bg color.Color) image.Image {
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	
	titleHeight := 0
	if title != "" {
		titleHeight = 40
	}
	if border < 0 {
		border = 0
	}
	
	newW := srcW + 2*border
	newH := srcH + 2*border + titleHeight
	
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	bgRect := &image.Uniform{C: bg}
	draw.Draw(dst, dst.Bounds(), bgRect, image.Point{}, draw.Src)
	
	qrPos := image.Rect(border, border, border+srcW, border+srcH)
	draw.Draw(dst, qrPos, img, image.Point{}, draw.Src)
	
	return dst
}

func generateSVG(q *qrcodelib.QRCode, size int, border int, title string, fg, bg color.Color) ([]byte, error) {
	var buf bytes.Buffer
	
	bitmap := q.Bitmap()
	scale := float64(size) / float64(len(bitmap))
	
	if border < 0 {
		border = 0
	}
	
	titleHeight := 0
	if title != "" {
		titleHeight = 40
	}
	
	totalW := size + 2*border
	totalH := size + 2*border + titleHeight
	
	fgHex := colorToHex(fg)
	bgHex := colorToHex(bg)
	
	buf.WriteString(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">
  <rect width="%d" height="%d" fill="%s"/>
`, totalW, totalH, totalW, totalH, totalW, totalH, bgHex))
	
	for y, row := range bitmap {
		for x, on := range row {
			if on {
				px := int(math.Round(float64(x)*scale)) + border
				py := int(math.Round(float64(y)*scale)) + border
				pw := int(math.Ceil(scale))
				ph := int(math.Ceil(scale))
				buf.WriteString(fmt.Sprintf(`  <rect x="%d" y="%d" width="%d" height="%d" fill="%s"/>
`, px, py, pw, ph, fgHex))
			}
		}
	}
	
	buf.WriteString(`</svg>`)
	
	return buf.Bytes(), nil
}

func colorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02X%02X%02X", r>>8, g>>8, b>>8)
}
