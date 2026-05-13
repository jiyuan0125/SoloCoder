package services

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image-batch/config"
	"image-batch/models"
	"image-batch/utils"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageService struct {
}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) ValidateOperations(ops []models.ImageOperation) error {
	if len(ops) == 0 {
		return errors.New("at least one operation is required")
	}

	for _, op := range ops {
		switch op.Type {
		case "resize":
			if op.Width <= 0 || op.Height <= 0 {
				return errors.New("resize operation requires positive width and height")
			}
		case "crop":
			if op.Width <= 0 || op.Height <= 0 {
				return errors.New("crop operation requires positive width and height")
			}
			if op.X < 0 || op.Y < 0 {
				return errors.New("crop coordinates cannot be negative")
			}
		case "convert":
			format := strings.ToLower(op.Format)
			validFormats := []string{"jpeg", "png", "gif", "bmp", "webp"}
			valid := false
			for _, f := range validFormats {
				if format == f {
					valid = true
					break
				}
			}
			if !valid {
				return errors.New("invalid output format. Supported: JPEG, PNG, GIF, BMP, WebP")
			}
		case "watermark":
			if op.Watermark == "" {
				return errors.New("watermark text is required")
			}
			if op.Opacity <= 0 || op.Opacity > 1 {
				return errors.New("opacity must be between 0 and 1")
			}
		default:
			return errors.New("unsupported operation type: " + op.Type + ". Supported: resize, crop, convert, watermark")
		}
	}

	return nil
}

func (s *ImageService) ProcessImage(inputPath string, ops []models.ImageOperation, outputDir string) (string, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	format := utils.GetImageFormat(inputPath)
	var outputPath string

	if format == "gif" {
		outputPath, err = s.processGif(file, ops, inputPath, outputDir)
	} else {
		outputPath, err = s.processStaticImage(file, ops, inputPath, outputDir)
	}

	return outputPath, err
}

func (s *ImageService) processStaticImage(file *os.File, ops []models.ImageOperation, inputPath string, outputDir string) (string, error) {
	img, _, err := utils.DecodeImageFromFile(file)
	if err != nil {
		return "", err
	}

	runtime.GC()

	outputFormat := utils.GetImageFormat(inputPath)

	for _, op := range ops {
		switch op.Type {
		case "resize":
			img = imaging.Resize(img, op.Width, op.Height, imaging.Lanczos)
		case "crop":
			img = imaging.Crop(img, image.Rect(op.X, op.Y, op.X+op.Width, op.Y+op.Height))
		case "convert":
			outputFormat = strings.ToLower(op.Format)
		case "watermark":
			img = s.addTextWatermark(img, op.Watermark, op.Opacity)
		}
		runtime.GC()
	}

	baseName := filepath.Base(inputPath)
	ext := utils.GetExtensionFromFormat(outputFormat)
	outputBase := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ext
	outputPath := filepath.Join(outputDir, outputBase)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	quality := 80
	if len(ops) > 0 && ops[len(ops)-1].Quality > 0 {
		quality = ops[len(ops)-1].Quality
	}

	err = utils.EncodeImage(outFile, img, outputFormat, quality)
	if err != nil {
		return "", err
	}

	img = nil
	runtime.GC()

	return outputPath, nil
}

func (s *ImageService) processGif(file *os.File, ops []models.ImageOperation, inputPath string, outputDir string) (string, error) {
	gifData, err := utils.DecodeGifFromFile(file)
	if err != nil {
		return "", err
	}

	runtime.GC()

	outputFormat := "gif"
	convertTo := ""

	for _, op := range ops {
		if op.Type == "convert" {
			convertTo = strings.ToLower(op.Format)
			outputFormat = convertTo
		}
	}

	if convertTo != "" && convertTo != "gif" {
		return s.convertGifToStatic(gifData, ops, inputPath, outputDir, convertTo)
	}

	for i := range gifData.Image {
		var processed image.Image = gifData.Image[i]

		for _, op := range ops {
			switch op.Type {
			case "resize":
				processed = imaging.Resize(processed, op.Width, op.Height, imaging.Lanczos)
			case "crop":
				processed = imaging.Crop(processed, image.Rect(op.X, op.Y, op.X+op.Width, op.Y+op.Height))
			case "watermark":
				processed = s.addTextWatermark(processed, op.Watermark, op.Opacity)
			}
		}

		paletted := image.NewPaletted(processed.Bounds(), gifData.Image[i].Palette)
		draw.Draw(paletted, processed.Bounds(), processed, processed.Bounds().Min, draw.Src)
		gifData.Image[i] = paletted
	}

	if len(ops) > 0 && ops[0].Type == "resize" {
		gifData.Config.Width = ops[0].Width
		gifData.Config.Height = ops[0].Height
	}

	baseName := filepath.Base(inputPath)
	ext := utils.GetExtensionFromFormat(outputFormat)
	outputBase := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ext
	outputPath := filepath.Join(outputDir, outputBase)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	err = utils.EncodeGif(outFile, gifData)
	if err != nil {
		return "", err
	}

	gifData = nil
	runtime.GC()

	return outputPath, nil
}

func (s *ImageService) convertGifToStatic(gifData *gif.GIF, ops []models.ImageOperation, inputPath string, outputDir string, outputFormat string) (string, error) {
	if len(gifData.Image) == 0 {
		return "", errors.New("gif has no frames")
	}

	img := image.Image(gifData.Image[0])
	runtime.GC()

	for _, op := range ops {
		switch op.Type {
		case "resize":
			img = imaging.Resize(img, op.Width, op.Height, imaging.Lanczos)
		case "crop":
			img = imaging.Crop(img, image.Rect(op.X, op.Y, op.X+op.Width, op.Y+op.Height))
		case "watermark":
			img = s.addTextWatermark(img, op.Watermark, op.Opacity)
		}
		runtime.GC()
	}

	baseName := filepath.Base(inputPath)
	ext := utils.GetExtensionFromFormat(outputFormat)
	outputBase := strings.TrimSuffix(baseName, filepath.Ext(baseName)) + ext
	outputPath := filepath.Join(outputDir, outputBase)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	quality := 80
	if len(ops) > 0 && ops[len(ops)-1].Quality > 0 {
		quality = ops[len(ops)-1].Quality
	}

	err = utils.EncodeImage(outFile, img, outputFormat, quality)
	if err != nil {
		return "", err
	}

	img = nil
	runtime.GC()

	return outputPath, nil
}

func (s *ImageService) addTextWatermark(img image.Image, text string, opacity float64) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	draw.Draw(dst, bounds, img, bounds.Min, draw.Src)

	alpha := uint8(opacity * 255)
	watermarkColor := color.RGBA{255, 255, 255, alpha}
	textColor := color.RGBA{0, 0, 0, alpha}

	simpleDrawRect(dst, image.Rect(bounds.Min.X+10, bounds.Max.Y-40, bounds.Min.X+len(text)*10+20, bounds.Max.Y-10), watermarkColor)

	simpleDrawText(dst, bounds.Min.X+20, bounds.Max.Y-25, text, textColor)

	return dst
}

func simpleDrawRect(img *image.RGBA, rect image.Rectangle, c color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			if x >= 0 && x < img.Bounds().Max.X && y >= 0 && y < img.Bounds().Max.Y {
				img.Set(x, y, c)
			}
		}
	}
}

func simpleDrawText(img *image.RGBA, x, y int, text string, c color.Color) {
	for i, ch := range text {
		chX := x + i*10
		simpleDrawChar(img, chX, y, ch, c)
	}
}

func simpleDrawChar(img *image.RGBA, x, y int, ch rune, c color.Color) {
	for i := 0; i < 8; i++ {
		for j := 0; j < 10; j++ {
			if (i+j)%3 == 0 {
				if x+i >= 0 && x+i < img.Bounds().Max.X && y+j >= 0 && y+j < img.Bounds().Max.Y {
					img.Set(x+i, y+j, c)
				}
			}
		}
	}
}

func (s *ImageService) CheckMemoryUsage() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc < uint64(config.MaxMemoryPerImage)
}
