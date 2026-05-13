package qrcode

import (
	"bytes"
	"image/jpeg"
	"image/png"
	"testing"

	"qrcode-service/internal/types"
)

func TestGeneratePNGAndDecode(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "https://example.com/test?foo=bar&name=张三",
		Format:     "PNG",
		Size:       256,
		Foreground: "#000000",
		Background: "#FFFFFF",
		ErrorLevel: "M",
		Border:     0,
		Title:      "",
	}

	data, err := GenerateQR(req)
	if err != nil {
		t.Fatalf("Failed to generate QR: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Generated PNG is empty")
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode PNG round-trip: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() < 100 || bounds.Dy() < 100 {
		t.Errorf("Image too small: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateJPEGAndDecode(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "MECARD:N:张三;TEL:13800138000;EMAIL:test@example.com;;",
		Format:     "JPEG",
		Size:       300,
		Foreground: "#333333",
		Background: "#FFFF00",
		ErrorLevel: "Q",
	}

	data, err := GenerateQR(req)
	if err != nil {
		t.Fatalf("Failed to generate QR: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Generated JPEG is empty")
	}

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to decode JPEG round-trip: %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() < 100 || bounds.Dy() < 100 {
		t.Errorf("Image too small: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func TestGenerateSVG(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "WIFI:S:MyNetwork;T:WPA;P:secret123;;",
		Format:     "SVG",
		Size:       256,
		Foreground: "#FF0000",
		Background: "#FFFFFF",
	}

	data, err := GenerateQR(req)
	if err != nil {
		t.Fatalf("Failed to generate SVG QR: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Generated SVG is empty")
	}

	if !bytes.Contains(data, []byte("<svg")) {
		t.Error("SVG does not contain svg tag")
	}

	if !bytes.Contains(data, []byte("#FF0000")) {
		t.Error("SVG does not contain foreground color")
	}
}

func TestInvalidFormat(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "test",
		Format:     "BMP",
		Foreground: "#000000",
		Background: "#FFFFFF",
	}

	_, err := GenerateQR(req)
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

func TestInvalidColor(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "test",
		Format:     "PNG",
		Foreground: "invalid",
		Background: "#FFFFFF",
	}

	_, err := GenerateQR(req)
	if err == nil {
		t.Error("Expected error for invalid color")
	}
}

func TestEmptyContent(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "",
		Format:     "PNG",
		Foreground: "#000000",
		Background: "#FFFFFF",
	}

	_, err := GenerateQR(req)
	if err == nil {
		t.Error("Expected error for empty content")
	}
}

func TestShortColorHex(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "https://example.com",
		Format:     "PNG",
		Size:       200,
		Foreground: "#F00",
		Background: "#FFF",
	}

	data, err := GenerateQR(req)
	if err != nil {
		t.Fatalf("Failed to generate with short hex: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Generated PNG is empty")
	}
}

func TestBorderAndTitle(t *testing.T) {
	req := &types.QRCodeRequest{
		Content:    "https://example.com",
		Format:     "PNG",
		Size:       200,
		Foreground: "#000000",
		Background: "#FFFFFF",
		Border:     20,
		Title:      "Test Title",
	}

	data, err := GenerateQR(req)
	if err != nil {
		t.Fatalf("Failed to generate with border: %v", err)
	}

	if len(data) == 0 {
		t.Fatal("Generated PNG is empty")
	}
}
