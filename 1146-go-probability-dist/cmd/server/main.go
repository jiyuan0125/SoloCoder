package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"prob-dist/pkg/api"
	"prob-dist/pkg/probdist"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, &api.ErrorResponse{Error: msg})
}

func parseFloatParam(r *http.Request, name string) (float64, bool) {
	val := r.URL.Query().Get(name)
	if val == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(val, 64)
	return f, err == nil
}

func parseIntParam(r *http.Request, name string) (int, bool) {
	val := r.URL.Query().Get(name)
	if val == "" {
		return 0, false
	}
	i, err := strconv.Atoi(val)
	return i, err == nil
}

func handleNormalPDF(w http.ResponseWriter, r *http.Request) {
	mu, ok := parseFloatParam(r, "mu")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid mu parameter")
		return
	}
	sigma, ok := parseFloatParam(r, "sigma")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid sigma parameter")
		return
	}
	x, ok := parseFloatParam(r, "x")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid x parameter")
		return
	}

	dist, err := probdist.NewNormal(mu, sigma)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pdf := dist.PDF(x)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: pdf})
}

func handleNormalCDF(w http.ResponseWriter, r *http.Request) {
	mu, ok := parseFloatParam(r, "mu")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid mu parameter")
		return
	}
	sigma, ok := parseFloatParam(r, "sigma")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid sigma parameter")
		return
	}
	x, ok := parseFloatParam(r, "x")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid x parameter")
		return
	}

	dist, err := probdist.NewNormal(mu, sigma)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cdf := dist.CDF(x)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: cdf})
}

func handleNormalSample(w http.ResponseWriter, r *http.Request) {
	mu, ok := parseFloatParam(r, "mu")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid mu parameter")
		return
	}
	sigma, ok := parseFloatParam(r, "sigma")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid sigma parameter")
		return
	}
	n, ok := parseIntParam(r, "n")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid n parameter")
		return
	}

	dist, err := probdist.NewNormal(mu, sigma)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	samples, err := dist.Sample(n)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &api.FloatSampleResponse{Samples: samples})
}

func handlePoissonPDF(w http.ResponseWriter, r *http.Request) {
	lambda, ok := parseFloatParam(r, "lambda")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid lambda parameter")
		return
	}
	k, ok := parseIntParam(r, "k")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid k parameter (must be integer)")
		return
	}

	dist, err := probdist.NewPoisson(lambda)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pmf := dist.PMF(k)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: pmf})
}

func handlePoissonCDF(w http.ResponseWriter, r *http.Request) {
	lambda, ok := parseFloatParam(r, "lambda")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid lambda parameter")
		return
	}
	k, ok := parseIntParam(r, "k")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid k parameter (must be integer)")
		return
	}

	dist, err := probdist.NewPoisson(lambda)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cdf := dist.CDF(k)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: cdf})
}

func handlePoissonSample(w http.ResponseWriter, r *http.Request) {
	lambda, ok := parseFloatParam(r, "lambda")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid lambda parameter")
		return
	}
	n, ok := parseIntParam(r, "n")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid n parameter")
		return
	}

	dist, err := probdist.NewPoisson(lambda)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	samples, err := dist.Sample(n)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &api.IntSampleResponse{Samples: samples})
}

func handleUniformPDF(w http.ResponseWriter, r *http.Request) {
	a, ok := parseFloatParam(r, "a")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid a parameter")
		return
	}
	b, ok := parseFloatParam(r, "b")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid b parameter")
		return
	}
	x, ok := parseFloatParam(r, "x")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid x parameter")
		return
	}

	dist, err := probdist.NewUniform(a, b)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pdf := dist.PDF(x)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: pdf})
}

func handleUniformCDF(w http.ResponseWriter, r *http.Request) {
	a, ok := parseFloatParam(r, "a")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid a parameter")
		return
	}
	b, ok := parseFloatParam(r, "b")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid b parameter")
		return
	}
	x, ok := parseFloatParam(r, "x")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid x parameter")
		return
	}

	dist, err := probdist.NewUniform(a, b)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cdf := dist.CDF(x)
	writeJSON(w, http.StatusOK, &api.FloatResponse{Value: cdf})
}

func handleUniformSample(w http.ResponseWriter, r *http.Request) {
	a, ok := parseFloatParam(r, "a")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid a parameter")
		return
	}
	b, ok := parseFloatParam(r, "b")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid b parameter")
		return
	}
	n, ok := parseIntParam(r, "n")
	if !ok {
		writeError(w, http.StatusBadRequest, "missing or invalid n parameter")
		return
	}

	dist, err := probdist.NewUniform(a, b)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	samples, err := dist.Sample(n)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, &api.FloatSampleResponse{Samples: samples})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/normal/pdf", handleNormalPDF)
	mux.HandleFunc("/api/normal/cdf", handleNormalCDF)
	mux.HandleFunc("/api/normal/sample", handleNormalSample)

	mux.HandleFunc("/api/poisson/pdf", handlePoissonPDF)
	mux.HandleFunc("/api/poisson/cdf", handlePoissonCDF)
	mux.HandleFunc("/api/poisson/sample", handlePoissonSample)

	mux.HandleFunc("/api/uniform/pdf", handleUniformPDF)
	mux.HandleFunc("/api/uniform/cdf", handleUniformCDF)
	mux.HandleFunc("/api/uniform/sample", handleUniformSample)

	http.ListenAndServe(":8080", mux)
}
