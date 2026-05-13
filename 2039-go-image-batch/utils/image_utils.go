package utils

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	bmp "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"

	webp "github.com/chai2010/webp"
)

var supportedFormats = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".bmp":  true,
	".webp": true,
}

var supportedFormatNames = []string{"JPEG", "PNG", "GIF", "BMP", "WebP"}

func GetSupportedFormats() string {
	return strings.Join(supportedFormatNames, ", ")
}

func GetSupportedFormatList() []string {
	return supportedFormatNames
}

func IsSupportedFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return supportedFormats[ext]
}

func GetImageFormat(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "jpeg"
	case ".png":
		return "png"
	case ".gif":
		return "gif"
	case ".bmp":
		return "bmp"
	case ".webp":
		return "webp"
	}
	return ""
}

func ValidateImageFile(file *os.File) error {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, format, err := image.Decode(file)
	if err != nil {
		return errors.New("invalid or corrupted image file: " + err.Error())
	}

	format = strings.ToLower(format)
	if format != "jpeg" && format != "png" && format != "gif" &&
		format != "bmp" && format != "webp" {
		return errors.New("unsupported image format: " + format)
	}

	_, err = file.Seek(0, io.SeekStart)
	return err
}

func DecodeImageFromFile(file *os.File) (image.Image, string, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, "", err
	}

	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}

	return img, strings.ToLower(format), nil
}

func DecodeGifFromFile(file *os.File) (*gif.GIF, error) {
	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return gif.DecodeAll(file)
}

func EncodeImage(w io.Writer, img image.Image, format string, quality int) error {
	if quality <= 0 {
		quality = 80
	}

	switch format {
	case "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	case "png":
		return png.Encode(w, img)
	case "bmp":
		return bmp.Encode(w, img)
	case "webp":
		return webp.Encode(w, img, &webp.Options{Lossless: false, Quality: float32(quality)})
	case "gif":
		return gif.Encode(w, img, &gif.Options{NumColors: 256})
	default:
		return errors.New("unsupported output format: " + format)
	}
}

func EncodeGif(w io.Writer, g *gif.GIF) error {
	return gif.EncodeAll(w, g)
}

func GetExtensionFromFormat(format string) string {
	switch format {
	case "jpeg":
		return ".jpg"
	case "png":
		return ".png"
	case "gif":
		return ".gif"
	case "bmp":
		return ".bmp"
	case "webp":
		return ".webp"
	}
	return ""
}
