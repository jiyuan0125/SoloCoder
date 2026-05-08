package qrcode

import (
	"bytes"
	"errors"
	"image/png"

	"github.com/skip2/go-qrcode"

	"github.com/qrcode/generator/pkg/common"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(req common.GenerateRequest) (*common.GenerateResponse, error) {
	opts, err := NewOptions(req.Content, req.ErrorLevel, req.Size, req.Logo)
	if err != nil {
		return &common.GenerateResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	qrCode, err := qrcode.New(opts.Content, opts.ErrorLevel)
	if err != nil {
		errMsg := err.Error()
		if errMsg == "content too long" {
			errMsg = "content is too long for the selected error correction level"
		}
		return &common.GenerateResponse{
			Success: false,
			Message: errMsg,
		}, errors.New(errMsg)
	}

	qrImg := qrCode.Image(opts.Size)

	if opts.HasLogo && opts.Logo != nil {
		qrImg, err = overlayLogo(qrImg, opts.Logo)
		if err != nil {
			return &common.GenerateResponse{
				Success: false,
				Message: err.Error(),
			}, err
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, qrImg); err != nil {
		return &common.GenerateResponse{
			Success: false,
			Message: "failed to encode PNG: " + err.Error(),
		}, err
	}

	return &common.GenerateResponse{
		Success: true,
		Message: "QR code generated successfully",
		Image:   buf.Bytes(),
	}, nil
}
