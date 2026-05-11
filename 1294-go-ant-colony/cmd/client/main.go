package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/example/antcolony/pkg/common"
)

func loadGraphFromFile(path string) (common.Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return common.Graph{}, err
	}

	var g common.Graph
	if err := json.Unmarshal(data, &g); err == nil {
		return g, nil
	}

	return parseDOT(string(data))
}

func parseDOT(content string) (common.Graph, error) {
	lines := strings.Split(content, "\n")
	nodesMap := make(map[string]struct{})
	var edges []common.Edge

	inGraph := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "digraph") || strings.HasPrefix(line, "graph") {
			inGraph = true
			continue
		}
		if line == "}" {
			inGraph = false
			continue
		}
		if !inGraph {
			continue
		}

		if strings.Contains(line, "->") {
			parts := strings.SplitN(line, "->", 2)
			if len(parts) != 2 {
				continue
			}
			from := strings.TrimSpace(parts[0])
			from = strings.Trim(from, "\";")

			restParts := strings.SplitN(parts[1], "[", 2)
			to := strings.TrimSpace(restParts[0])
			to = strings.Trim(to, "\";")

			weight := 1.0
			if len(restParts) > 1 {
				weightPart := restParts[1]
				if strings.Contains(weightPart, "weight=") {
					startIdx := strings.Index(weightPart, "weight=") + 7
					endIdx := strings.IndexAny(weightPart[startIdx:], "];, \t")
					if endIdx != -1 {
						endIdx += startIdx
					} else {
						endIdx = len(weightPart)
					}
					weightStr := strings.TrimSpace(weightPart[startIdx:endIdx])
					weightStr = strings.Trim(weightStr, "\"")
					var w float64
					fmt.Sscanf(weightStr, "%f", &w)
					if w > 0 {
						weight = w
					}
				}
			}

			nodesMap[from] = struct{}{}
			nodesMap[to] = struct{}{}
			edges = append(edges, common.Edge{From: from, To: to, Weight: weight})
		} else if !strings.Contains(line, "label=") && !strings.HasPrefix(line, "//") {
			nodeID := strings.TrimRight(line, ";")
			nodeID = strings.TrimSpace(nodeID)
			nodeID = strings.Trim(nodeID, "\"")
			if nodeID != "" && !strings.Contains(nodeID, "[") {
				nodesMap[nodeID] = struct{}{}
			}
		}
	}

	nodes := make([]common.Node, 0, len(nodesMap))
	for id := range nodesMap {
		nodes = append(nodes, common.Node{ID: id})
	}

	if len(nodes) == 0 {
		return common.Graph{}, fmt.Errorf("no nodes found in DOT file")
	}

	return common.Graph{Nodes: nodes, Edges: edges}, nil
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client -file <graph_file> -start <node> -end <node> [options]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -file <path>        Path to graph file (JSON or DOT format)")
	fmt.Println("  -start <node>       Start node ID")
	fmt.Println("  -end <node>         End node ID")
	fmt.Println("  -server <url>       Server URL (default: http://localhost:8103)")
	fmt.Println("  -ants <count>       Number of ants (default: 20)")
	fmt.Println("  -iters <count>      Max iterations (default: 100)")
	fmt.Println("  -evap <rate>        Evaporation rate (0-1, default: 0.1)")
	fmt.Println("  -alpha <value>      Pheromone weight (default: 1.0)")
	fmt.Println("  -beta <value>       Heuristic weight (default: 2.0)")
	fmt.Println("  -elite <weight>     Elite ant weight (default: 2.0)")
	fmt.Println("  -wait               Wait for completion and show result")
	fmt.Println("")
	fmt.Println("Graph file formats:")
	fmt.Println("  1. JSON adjacency table:")
	fmt.Println("     {")
	fmt.Println("       \"nodes\": [{\"id\": \"A\"}, {\"id\": \"B\"}],")
	fmt.Println("       \"edges\": [{\"from\": \"A\", \"to\": \"B\", \"weight\": 10}]")
	fmt.Println("     }")
	fmt.Println("  2. DOT format:")
	fmt.Println("     digraph G {")
	fmt.Println("       A -> B [weight=10];")
	fmt.Println("       B -> C [weight=5];")
	fmt.Println("     }")
}

func main() {
	filePath := flag.String("file", "", "Path to graph file")
	startNode := flag.String("start", "", "Start node")
	endNode := flag.String("end", "", "End node")
	serverURL := flag.String("server", "http://localhost:8103", "Server URL")
	ants := flag.Int("ants", 0, "Number of ants")
	iters := flag.Int("iters", 0, "Max iterations")
	evap := flag.Float64("evap", 0, "Evaporation rate")
	alpha := flag.Float64("alpha", 0, "Alpha parameter")
	beta := flag.Float64("beta", 0, "Beta parameter")
	elite := flag.Float64("elite", 0, "Elite weight")
	wait := flag.Bool("wait", false, "Wait for completion")
	flag.Parse()

	if *filePath == "" || *startNode == "" || *endNode == "" {
		printUsage()
		os.Exit(1)
	}

	graph, err := loadGraphFromFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading graph: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Loaded graph: %d nodes, %d edges\n", len(graph.Nodes), len(graph.Edges))

	req := common.SubmitRequest{
		Graph: graph,
		Start: *startNode,
		End:   *endNode,
	}

	if *ants > 0 {
		req.Parameters.AntCount = *ants
	}
	if *iters > 0 {
		req.Parameters.MaxIterations = *iters
	}
	if *evap != 0 {
		req.Parameters.EvaporationRate = *evap
	}
	if *alpha != 0 {
		req.Parameters.Alpha = *alpha
	}
	if *beta != 0 {
		req.Parameters.Beta = *beta
	}
	if *elite != 0 {
		req.Parameters.EliteWeight = *elite
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding request: %v\n", err)
		os.Exit(1)
	}

	submitResp, err := http.Post(*serverURL+"/submit", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error submitting task: %v\n", err)
		os.Exit(1)
	}
	defer submitResp.Body.Close()

	submitBody, _ := io.ReadAll(submitResp.Body)
	var submitResult common.SubmitResponse
	json.Unmarshal(submitBody, &submitResult)

	if !submitResult.Success {
		fmt.Fprintf(os.Stderr, "Submission failed: %s\n", submitResult.Message)
		os.Exit(1)
	}

	fmt.Printf("Task submitted. ID: %s\n", submitResult.TaskID)
	if submitResult.IsLarge {
		fmt.Println("Warning: Large graph detected, computation may be slow.")
	}

	_, err = http.Post(*serverURL+"/start?taskId="+submitResult.TaskID, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting task: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Task started.")

	if !*wait {
		fmt.Println("To check status: GET " + *serverURL + "/status?taskId=" + submitResult.TaskID)
		fmt.Println("To get result: GET " + *serverURL + "/result?taskId=" + submitResult.TaskID)
		return
	}

	fmt.Println("Waiting for completion...")
	for {
		statusResp, err := http.Get(*serverURL + "/status?taskId=" + submitResult.TaskID)
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		statusBody, _ := io.ReadAll(statusResp.Body)
		statusResp.Body.Close()

		var statusResult common.StatusResponse
		json.Unmarshal(statusBody, &statusResult)

		fmt.Printf("\rProgress: %.1f%% | Iteration: %d | Best length: %.4f",
			statusResult.Progress*100, statusResult.CurrentIteration, statusResult.BestLength)

		if statusResult.Completed {
			fmt.Println()
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	resultResp, err := http.Get(*serverURL + "/result?taskId=" + submitResult.TaskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting result: %v\n", err)
		os.Exit(1)
	}
	defer resultResp.Body.Close()

	resultBody, _ := io.ReadAll(resultResp.Body)
	var resultResult common.ResultResponse
	json.Unmarshal(resultBody, &resultResult)

	if !resultResult.Success {
		fmt.Printf("Solver failed: %s\n", resultResult.Error)
		os.Exit(1)
	}

	fmt.Println("\n=== Result ===")
	fmt.Printf("Status: success (iterations: %d, converged: %v)\n", resultResult.Iterations, resultResult.Converged)
	fmt.Printf("Path length: %.4f\n", resultResult.Path.Length)
	fmt.Printf("Path: %s\n", strings.Join(resultResult.Path.Nodes, " -> "))
}
