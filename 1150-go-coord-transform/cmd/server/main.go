package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"coordconv/pkg/coordconv"
	"coordconv/pkg/model"
)

func main() {
	http.HandleFunc("/api/convert", convertHandler)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req model.ConvertRequest
	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
	} else if r.Method == http.MethodGet {
		req.From = r.URL.Query().Get("from")
		req.To = r.URL.Query().Get("to")
		lngStr := r.URL.Query().Get("lng")
		latStr := r.URL.Query().Get("lat")
		pointsStr := r.URL.Query().Get("points")

		if lngStr != "" && latStr != "" {
			lng, err := strconv.ParseFloat(lngStr, 64)
			if err != nil {
				sendError(w, http.StatusBadRequest, "Invalid longitude")
				return
			}
			lat, err := strconv.ParseFloat(latStr, 64)
			if err != nil {
				sendError(w, http.StatusBadRequest, "Invalid latitude")
				return
			}
			req.Lng = &lng
			req.Lat = &lat
		}

		if pointsStr != "" {
			pointsParts := strings.Split(pointsStr, "|")
			for _, pp := range pointsParts {
				parts := strings.Split(pp, ",")
				if len(parts) != 2 {
					sendError(w, http.StatusBadRequest, "Invalid points format")
					return
				}
				lng, err := strconv.ParseFloat(parts[0], 64)
				if err != nil {
					sendError(w, http.StatusBadRequest, "Invalid longitude in points")
					return
				}
				lat, err := strconv.ParseFloat(parts[1], 64)
				if err != nil {
					sendError(w, http.StatusBadRequest, "Invalid latitude in points")
					return
				}
				req.Points = append(req.Points, model.PointRequest{Lng: lng, Lat: lat})
			}
		}
	} else {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if req.From == "" || req.To == "" {
		sendError(w, http.StatusBadRequest, "Missing from or to parameter")
		return
	}

	from := coordconv.CoordSystem(strings.ToLower(req.From))
	to := coordconv.CoordSystem(strings.ToLower(req.To))

	var points []coordconv.Point
	if req.Lng != nil && req.Lat != nil {
		points = append(points, coordconv.Point{Lng: *req.Lng, Lat: *req.Lat})
	}
	for _, p := range req.Points {
		points = append(points, coordconv.Point{Lng: p.Lng, Lat: p.Lat})
	}

	if len(points) == 0 {
		sendError(w, http.StatusBadRequest, "No coordinates provided")
		return
	}

	results := make([]model.PointResult, 0, len(points))
	for _, p := range points {
		converted, err := coordconv.Convert(from, to, p)
		if err != nil {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		results = append(results, model.PointResult{Lng: converted.Lng, Lat: converted.Lat})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.ConvertResponse{
		Success: true,
		Points:  results,
	})
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(model.ConvertResponse{
		Success: false,
		Error:   msg,
	})
}
