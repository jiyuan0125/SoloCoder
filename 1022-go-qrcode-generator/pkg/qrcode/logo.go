package qrcode

import (
	"errors"
	"image"
	"image/draw"
	"math"

	"github.com/qrcode/generator/pkg/common"
)

func validateAndResizeLogo(qrSize int, logo image.Image) (image.Image, error) {
	qrArea := float64(qrSize * qrSize)
	maxLogoArea := qrArea * common.MaxLogoRatio

	logoBounds := logo.Bounds()
	logoWidth := logoBounds.Dx()
	logoHeight := logoBounds.Dy()
	logoArea := float64(logoWidth * logoHeight)

	if logoArea > maxLogoArea {
		scale := math.Sqrt(maxLogoArea / logoArea)
		newWidth := int(float64(logoWidth) * scale)
		newHeight := int(float64(logoHeight) * scale)
		logo = resizeImage(logo, newWidth, newHeight)
	}

	if logo.Bounds().Dx() > qrSize || logo.Bounds().Dy() > qrSize {
		return nil, errors.New("logo is too large even after resizing")
	}

	return logo, nil
}

func overlayLogo(qrImg image.Image, logo image.Image) (image.Image, error) {
	qrBounds := qrImg.Bounds()
	qrWidth := qrBounds.Dx()
	qrHeight := qrBounds.Dy()

	logo, err := validateAndResizeLogo(qrWidth, logo)
	if err != nil {
		return nil, err
	}

	logoBounds := logo.Bounds()
	logoWidth := logoBounds.Dx()
	logoHeight := logoBounds.Dy()

	offsetX := (qrWidth - logoWidth) / 2
	offsetY := (qrHeight - logoHeight) / 2

	result := image.NewRGBA(qrBounds)
	draw.Draw(result, qrBounds, qrImg, image.Point{}, draw.Src)

	logoRect := image.Rect(offsetX, offsetY, offsetX+logoWidth, offsetY+logoHeight)
	draw.Draw(result, logoRect, logo, logoBounds.Min, draw.Over)

	return result, nil
}

func resizeImage(img image.Image, newWidth, newHeight int) image.Image {
	srcBounds := img.Bounds()
	srcWidth := srcBounds.Dx()
	srcHeight := srcBounds.Dy()

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	scaleX := float64(srcWidth) / float64(newWidth)
	scaleY := float64(srcHeight) / float64(newHeight)

	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			if srcX >= srcWidth {
				srcX = srcWidth - 1
			}
			if srcY >= srcHeight {
				srcY = srcHeight - 1
			}
			dst.Set(x, y, img.At(srcBounds.Min.X+srcX, srcBounds.Min.Y+srcY))
		}
	}

	return dst
}
