package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"genetic-algo/cmd/client/client"
	"genetic-algo/pkg/api"
	"genetic-algo/pkg/ga"
	"genetic-algo/pkg/testfuncs"
)

type Args struct {
	mode        string
	server      string
	function    string
	dimensions  int
	popSize     int
	maxGens     int
	eliteCount  int
	crossover   float64
	mutation    float64
	sbxIndex    float64
	initSigma   float64
	finalSigma  float64
	showStats   bool
	statsFile   string
	interval    int
	taskID      string
	variables   string
	expr        string
	lower       string
	upper       string
}

func main() {
	args := parseArgs()

	switch args.mode {
	case "list":
		listFunctions(args)
	case "local":
		runLocal(args)
	case "remote":
		runRemote(args)
	case "status":
		checkStatus(args)
	case "stats":
		getStats(args)
	default:
		fmt.Println("Unknown mode. Use 'list', 'local', 'remote', 'status', or 'stats'")
		os.Exit(1)
	}
}

func parseArgs() *Args {
	args := &Args{}

	fs := flag.NewFlagSet("ga-client", flag.ExitOnError)

	fs.StringVar(&args.mode, "mode", "local", "Mode: list, local, remote, status, stats")
	fs.StringVar(&args.server, "server", "http://localhost:8102", "Server URL")
	fs.StringVar(&args.function, "function", "rastrigin", "Test function: rastrigin, sphere, rosenbrock, ackley, griewank")
	fs.IntVar(&args.dimensions, "dim", 5, "Dimensions")
	fs.IntVar(&args.popSize, "popsize", 0, "Population size")
	fs.IntVar(&args.maxGens, "gens", 100, "Maximum generations")
	fs.IntVar(&args.eliteCount, "elite", 2, "Elite count")
	fs.Float64Var(&args.crossover, "cx", 0.8, "Crossover rate")
	fs.Float64Var(&args.mutation, "mut", 0.1, "Mutation rate")
	fs.Float64Var(&args.sbxIndex, "sbx", 20.0, "SBX distribution index")
	fs.Float64Var(&args.initSigma, "init-sigma", 0.5, "Initial mutation sigma")
	fs.Float64Var(&args.finalSigma, "final-sigma", 0.01, "Final mutation sigma")
	fs.BoolVar(&args.showStats, "show-stats", false, "Show convergence statistics")
	fs.StringVar(&args.statsFile, "stats-file", "", "Save statistics to file")
	fs.IntVar(&args.interval, "interval", 500, "Poll interval in ms (remote mode)")
	fs.StringVar(&args.taskID, "task-id", "", "Task ID for status/stats mode")
	fs.StringVar(&args.variables, "vars", "", "Variables for custom function (comma-separated)")
	fs.StringVar(&args.expr, "expr", "", "Custom function expression")
	fs.StringVar(&args.lower, "lower", "", "Lower bounds (comma-separated)")
	fs.StringVar(&args.upper, "upper", "", "Upper bounds (comma-separated)")

	fs.Parse(os.Args[1:])

	return args
}

func listFunctions(args *Args) {
	fmt.Println("Available test functions:")
	fmt.Println("")

	for _, name := range testfuncs.ListFunctions() {
		fn, err := testfuncs.GetFunction(name)
		if err != nil {
			continue
		}
		fmt.Printf("  %s\n", name)
		fmt.Printf("    %s\n", fn.Description)
		fmt.Printf("    Default bounds: [%.2f, %.2f]\n", fn.DefaultLower[0], fn.DefaultUpper[0])
		fmt.Printf("    Global minimum: f(x)=0 at x=0\n")
		fmt.Println("")
	}
}

func runLocal(args *Args) {
	var obj func([]float64) float64
	var lower, upper []float64

	if args.expr != "" {
		var vars []string
		if args.variables != "" {
			vars = splitStrings(args.variables)
		} else {
			vars = make([]string, args.dimensions)
			for i := 0; i < args.dimensions; i++ {
				vars[i] = string(rune('x' + i))
			}
		}
		if len(vars) != args.dimensions {
			fmt.Printf("Error: variables count (%d) must match dimensions (%d)\n", len(vars), args.dimensions)
			os.Exit(1)
		}

		customFn, err := testfuncs.ParseCustomFunction(args.expr, vars)
		if err != nil {
			fmt.Printf("Error parsing custom function: %v\n", err)
			os.Exit(1)
		}
		obj = customFn.Evaluate

		lower, upper = parseBounds(args.lower, args.upper, args.dimensions, -10.0, 10.0)
	} else {
		fn, err := testfuncs.GetFunction(args.function)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		obj = fn.Function

		defLower, defUpper, _ := testfuncs.GetDefaultBounds(args.function, args.dimensions)
		lower, upper = parseBounds(args.lower, args.upper, args.dimensions, defLower[0], defUpper[0])
	}

	problem := &ga.Problem{
		Dimensions: args.dimensions,
		LowerBound: lower,
		UpperBound: upper,
		Minimize:   true,
	}

	config := ga.DefaultConfig(args.dimensions)
	config.MaxGenerations = args.maxGens
	config.EliteCount = args.eliteCount
	config.CrossoverRate = args.crossover
	config.MutationRate = args.mutation
	config.SBXIndex = args.sbxIndex
	config.InitialMutationSigma = args.initSigma
	config.FinalMutationSigma = args.finalSigma
	if args.popSize > 1 {
		config.PopulationSize = args.popSize
	}

	if err := config.Validate(problem); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Genetic Algorithm Optimization (Local) ===")
	fmt.Printf("Function:     %s\n", funcName(args))
	fmt.Printf("Dimensions:   %d\n", args.dimensions)
	fmt.Printf("Population:   %d\n", config.PopulationSize)
	fmt.Printf("Generations:  %d\n", config.MaxGenerations)
	fmt.Printf("Elite count:  %d\n", config.EliteCount)
	fmt.Printf("Crossover:    %.2f\n", config.CrossoverRate)
	fmt.Printf("Mutation:     %.2f\n", config.MutationRate)
	fmt.Printf("SBX index:    %.1f\n", config.SBXIndex)
	fmt.Printf("Initial sigma: %.2f\n", config.InitialMutationSigma)
	fmt.Printf("Final sigma:   %.3f\n", config.FinalMutationSigma)
	fmt.Printf("Bounds:       [%.2f, %.2f]\n", lower[0], upper[0])
	fmt.Println("")

	startTime := time.Now()
	alg, err := ga.NewAlgorithm(problem, config, obj)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	best, stats, err := alg.Run()
	elapsed := time.Since(startTime)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("=== Results ===")
	fmt.Printf("Best fitness: %.15g\n", best.Fitness)
	fmt.Printf("Best solution: ")
	if len(best.Genes) <= 10 {
		for i, g := range best.Genes {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%.6f", g)
		}
	} else {
		fmt.Printf("[%d dimensions, first 5: ", len(best.Genes))
		for i := 0; i < 5; i++ {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%.6f", best.Genes[i])
		}
		fmt.Printf(", ...]")
	}
	fmt.Println("")
	fmt.Printf("Generations:  %d\n", len(stats))
	fmt.Printf("Elapsed time: %v\n", elapsed)

	if args.showStats {
		fmt.Println("")
		fmt.Println("=== Convergence Statistics ===")
		fmt.Println("Gen |   Best Fitness   |   Mean Fitness   |    Std Dev")
		fmt.Println("--------------------------------------------------------")
		step := max(1, len(stats)/20)
		for i := 0; i < len(stats); i += step {
			s := stats[i]
			fmt.Printf("%3d | %14.6g | %14.6g | %12.6g\n",
				s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
		}
		if len(stats) > 0 {
			s := stats[len(stats)-1]
			fmt.Printf("%3d | %14.6g | %14.6g | %12.6g\n",
				s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
		}
	}

	if args.statsFile != "" {
		if err := saveStats(args.statsFile, convertGAStats(stats)); err != nil {
			fmt.Printf("Warning: could not save stats: %v\n", err)
		} else {
			fmt.Printf("\nStatistics saved to: %s\n", args.statsFile)
		}
	}
}

func runRemote(args *Args) {
	c := client.New(args.server)

	lower, upper := parseBounds(args.lower, args.upper, args.dimensions, 0, 0)

	req := &api.OptimizeRequest{
		Problem: &api.ProblemSpec{
			Function:   args.function,
			CustomExpr: args.expr,
			Variables:  splitStrings(args.variables),
			Dimensions: args.dimensions,
			LowerBound: lower,
			UpperBound: upper,
			Minimize:   true,
		},
		Config: &api.AlgorithmConfig{
			PopulationSize:       args.popSize,
			MaxGenerations:       args.maxGens,
			EliteCount:           args.eliteCount,
			CrossoverRate:        args.crossover,
			MutationRate:         args.mutation,
			SBXIndex:             args.sbxIndex,
			InitialMutationSigma: args.initSigma,
			FinalMutationSigma:   args.finalSigma,
		},
	}

	fmt.Println("=== Genetic Algorithm Optimization (Remote) ===")
	fmt.Printf("Server:       %s\n", args.server)
	fmt.Printf("Function:     %s\n", funcName(args))
	fmt.Printf("Dimensions:   %d\n", args.dimensions)
	if args.popSize > 1 {
		fmt.Printf("Population:   %d\n", args.popSize)
	}
	fmt.Printf("Generations:  %d\n", args.maxGens)
	fmt.Println("")

	task, err := c.SubmitTask(req)
	if err != nil {
		fmt.Printf("Error submitting task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task ID:      %s\n", task.TaskID)
	fmt.Printf("Status:       %s\n", task.Status)
	fmt.Println("")

	interval := time.Duration(args.interval) * time.Millisecond
	var lastGen int = -1
	var finalResult *api.TaskResult

	for {
		result, err := c.GetTask(task.TaskID)
		if err != nil {
			fmt.Printf("\nError checking status: %v\n", err)
			os.Exit(1)
		}

		if result.CurrentGen != lastGen {
			lastGen = result.CurrentGen
			if result.TotalGens > 0 {
				fmt.Printf("\rProgress:     %d/%d gens, best: %.6g", result.CurrentGen, result.TotalGens, result.BestFitness)
			} else {
				fmt.Printf("\rProgress:     %d gens, best: %.6g", result.CurrentGen, result.BestFitness)
			}
		}

		if result.Status == api.TaskStatusCompleted ||
			result.Status == api.TaskStatusFailed ||
			result.Status == api.TaskStatusCancelled {
			finalResult = result
			fmt.Println("")
			break
		}

		time.Sleep(interval)
	}

	fmt.Println("")
	fmt.Println("=== Results ===")
	fmt.Printf("Status:       %s\n", finalResult.Status)
	if finalResult.Error != "" {
		fmt.Printf("Error:        %s\n", finalResult.Error)
		os.Exit(1)
	}
	fmt.Printf("Best fitness: %.15g\n", finalResult.BestFitness)
	fmt.Printf("Best solution: ")
	if len(finalResult.BestGenes) <= 10 {
		for i, g := range finalResult.BestGenes {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%.6f", g)
		}
	} else {
		fmt.Printf("[%d dimensions, first 5: ", len(finalResult.BestGenes))
		for i := 0; i < 5; i++ {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%.6f", finalResult.BestGenes[i])
		}
		fmt.Printf(", ...]")
	}
	fmt.Println("")
	fmt.Printf("Generations:  %d\n", finalResult.CurrentGen)

	if args.showStats {
		stats, err := c.GetStatistics(task.TaskID)
		if err != nil {
			fmt.Printf("\nWarning: could not get statistics: %v\n", err)
		} else {
			fmt.Println("")
			fmt.Println("=== Convergence Statistics ===")
			fmt.Println("Gen |   Best Fitness   |   Mean Fitness   |    Std Dev")
			fmt.Println("--------------------------------------------------------")
			step := max(1, len(stats)/20)
			for i := 0; i < len(stats); i += step {
				s := stats[i]
				fmt.Printf("%3d | %14.6g | %14.6g | %12.6g\n",
					s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
			}
			if len(stats) > 0 {
				s := stats[len(stats)-1]
				fmt.Printf("%3d | %14.6g | %14.6g | %12.6g\n",
					s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
			}
		}
	}

	if args.statsFile != "" {
		stats, err := c.GetStatistics(task.TaskID)
		if err != nil {
			fmt.Printf("\nWarning: could not get statistics: %v\n", err)
		} else if err := saveStats(args.statsFile, convertAPIStats(stats)); err != nil {
			fmt.Printf("\nWarning: could not save stats: %v\n", err)
		} else {
			fmt.Printf("\nStatistics saved to: %s\n", args.statsFile)
		}
	}
}

func checkStatus(args *Args) {
	if args.taskID == "" {
		fmt.Println("Error: -task-id is required")
		os.Exit(1)
	}

	c := client.New(args.server)
	result, err := c.GetTask(args.taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task ID:      %s\n", result.TaskID)
	fmt.Printf("Status:       %s\n", result.Status)
	fmt.Printf("Current gen:  %d\n", result.CurrentGen)
	if result.TotalGens > 0 {
		fmt.Printf("Total gens:   %d\n", result.TotalGens)
	}
	if result.Status == api.TaskStatusCompleted {
		fmt.Printf("Best fitness: %.15g\n", result.BestFitness)
		fmt.Printf("Best genes:   %v\n", result.BestGenes)
	}
	if result.Error != "" {
		fmt.Printf("Error:        %s\n", result.Error)
	}
}

func getStats(args *Args) {
	if args.taskID == "" {
		fmt.Println("Error: -task-id is required")
		os.Exit(1)
	}

	c := client.New(args.server)
	stats, err := c.GetStatistics(args.taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if args.statsFile != "" {
		if err := saveStats(args.statsFile, convertAPIStats(stats)); err != nil {
			fmt.Printf("Error saving stats: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Statistics saved to: %s\n", args.statsFile)
		return
	}

	fmt.Println("=== Convergence Statistics ===")
	fmt.Println("Gen |   Best Fitness   |   Mean Fitness   |    Std Dev")
	fmt.Println("--------------------------------------------------------")
	for _, s := range stats {
		fmt.Printf("%3d | %14.6g | %14.6g | %12.6g\n",
			s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
	}
}

func splitStrings(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func parseBounds(lowerStr, upperStr string, dim int, defLower, defUpper float64) ([]float64, []float64) {
	lower := make([]float64, dim)
	upper := make([]float64, dim)

	lowerParts := splitStrings(lowerStr)
	upperParts := splitStrings(upperStr)

	for i := 0; i < dim; i++ {
		if i < len(lowerParts) {
			fmt.Sscanf(lowerParts[i], "%f", &lower[i])
		} else if len(lowerParts) > 0 {
			fmt.Sscanf(lowerParts[0], "%f", &lower[i])
		} else {
			lower[i] = defLower
		}

		if i < len(upperParts) {
			fmt.Sscanf(upperParts[i], "%f", &upper[i])
		} else if len(upperParts) > 0 {
			fmt.Sscanf(upperParts[0], "%f", &upper[i])
		} else {
			upper[i] = defUpper
		}
	}

	return lower, upper
}

func funcName(args *Args) string {
	if args.expr != "" {
		return "custom: " + args.expr
	}
	return args.function
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type statRecord struct {
	Generation   int
	BestFitness  float64
	MeanFitness  float64
	StdDeviation float64
}

func convertGAStats(stats []ga.Statistics) []statRecord {
	result := make([]statRecord, len(stats))
	for i, s := range stats {
		result[i] = statRecord{
			Generation:   s.Generation,
			BestFitness:  s.BestFitness,
			MeanFitness:  s.MeanFitness,
			StdDeviation: s.StdDeviation,
		}
	}
	return result
}

func convertAPIStats(stats []api.Statistics) []statRecord {
	result := make([]statRecord, len(stats))
	for i, s := range stats {
		result[i] = statRecord{
			Generation:   s.Generation,
			BestFitness:  s.BestFitness,
			MeanFitness:  s.MeanFitness,
			StdDeviation: s.StdDeviation,
		}
	}
	return result
}

func saveStats(filename string, stats []statRecord) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintln(f, "generation,best_fitness,mean_fitness,std_deviation")
	for _, s := range stats {
		fmt.Fprintf(f, "%d,%.15g,%.15g,%.15g\n",
			s.Generation, s.BestFitness, s.MeanFitness, s.StdDeviation)
	}
	return nil
}
