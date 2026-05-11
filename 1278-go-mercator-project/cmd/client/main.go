package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"mercury/api"
)

const baseURL = "http://localhost:8505/api/v1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "latlng-to-mercator":
		err = latLngToMercator(args)
	case "mercator-to-latlng":
		err = mercatorToLatLng(args)
	case "latlng-to-tile":
		err = latLngToTile(args)
	case "tile-to-bounds":
		err = tileToBounds(args)
	case "latlng-to-pixel":
		err = latLngToPixel(args)
	case "pixel-to-latlng":
		err = pixelToLatLng(args)
	case "view-bounds":
		err = viewBounds(args)
	case "distance":
		err = distance(args)
	case "area":
		err = area(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Mercator Coordinate Conversion Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mercator-client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  latlng-to-mercator <lat>,<lng> [lat,lng...]")
	fmt.Println("  mercator-to-latlng <x>,<y> [x,y...]")
	fmt.Println("  latlng-to-tile <zoom> <lat>,<lng> [lat,lng...]")
	fmt.Println("  tile-to-bounds <x>,<y>,<z> [x,y,z...]")
	fmt.Println("  latlng-to-pixel <center-lat>,<center-lng>,<zoom>,<width>,<height> <lat>,<lng> [lat,lng...]")
	fmt.Println("  pixel-to-latlng <center-lat>,<center-lng>,<zoom>,<width>,<height> <x>,<y> [x,y...]")
	fmt.Println("  view-bounds <center-lat>,<center-lng>,<zoom>,<width>,<height>")
	fmt.Println("  distance <from-lat>,<from-lng> <to-lat>,<to-lng>")
	fmt.Println("  area <lat1>,<lng1> <lat2>,<lng2> <lat3>,<lng3> [lat,lng...]")
	fmt.Println("  help - Show this help")
}

func makeRequest(endpoint string, requestBody interface{}, responseBody interface{}) error {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := baseURL + "/" + endpoint
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.Unmarshal(body, &errResp); err == nil {
			return fmt.Errorf("server error: %s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, responseBody); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	return nil
}

func parseLatLng(s string) (api.LatLng, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return api.LatLng{}, fmt.Errorf("invalid format, expected lat,lng")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return api.LatLng{}, fmt.Errorf("invalid latitude: %w", err)
	}

	lng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return api.LatLng{}, fmt.Errorf("invalid longitude: %w", err)
	}

	return api.LatLng{Lat: lat, Lng: lng}, nil
}

func parseMercator(s string) (api.Mercator, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return api.Mercator{}, fmt.Errorf("invalid format, expected x,y")
	}

	x, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return api.Mercator{}, fmt.Errorf("invalid x: %w", err)
	}

	y, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return api.Mercator{}, fmt.Errorf("invalid y: %w", err)
	}

	return api.Mercator{X: x, Y: y}, nil
}

func parseTile(s string) (api.Tile, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 3 {
		return api.Tile{}, fmt.Errorf("invalid format, expected x,y,z")
	}

	x, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return api.Tile{}, fmt.Errorf("invalid x: %w", err)
	}

	y, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return api.Tile{}, fmt.Errorf("invalid y: %w", err)
	}

	z, err := strconv.Atoi(parts[2])
	if err != nil {
		return api.Tile{}, fmt.Errorf("invalid z: %w", err)
	}

	return api.Tile{X: x, Y: y, Z: z}, nil
}

func parsePixel(s string) (api.Pixel, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return api.Pixel{}, fmt.Errorf("invalid format, expected x,y")
	}

	x, err := strconv.Atoi(parts[0])
	if err != nil {
		return api.Pixel{}, fmt.Errorf("invalid x: %w", err)
	}

	y, err := strconv.Atoi(parts[1])
	if err != nil {
		return api.Pixel{}, fmt.Errorf("invalid y: %w", err)
	}

	return api.Pixel{X: x, Y: y}, nil
}

func parseView(s string) (api.MapView, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 5 {
		return api.MapView{}, fmt.Errorf("invalid format, expected center-lat,center-lng,zoom,width,height")
	}

	centerLat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return api.MapView{}, fmt.Errorf("invalid center latitude: %w", err)
	}

	centerLng, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return api.MapView{}, fmt.Errorf("invalid center longitude: %w", err)
	}

	zoom, err := strconv.Atoi(parts[2])
	if err != nil {
		return api.MapView{}, fmt.Errorf("invalid zoom: %w", err)
	}

	width, err := strconv.Atoi(parts[3])
	if err != nil {
		return api.MapView{}, fmt.Errorf("invalid width: %w", err)
	}

	height, err := strconv.Atoi(parts[4])
	if err != nil {
		return api.MapView{}, fmt.Errorf("invalid height: %w", err)
	}

	return api.MapView{
		Center: api.LatLng{Lat: centerLat, Lng: centerLng},
		Zoom:   zoom,
		Width:  width,
		Height: height,
	}, nil
}

func latLngToMercator(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: latlng-to-mercator <lat>,<lng> [lat,lng...]")
	}

	coords := make([]api.LatLng, len(args))
	for i, arg := range args {
		coord, err := parseLatLng(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
		coords[i] = coord
	}

	var resp api.LatLngToMercatorResponse
	if err := makeRequest("latlng-to-mercator", api.LatLngToMercatorRequest{Coordinates: coords}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Coordinate %d:\n", i+1)
		fmt.Printf("  Mercator: X=%.6f, Y=%.6f\n", result.Mercator.X, result.Mercator.Y)
		if result.Warning != "" {
			fmt.Printf("  Warning: %s\n", result.Warning)
		}
	}

	return nil
}

func mercatorToLatLng(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mercator-to-latlng <x>,<y> [x,y...]")
	}

	coords := make([]api.Mercator, len(args))
	for i, arg := range args {
		coord, err := parseMercator(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
		coords[i] = coord
	}

	var resp api.MercatorToLatLngResponse
	if err := makeRequest("mercator-to-latlng", api.MercatorToLatLngRequest{Coordinates: coords}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Coordinate %d:\n", i+1)
		fmt.Printf("  LatLng: Lat=%.6f, Lng=%.6f\n", result.LatLng.Lat, result.LatLng.Lng)
		if result.Warning != "" {
			fmt.Printf("  Warning: %s\n", result.Warning)
		}
	}

	return nil
}

func latLngToTile(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: latlng-to-tile <zoom> <lat>,<lng> [lat,lng...]")
	}

	zoom, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("invalid zoom: %w", err)
	}

	coords := make([]api.LatLng, len(args)-1)
	for i, arg := range args[1:] {
		coord, err := parseLatLng(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+2, err)
		}
		coords[i] = coord
	}

	var resp api.LatLngToTileResponse
	if err := makeRequest("latlng-to-tile", api.LatLngToTileRequest{Coordinates: coords, Zoom: zoom}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Coordinate %d:\n", i+1)
		fmt.Printf("  Tile: X=%d, Y=%d, Z=%d\n", result.Tile.X, result.Tile.Y, result.Tile.Z)
		if result.Warning != "" {
			fmt.Printf("  Warning: %s\n", result.Warning)
		}
	}

	return nil
}

func tileToBounds(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: tile-to-bounds <x>,<y>,<z> [x,y,z...]")
	}

	tiles := make([]api.Tile, len(args))
	for i, arg := range args {
		tile, err := parseTile(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
		tiles[i] = tile
	}

	var resp api.TileToBoundsResponse
	if err := makeRequest("tile-to-bounds", api.TileToBoundsRequest{Tiles: tiles}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Tile %d (X=%d, Y=%d, Z=%d):\n", i+1, result.Tile.X, result.Tile.Y, result.Tile.Z)
		fmt.Printf("  Bounds: North=%.6f, South=%.6f, East=%.6f, West=%.6f\n",
			result.North, result.South, result.East, result.West)
	}

	return nil
}

func latLngToPixel(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: latlng-to-pixel <center-lat>,<center-lng>,<zoom>,<width>,<height> <lat>,<lng> [lat,lng...]")
	}

	view, err := parseView(args[0])
	if err != nil {
		return fmt.Errorf("invalid view: %w", err)
	}

	coords := make([]api.LatLng, len(args)-1)
	for i, arg := range args[1:] {
		coord, err := parseLatLng(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+2, err)
		}
		coords[i] = coord
	}

	var resp api.LatLngToPixelResponse
	if err := makeRequest("latlng-to-pixel", api.LatLngToPixelRequest{View: view, Coordinates: coords}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Coordinate %d:\n", i+1)
		fmt.Printf("  Pixel: X=%d, Y=%d\n", result.Pixel.X, result.Pixel.Y)
		if result.Warning != "" {
			fmt.Printf("  Warning: %s\n", result.Warning)
		}
	}

	return nil
}

func pixelToLatLng(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: pixel-to-latlng <center-lat>,<center-lng>,<zoom>,<width>,<height> <x>,<y> [x,y...]")
	}

	view, err := parseView(args[0])
	if err != nil {
		return fmt.Errorf("invalid view: %w", err)
	}

	pixels := make([]api.Pixel, len(args)-1)
	for i, arg := range args[1:] {
		pixel, err := parsePixel(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+2, err)
		}
		pixels[i] = pixel
	}

	var resp api.PixelToLatLngResponse
	if err := makeRequest("pixel-to-latlng", api.PixelToLatLngRequest{View: view, Pixels: pixels}, &resp); err != nil {
		return err
	}

	for i, result := range resp.Results {
		fmt.Printf("Pixel %d:\n", i+1)
		fmt.Printf("  LatLng: Lat=%.6f, Lng=%.6f\n", result.LatLng.Lat, result.LatLng.Lng)
		if result.Warning != "" {
			fmt.Printf("  Warning: %s\n", result.Warning)
		}
	}

	return nil
}

func viewBounds(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: view-bounds <center-lat>,<center-lng>,<zoom>,<width>,<height>")
	}

	view, err := parseView(args[0])
	if err != nil {
		return fmt.Errorf("invalid view: %w", err)
	}

	var resp api.ViewBoundsResponse
	if err := makeRequest("view-bounds", api.ViewBoundsRequest{View: view}, &resp); err != nil {
		return err
	}

	fmt.Println("View Bounds:")
	fmt.Printf("  NW: Lat=%.6f, Lng=%.6f\n", resp.Bounds.NW.Lat, resp.Bounds.NW.Lng)
	fmt.Printf("  NE: Lat=%.6f, Lng=%.6f\n", resp.Bounds.NE.Lat, resp.Bounds.NE.Lng)
	fmt.Printf("  SW: Lat=%.6f, Lng=%.6f\n", resp.Bounds.SW.Lat, resp.Bounds.SW.Lng)
	fmt.Printf("  SE: Lat=%.6f, Lng=%.6f\n", resp.Bounds.SE.Lat, resp.Bounds.SE.Lng)
	if resp.Warning != "" {
		fmt.Printf("  Warning: %s\n", resp.Warning)
	}

	return nil
}

func distance(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: distance <from-lat>,<from-lng> <to-lat>,<to-lng>")
	}

	from, err := parseLatLng(args[0])
	if err != nil {
		return fmt.Errorf("invalid from coordinate: %w", err)
	}

	to, err := parseLatLng(args[1])
	if err != nil {
		return fmt.Errorf("invalid to coordinate: %w", err)
	}

	var resp api.DistanceResponse
	if err := makeRequest("distance", api.DistanceRequest{From: from, To: to}, &resp); err != nil {
		return err
	}

	fmt.Printf("Distance:\n")
	fmt.Printf("  Meters: %.2f\n", resp.DistanceMeters)
	fmt.Printf("  Kilometers: %.2f\n", resp.DistanceKilometers)
	if resp.Warning != "" {
		fmt.Printf("  Warning: %s\n", resp.Warning)
	}

	return nil
}

func area(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: area <lat1>,<lng1> <lat2>,<lng2> <lat3>,<lng3> [lat,lng...]")
	}

	polygon := make([]api.LatLng, len(args))
	for i, arg := range args {
		coord, err := parseLatLng(arg)
		if err != nil {
			return fmt.Errorf("argument %d: %w", i+1, err)
		}
		polygon[i] = coord
	}

	var resp api.AreaResponse
	if err := makeRequest("area", api.AreaRequest{Polygon: polygon}, &resp); err != nil {
		return err
	}

	fmt.Printf("Area:\n")
	fmt.Printf("  Square Meters: %.2f\n", resp.AreaSquareMeters)
	fmt.Printf("  Square Kilometers: %.2f\n", resp.AreaSquareKilometers)
	if resp.Warning != "" {
		fmt.Printf("  Warning: %s\n", resp.Warning)
	}

	return nil
}
