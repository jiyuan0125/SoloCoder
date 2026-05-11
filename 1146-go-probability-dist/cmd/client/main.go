package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"prob-dist/pkg/api"
)

const serverURL = "http://localhost:8080"

func makeRequest(endpoint string, params map[string]string, result interface{}) error {
	u, err := url.Parse(serverURL + endpoint)
	if err != nil {
		return err
	}

	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		json.Unmarshal(body, &errResp)
		return fmt.Errorf("server error: %s", errResp.Error)
	}

	return json.Unmarshal(body, result)
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: prob <distribution> <operation> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Distributions: normal, poisson, uniform")
	fmt.Fprintln(os.Stderr, "Operations: pdf, cdf, sample")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Normal distribution options:")
	fmt.Fprintln(os.Stderr, "  --mu float       Mean (default 0)")
	fmt.Fprintln(os.Stderr, "  --sigma float    Standard deviation (default 1)")
	fmt.Fprintln(os.Stderr, "  --x float        Value for pdf/cdf")
	fmt.Fprintln(os.Stderr, "  --n int          Sample count for sample")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Poisson distribution options:")
	fmt.Fprintln(os.Stderr, "  --lambda float   Rate parameter (default 1)")
	fmt.Fprintln(os.Stderr, "  --k int          Value for pdf/cdf")
	fmt.Fprintln(os.Stderr, "  --n int          Sample count for sample")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Uniform distribution options:")
	fmt.Fprintln(os.Stderr, "  --a float        Lower bound (default 0)")
	fmt.Fprintln(os.Stderr, "  --b float        Upper bound (default 1)")
	fmt.Fprintln(os.Stderr, "  --x float        Value for pdf/cdf")
	fmt.Fprintln(os.Stderr, "  --n int          Sample count for sample")
	os.Exit(1)
}

func parseFloat(name string, flags *flag.FlagSet) float64 {
	f := flags.Lookup(name)
	if f == nil {
		return 0
	}
	v, _ := strconv.ParseFloat(f.Value.String(), 64)
	return v
}

func parseInt(name string, flags *flag.FlagSet) int {
	f := flags.Lookup(name)
	if f == nil {
		return 0
	}
	v, _ := strconv.Atoi(f.Value.String())
	return v
}

func main() {
	if len(os.Args) < 3 {
		usage()
	}

	distribution := strings.ToLower(os.Args[1])
	operation := strings.ToLower(os.Args[2])

	args := os.Args[3:]

	var endpoint string
	var params map[string]string

	switch distribution {
	case "normal":
		fs := flag.NewFlagSet("normal", flag.ContinueOnError)
		mu := fs.Float64("mu", 0, "Mean")
		sigma := fs.Float64("sigma", 1, "Standard deviation")
		x := fs.Float64("x", 0, "Value")
		n := fs.Int("n", 1, "Sample count")
		fs.Parse(args)

		params = map[string]string{
			"mu":    fmt.Sprintf("%g", *mu),
			"sigma": fmt.Sprintf("%g", *sigma),
		}

		switch operation {
		case "pdf":
			endpoint = "/api/normal/pdf"
			params["x"] = fmt.Sprintf("%g", *x)
		case "cdf":
			endpoint = "/api/normal/cdf"
			params["x"] = fmt.Sprintf("%g", *x)
		case "sample":
			endpoint = "/api/normal/sample"
			params["n"] = fmt.Sprintf("%d", *n)
		default:
			fmt.Fprintf(os.Stderr, "Unknown operation: %s\n", operation)
			usage()
		}

	case "poisson":
		fs := flag.NewFlagSet("poisson", flag.ContinueOnError)
		lambda := fs.Float64("lambda", 1, "Rate parameter")
		k := fs.Int("k", 0, "Value")
		n := fs.Int("n", 1, "Sample count")
		fs.Parse(args)

		params = map[string]string{
			"lambda": fmt.Sprintf("%g", *lambda),
		}

		switch operation {
		case "pdf":
			endpoint = "/api/poisson/pdf"
			params["k"] = fmt.Sprintf("%d", *k)
		case "cdf":
			endpoint = "/api/poisson/cdf"
			params["k"] = fmt.Sprintf("%d", *k)
		case "sample":
			endpoint = "/api/poisson/sample"
			params["n"] = fmt.Sprintf("%d", *n)
		default:
			fmt.Fprintf(os.Stderr, "Unknown operation: %s\n", operation)
			usage()
		}

	case "uniform":
		fs := flag.NewFlagSet("uniform", flag.ContinueOnError)
		a := fs.Float64("a", 0, "Lower bound")
		b := fs.Float64("b", 1, "Upper bound")
		x := fs.Float64("x", 0, "Value")
		n := fs.Int("n", 1, "Sample count")
		fs.Parse(args)

		params = map[string]string{
			"a": fmt.Sprintf("%g", *a),
			"b": fmt.Sprintf("%g", *b),
		}

		switch operation {
		case "pdf":
			endpoint = "/api/uniform/pdf"
			params["x"] = fmt.Sprintf("%g", *x)
		case "cdf":
			endpoint = "/api/uniform/cdf"
			params["x"] = fmt.Sprintf("%g", *x)
		case "sample":
			endpoint = "/api/uniform/sample"
			params["n"] = fmt.Sprintf("%d", *n)
		default:
			fmt.Fprintf(os.Stderr, "Unknown operation: %s\n", operation)
			usage()
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown distribution: %s\n", distribution)
		usage()
	}

	if operation == "sample" {
		if distribution == "poisson" {
			var result api.IntSampleResponse
			if err := makeRequest(endpoint, params, &result); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			for _, s := range result.Samples {
				fmt.Println(s)
			}
		} else {
			var result api.FloatSampleResponse
			if err := makeRequest(endpoint, params, &result); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			for _, s := range result.Samples {
				fmt.Printf("%.6g\n", s)
			}
		}
	} else {
		var result api.FloatResponse
		if err := makeRequest(endpoint, params, &result); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%.6g\n", result.Value)
	}
}
