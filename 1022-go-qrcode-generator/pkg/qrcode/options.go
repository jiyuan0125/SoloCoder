package qrcode

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"bytes"

	"github.com/skip2/go-qrcode"

	"github.com/qrcode/generator/pkg/common"
)

type Options struct {
	Content    string
	ErrorLevel qrcode.RecoveryLevel
	Size       int
	Logo       image.Image
	HasLogo    bool
}

func NewOptions(content string, errorLevel common.ErrorLevel, size int, logoData []byte) (*Options, error) {
	if content == "" {
		return nil, errors.New("content cannot be empty")
	}

	opts := &Options{
		Content: content,
	}

	switch errorLevel {
	case common.ErrorLevelL:
		opts.ErrorLevel = qrcode.Low
	case common.ErrorLevelM:
		opts.ErrorLevel = qrcode.Medium
	case common.ErrorLevelQ:
		opts.ErrorLevel = qrcode.High
	case common.ErrorLevelH:
		opts.ErrorLevel = qrcode.Highest
	default:
		opts.ErrorLevel = qrcode.Medium
	}

	if size <= 0 {
		size = common.DefaultSize
	}
	opts.Size = size

	if len(logoData) > 0 {
		img, _, err := image.Decode(bytes.NewReader(logoData))
		if err != nil {
			return nil, errors.New("invalid logo image format: " + err.Error())
		}
		opts.Logo = img
		opts.HasLogo = true

		if opts.ErrorLevel != qrcode.Highest {
			opts.ErrorLevel = qrcode.Highest
		}
	}

	return opts, nil
}
