package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"geodist/common"
)

func printUsage() {
	fmt.Print(`Usage: geoclient <command> [options]

Commands:
  health                  Check server health
  distance                Calculate distance between two points
  polyline-length         Calculate total length of a polyline
  closest                 Find closest point on polyline to a given point
  batch                   Calculate distances from center to multiple targets
  circle                  Generate great circle boundary points

Global Options:
  -server URL             Server URL (default: http://localhost:8080)
  -ellipsoid              Use Vincenty ellipsoid model (default: Haversine sphere)

Command-specific Options:
  distance:
    -from lat,lng         Starting point
    -to lat,lng           Ending point

  polyline-length:
    -points "lat1,lng1;lat2,lng2;..."   List of waypoints

  closest:
    -point lat,lng        Reference point
    -polyline "lat1,lng1;lat2,lng2;..." Polyline points

  batch:
    -center lat,lng       Center point
    -targets "lat1,lng1;lat2,lng2;..."  Target points

  circle:
    -center lat,lng       Center point
    -radius km            Radius in kilometers
    -points N             Number of boundary points (default: 36, minimum: 36)
`)
}

func parseCoord(s string) (common.Coordinate, error) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) != 2 {
		return common.Coordinate{}, fmt.Errorf("invalid coordinate format: %s (expected lat,lng)", s)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return common.Coordinate{}, fmt.Errorf("invalid latitude: %s", parts[0])
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return common.Coordinate{}, fmt.Errorf("invalid longitude: %s", parts[1])
	}
	return common.Coordinate{Lat: lat, Lng: lng}, nil
}

func parseCoordList(s string) ([]common.Coordinate, error) {
	segments := strings.Split(s, ";")
	coords := make([]common.Coordinate, 0, len(segments))
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		c, err := parseCoord(seg)
		if err != nil {
			return nil, err
		}
		coords = append(coords, c)
	}
	return coords, nil
}

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}
	fmt.Println(string(b))
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	fs := flag.NewFlagSet("geoclient", flag.ExitOnError)
	serverURL := fs.String("server", "http://localhost:8080", "Server URL")
	useEllipsoid := fs.Bool("ellipsoid", false, "Use Vincenty ellipsoid model")

	var from, to, pointStr, pointsStr, polylineStr, centerStr, targetsStr string
	var radiusKm float64
	var numPoints int

	switch cmd {
	case "distance":
		fs.StringVar(&from, "from", "", "Starting point (lat,lng)")
		fs.StringVar(&to, "to", "", "Ending point (lat,lng)")
	case "polyline-length":
		fs.StringVar(&pointsStr, "points", "", "Waypoints (lat1,lng1;lat2,lng2;...)")
	case "closest":
		fs.StringVar(&pointStr, "point", "", "Reference point (lat,lng)")
		fs.StringVar(&polylineStr, "polyline", "", "Polyline points (lat1,lng1;lat2,lng2;...)")
	case "batch":
		fs.StringVar(&centerStr, "center", "", "Center point (lat,lng)")
		fs.StringVar(&targetsStr, "targets", "", "Target points (lat1,lng1;lat2,lng2;...)")
	case "circle":
		fs.StringVar(&centerStr, "center", "", "Center point (lat,lng)")
		fs.Float64Var(&radiusKm, "radius", 0, "Radius in kilometers")
		fs.IntVar(&numPoints, "points", 36, "Number of boundary points")
	case "health", "-h", "--help":
	default:
		fmt.Printf("Unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	fs.Parse(os.Args[2:])

	client := NewAPIClient(*serverURL)

	switch cmd {
	case "health":
		err := client.Health()
		if err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Server is healthy")

	case "distance":
		if from == "" || to == "" {
			fmt.Println("Error: -from and -to are required")
			os.Exit(1)
		}
		pointA, err := parseCoord(from)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		pointB, err := parseCoord(to)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		dist, err := client.Distance(pointA, pointB, *useEllipsoid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Distance: %.4f km\n", dist)

	case "polyline-length":
		if pointsStr == "" {
			fmt.Println("Error: -points is required")
			os.Exit(1)
		}
		points, err := parseCoordList(pointsStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(points) < 2 {
			fmt.Println("Error: polyline requires at least 2 points")
			os.Exit(1)
		}
		result, err := client.PolylineLength(points, *useEllipsoid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		prettyPrint(result)

	case "closest":
		if pointStr == "" || polylineStr == "" {
			fmt.Println("Error: -point and -polyline are required")
			os.Exit(1)
		}
		point, err := parseCoord(pointStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		polyline, err := parseCoordList(polylineStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(polyline) < 1 {
			fmt.Println("Error: polyline requires at least 1 point")
			os.Exit(1)
		}
		result, err := client.PointToPolyline(point, polyline, *useEllipsoid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		prettyPrint(result)

	case "batch":
		if centerStr == "" || targetsStr == "" {
			fmt.Println("Error: -center and -targets are required")
			os.Exit(1)
		}
		center, err := parseCoord(centerStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		targets, err := parseCoordList(targetsStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		result, err := client.BatchDistances(center, targets, *useEllipsoid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		prettyPrint(result)

	case "circle":
		if centerStr == "" || radiusKm <= 0 {
			fmt.Println("Error: -center and -radius (positive) are required")
			os.Exit(1)
		}
		center, err := parseCoord(centerStr)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		result, err := client.Circle(center, radiusKm, numPoints, *useEllipsoid)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		prettyPrint(result)

	case "-h", "--help":
		printUsage()
	}
}
