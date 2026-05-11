package main

import (
	"encoding/json"
	"net/http"

	"geodist/common"
	"geodist/geolib"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, common.ErrorResponse{Error: msg})
}

func toGeoCoord(c common.Coordinate) geolib.Coordinate {
	return geolib.Coordinate{Lat: c.Lat, Lng: c.Lng}
}

func toCommonCoord(c geolib.Coordinate) common.Coordinate {
	return common.Coordinate{Lat: c.Lat, Lng: c.Lng}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleDistance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.DistanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	dist, err := geolib.Distance(
		req.PointA.Lat, req.PointA.Lng,
		req.PointB.Lat, req.PointB.Lng,
		req.UseEllipsoid,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, common.DistanceResponse{DistanceKm: dist})
}

func handlePolylineLength(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.PolylineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Points) < 2 {
		writeError(w, http.StatusBadRequest, "polyline requires at least 2 points")
		return
	}

	geoPoints := make([]geolib.Coordinate, len(req.Points))
	for i, p := range req.Points {
		geoPoints[i] = toGeoCoord(p)
	}

	total, originalCount, simplifiedCount, err := geolib.PolylineLength(geoPoints, req.UseEllipsoid)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	simplified := originalCount != simplifiedCount

	writeJSON(w, http.StatusOK, common.PolylineLengthResponse{
		TotalLengthKm:   total,
		Simplified:      simplified,
		OriginalCount:   originalCount,
		SimplifiedCount: simplifiedCount,
	})
}

func handlePointToPolyline(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.PointToPolylineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Polyline) < 1 {
		writeError(w, http.StatusBadRequest, "polyline requires at least 1 point")
		return
	}

	geoPoly := make([]geolib.Coordinate, len(req.Polyline))
	for i, p := range req.Polyline {
		geoPoly[i] = toGeoCoord(p)
	}

	closestPoint, dist, segIdx, err := geolib.PointToPolylineDistance(
		toGeoCoord(req.Point),
		geoPoly,
		req.UseEllipsoid,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, common.PointToPolylineResponse{
		ClosestDistanceKm: dist,
		ClosestPoint:      toCommonCoord(closestPoint),
		SegmentIndex:      segIdx,
	})
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.BatchDistanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	geoTargets := make([]geolib.Coordinate, len(req.Targets))
	for i, t := range req.Targets {
		geoTargets[i] = toGeoCoord(t)
	}

	results, err := geolib.BatchDistances(
		toGeoCoord(req.Center),
		geoTargets,
		req.UseEllipsoid,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	response := common.BatchDistanceResponse{
		Results: make([]common.TargetDistance, len(results)),
	}

	for i, r := range results {
		response.Results[i] = common.TargetDistance{
			Target:     toCommonCoord(r.Point),
			DistanceKm: r.DistanceKm,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func handleCircle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req common.CircleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RadiusKm <= 0 {
		writeError(w, http.StatusBadRequest, "radius must be positive")
		return
	}

	if err := geolib.ValidateCoordinate(req.Center.Lat, req.Center.Lng); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	numPoints := req.NumPoints
	if numPoints < 36 {
		numPoints = 36
	}

	var boundary []geolib.Coordinate
	if req.UseEllipsoid {
		boundary = geolib.GreatCirclePointsEllipsoid(toGeoCoord(req.Center), req.RadiusKm, numPoints)
	} else {
		boundary = geolib.GreatCirclePoints(toGeoCoord(req.Center), req.RadiusKm, numPoints)
	}

	commonBoundary := make([]common.Coordinate, len(boundary))
	for i, p := range boundary {
		commonBoundary[i] = toCommonCoord(p)
	}

	writeJSON(w, http.StatusOK, common.CircleResponse{Boundary: commonBoundary})
}
