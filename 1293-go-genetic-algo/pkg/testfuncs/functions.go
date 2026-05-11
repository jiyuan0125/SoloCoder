package testfuncs

import (
	"errors"
	"math"
	"strings"
)

type TestFunction struct {
	Name         string
	Description  string
	Function     func([]float64) float64
	DefaultLower []float64
	DefaultUpper []float64
	Minimize     bool
	GlobalOptima func(dim int) [][]float64
	GlobalValue  func(dim int) float64
}

var TestFunctions = map[string]*TestFunction{
	"rastrigin": {
		Name:        "Rastrigin",
		Description: "Multimodal function with many local minima. Global minimum at x=0, f=0",
		Minimize:    true,
		GlobalValue: func(dim int) float64 { return 0 },
		GlobalOptima: func(dim int) [][]float64 {
			opt := make([]float64, dim)
			return [][]float64{opt}
		},
	},
	"sphere": {
		Name:        "Sphere",
		Description: "Unimodal convex function. Global minimum at x=0, f=0",
		Minimize:    true,
		GlobalValue: func(dim int) float64 { return 0 },
		GlobalOptima: func(dim int) [][]float64 {
			opt := make([]float64, dim)
			return [][]float64{opt}
		},
	},
	"rosenbrock": {
		Name:        "Rosenbrock",
		Description: "Valley-shaped function with narrow minimum. Global minimum at x=1, f=0",
		Minimize:    true,
		GlobalValue: func(dim int) float64 { return 0 },
		GlobalOptima: func(dim int) [][]float64 {
			opt := make([]float64, dim)
			for i := range opt {
				opt[i] = 1.0
			}
			return [][]float64{opt}
		},
	},
	"ackley": {
		Name:        "Ackley",
		Description: "Multimodal function with many local minima. Global minimum at x=0, f=0",
		Minimize:    true,
		GlobalValue: func(dim int) float64 { return 0 },
		GlobalOptima: func(dim int) [][]float64 {
			opt := make([]float64, dim)
			return [][]float64{opt}
		},
	},
	"griewank": {
		Name:        "Griewank",
		Description: "Multimodal function with many local minima. Global minimum at x=0, f=0",
		Minimize:    true,
		GlobalValue: func(dim int) float64 { return 0 },
		GlobalOptima: func(dim int) [][]float64 {
			opt := make([]float64, dim)
			return [][]float64{opt}
		},
	},
}

func init() {
	TestFunctions["rastrigin"].Function = Rastrigin
	TestFunctions["rastrigin"].DefaultLower = defaultBounds(-5.12)
	TestFunctions["rastrigin"].DefaultUpper = defaultBounds(5.12)

	TestFunctions["sphere"].Function = Sphere
	TestFunctions["sphere"].DefaultLower = defaultBounds(-100.0)
	TestFunctions["sphere"].DefaultUpper = defaultBounds(100.0)

	TestFunctions["rosenbrock"].Function = Rosenbrock
	TestFunctions["rosenbrock"].DefaultLower = defaultBounds(-30.0)
	TestFunctions["rosenbrock"].DefaultUpper = defaultBounds(30.0)

	TestFunctions["ackley"].Function = Ackley
	TestFunctions["ackley"].DefaultLower = defaultBounds(-32.768)
	TestFunctions["ackley"].DefaultUpper = defaultBounds(32.768)

	TestFunctions["griewank"].Function = Griewank
	TestFunctions["griewank"].DefaultLower = defaultBounds(-600.0)
	TestFunctions["griewank"].DefaultUpper = defaultBounds(600.0)
}

func defaultBounds(val float64) []float64 {
	bounds := make([]float64, 100)
	for i := range bounds {
		bounds[i] = val
	}
	return bounds
}

func GetFunction(name string) (*TestFunction, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	f, ok := TestFunctions[name]
	if !ok {
		return nil, errors.New("unknown test function: " + name)
	}
	return f, nil
}

func GetDefaultBounds(name string, dimensions int) (lower, upper []float64, err error) {
	f, err := GetFunction(name)
	if err != nil {
		return nil, nil, err
	}
	lower = make([]float64, dimensions)
	upper = make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		lower[i] = f.DefaultLower[i%len(f.DefaultLower)]
		upper[i] = f.DefaultUpper[i%len(f.DefaultUpper)]
	}
	return lower, upper, nil
}

func ListFunctions() []string {
	names := make([]string, 0, len(TestFunctions))
	for name := range TestFunctions {
		names = append(names, name)
	}
	return names
}

func Rastrigin(x []float64) float64 {
	n := len(x)
	sum := 10.0 * float64(n)
	for _, xi := range x {
		sum += xi*xi - 10.0*math.Cos(2.0*math.Pi*xi)
	}
	return sum
}

func Sphere(x []float64) float64 {
	sum := 0.0
	for _, xi := range x {
		sum += xi * xi
	}
	return sum
}

func Rosenbrock(x []float64) float64 {
	sum := 0.0
	for i := 0; i < len(x)-1; i++ {
		a := 1.0 - x[i]
		b := x[i+1] - x[i]*x[i]
		sum += a*a + 100.0*b*b
	}
	return sum
}

func Ackley(x []float64) float64 {
	n := float64(len(x))
	sum1 := 0.0
	sum2 := 0.0
	for _, xi := range x {
		sum1 += xi * xi
		sum2 += math.Cos(2.0 * math.Pi * xi)
	}
	return -20.0*math.Exp(-0.2*math.Sqrt(sum1/n)) - math.Exp(sum2/n) + 20.0 + math.E
}

func Griewank(x []float64) float64 {
	sum := 0.0
	product := 1.0
	for i, xi := range x {
		sum += xi * xi / 4000.0
		product *= math.Cos(xi / math.Sqrt(float64(i+1)))
	}
	return sum - product + 1.0
}
