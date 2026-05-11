package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"mercury/api"
	"mercury/mercator"
)

func main() {
	port := flag.String("port", "8505", "Server port")
	flag.Parse()

	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}

	http.HandleFunc("/api/v1/latlng-to-mercator", handleLatLngToMercator)
	http.HandleFunc("/api/v1/mercator-to-latlng", handleMercatorToLatLng)
	http.HandleFunc("/api/v1/latlng-to-tile", handleLatLngToTile)
	http.HandleFunc("/api/v1/tile-to-bounds", handleTileToBounds)
	http.HandleFunc("/api/v1/latlng-to-pixel", handleLatLngToPixel)
	http.HandleFunc("/api/v1/pixel-to-latlng", handlePixelToLatLng)
	http.HandleFunc("/api/v1/view-bounds", handleViewBounds)
	http.HandleFunc("/api/v1/distance", handleDistance)
	http.HandleFunc("/api/v1/area", handleArea)

	addr := ":" + *port
	fmt.Printf("Server listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, status int, message string) {
	sendJSON(w, status, api.ErrorResponse{Error: message})
}

func handleLatLngToMercator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.LatLngToMercatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.MercatorResult, len(req.Coordinates))
	for i, coord := range req.Coordinates {
		result := mercator.LatLngToMercator(coord.Lat, coord.Lng)
		results[i] = api.MercatorResult{
			Mercator: api.Mercator{X: result.X, Y: result.Y},
			Warning:  result.Warning,
		}
	}

	sendJSON(w, http.StatusOK, api.LatLngToMercatorResponse{Results: results})
}

func handleMercatorToLatLng(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.MercatorToLatLngRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.LatLngResult, len(req.Coordinates))
	for i, coord := range req.Coordinates {
		result := mercator.MercatorToLatLng(coord.X, coord.Y)
		results[i] = api.LatLngResult{
			LatLng:  api.LatLng{Lat: result.Y, Lng: result.X},
			Warning: result.Warning,
		}
	}

	sendJSON(w, http.StatusOK, api.MercatorToLatLngResponse{Results: results})
}

func handleLatLngToTile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.LatLngToTileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.TileResult, len(req.Coordinates))
	for i, coord := range req.Coordinates {
		result := mercator.LatLngToTile(coord.Lat, coord.Lng, req.Zoom)
		results[i] = api.TileResult{
			Tile:    api.Tile{X: result.X, Y: result.Y, Z: result.Z},
			Warning: result.Warning,
		}
	}

	sendJSON(w, http.StatusOK, api.LatLngToTileResponse{Results: results})
}

func handleTileToBounds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.TileToBoundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.TileBounds, len(req.Tiles))
	for i, tile := range req.Tiles {
		result := mercator.TileToLatLngBounds(tile.X, tile.Y, tile.Z)
		results[i] = api.TileBounds{
			Tile:  tile,
			North: result.North,
			South: result.South,
			East:  result.East,
			West:  result.West,
		}
	}

	sendJSON(w, http.StatusOK, api.TileToBoundsResponse{Results: results})
}

func handleLatLngToPixel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.LatLngToPixelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.PixelResult, len(req.Coordinates))
	for i, coord := range req.Coordinates {
		result := mercator.LatLngToPixel(
			req.View.Center.Lat, req.View.Center.Lng,
			req.View.Zoom, req.View.Width, req.View.Height,
			coord.Lat, coord.Lng,
		)
		results[i] = api.PixelResult{
			Pixel:   api.Pixel{X: result.X, Y: result.Y},
			Warning: result.Warning,
		}
	}

	sendJSON(w, http.StatusOK, api.LatLngToPixelResponse{Results: results})
}

func handlePixelToLatLng(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.PixelToLatLngRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	results := make([]api.LatLngResult, len(req.Pixels))
	for i, pixel := range req.Pixels {
		result := mercator.PixelToLatLng(
			req.View.Center.Lat, req.View.Center.Lng,
			req.View.Zoom, req.View.Width, req.View.Height,
			pixel.X, pixel.Y,
		)
		results[i] = api.LatLngResult{
			LatLng:  api.LatLng{Lat: result.Y, Lng: result.X},
			Warning: result.Warning,
		}
	}

	sendJSON(w, http.StatusOK, api.PixelToLatLngResponse{Results: results})
}

func handleViewBounds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.ViewBoundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result := mercator.GetViewBounds(
		req.View.Center.Lat, req.View.Center.Lng,
		req.View.Zoom, req.View.Width, req.View.Height,
	)

	sendJSON(w, http.StatusOK, api.ViewBoundsResponse{
		Bounds: api.ViewBounds{
			NW: api.LatLng{Lat: result.NW.Lat, Lng: result.NW.Lng},
			NE: api.LatLng{Lat: result.NE.Lat, Lng: result.NE.Lng},
			SW: api.LatLng{Lat: result.SW.Lat, Lng: result.SW.Lng},
			SE: api.LatLng{Lat: result.SE.Lat, Lng: result.SE.Lng},
		},
		Warning: result.Warning,
	})
}

func handleDistance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.DistanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	result := mercator.HaversineDistance(req.From.Lat, req.From.Lng, req.To.Lat, req.To.Lng)

	sendJSON(w, http.StatusOK, api.DistanceResponse{
		DistanceMeters:     result.DistanceMeters,
		DistanceKilometers: result.DistanceKilometers,
		Warning:            result.Warning,
	})
}

func handleArea(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.AreaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	polygon := make([]mercator.LatLng, len(req.Polygon))
	for i, coord := range req.Polygon {
		polygon[i] = mercator.LatLng{Lat: coord.Lat, Lng: coord.Lng}
	}

	result := mercator.SphericalPolygonArea(polygon)

	sendJSON(w, http.StatusOK, api.AreaResponse{
		AreaSquareMeters:     result.AreaSquareMeters,
		AreaSquareKilometers: result.AreaSquareKilometers,
		Warning:              result.Warning,
	})
}
