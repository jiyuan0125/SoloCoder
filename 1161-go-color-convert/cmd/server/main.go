package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"colorconv/pkg/colorconv"
	"colorconv/pkg/common"
)

func main() {
	port := flag.String("port", "", "Server port")
	flag.Parse()

	if *port == "" {
		*port = os.Getenv("COLORCONV_PORT")
	}
	if *port == "" {
		*port = "8080"
	}

	http.HandleFunc("/convert", handleConvert)
	http.HandleFunc("/brightness", handleBrightness)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Server listening on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func handleConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.ConvertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result, err := convertColor(req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeSuccess(w, result)
}

func handleBrightness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.BrightnessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	luminance, err := colorconv.RelativeLuminanceFromHex(req.Hex)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.BrightnessResponse{
		Success:   true,
		Luminance: luminance,
	})
}

func convertColor(req common.ConvertRequest) (interface{}, error) {
	var rgb *colorconv.RGB
	var err error

	switch strings.ToLower(req.From) {
	case "rgb":
		if len(req.Values) != 3 {
			return nil, fmt.Errorf("RGB requires 3 values")
		}
		r, er1 := strconv.Atoi(req.Values[0])
		g, er2 := strconv.Atoi(req.Values[1])
		b, er3 := strconv.Atoi(req.Values[2])
		if er1 != nil || er2 != nil || er3 != nil {
			return nil, fmt.Errorf("invalid RGB values")
		}
		rgb, err = colorconv.NewRGB(r, g, b)
	case "hex":
		if len(req.Values) != 1 {
			return nil, fmt.Errorf("hex requires 1 value")
		}
		rgb, err = colorconv.ParseHex(req.Values[0])
	case "hsl":
		if len(req.Values) != 3 {
			return nil, fmt.Errorf("HSL requires 3 values")
		}
		h, er1 := strconv.ParseFloat(req.Values[0], 64)
		s, er2 := strconv.ParseFloat(req.Values[1], 64)
		l, er3 := strconv.ParseFloat(req.Values[2], 64)
		if er1 != nil || er2 != nil || er3 != nil {
			return nil, fmt.Errorf("invalid HSL values")
		}
		hsl, err2 := colorconv.NewHSL(h, s, l)
		if err2 != nil {
			return nil, err2
		}
		rgb = hsl.ToRGB()
	case "hsv":
		if len(req.Values) != 3 {
			return nil, fmt.Errorf("HSV requires 3 values")
		}
		h, er1 := strconv.ParseFloat(req.Values[0], 64)
		s, er2 := strconv.ParseFloat(req.Values[1], 64)
		v, er3 := strconv.ParseFloat(req.Values[2], 64)
		if er1 != nil || er2 != nil || er3 != nil {
			return nil, fmt.Errorf("invalid HSV values")
		}
		hsv, err2 := colorconv.NewHSV(h, s, v)
		if err2 != nil {
			return nil, err2
		}
		rgb = hsv.ToRGB()
	case "cmyk":
		if len(req.Values) != 4 {
			return nil, fmt.Errorf("CMYK requires 4 values")
		}
		c, er1 := strconv.ParseFloat(req.Values[0], 64)
		m, er2 := strconv.ParseFloat(req.Values[1], 64)
		y, er3 := strconv.ParseFloat(req.Values[2], 64)
		k, er4 := strconv.ParseFloat(req.Values[3], 64)
		if er1 != nil || er2 != nil || er3 != nil || er4 != nil {
			return nil, fmt.Errorf("invalid CMYK values")
		}
		cmyk, err2 := colorconv.NewCMYK(c, m, y, k)
		if err2 != nil {
			return nil, err2
		}
		rgb = cmyk.ToRGB()
	default:
		return nil, fmt.Errorf("unsupported 'from' format: %s", req.From)
	}

	if err != nil {
		return nil, err
	}

	switch strings.ToLower(req.To) {
	case "rgb":
		return common.RGBResult{R: rgb.R, G: rgb.G, B: rgb.B}, nil
	case "hex":
		return common.HexResult{Hex: rgb.ToHex()}, nil
	case "hsl":
		hsl := rgb.ToHSL()
		return common.HSLResult{H: hsl.H, S: hsl.S, L: hsl.L}, nil
	case "hsv":
		hsv := rgb.ToHSV()
		return common.HSVResult{H: hsv.H, S: hsv.S, V: hsv.V}, nil
	case "cmyk":
		cmyk := rgb.ToCMYK()
		return common.CMYKResult{C: cmyk.C, M: cmyk.M, Y: cmyk.Y, K: cmyk.K}, nil
	default:
		return nil, fmt.Errorf("unsupported 'to' format: %s", req.To)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(common.ConvertResponse{
		Success: false,
		Error:   message,
	})
}

func writeSuccess(w http.ResponseWriter, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(common.ConvertResponse{
		Success: true,
		Result:  result,
	})
}
