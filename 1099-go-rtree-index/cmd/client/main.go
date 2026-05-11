package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"rtree/internal/api"
	"rtree/internal/rtree"
)

const (
	DefaultServerURL = "http://localhost:8201"
)

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: serverURL}
}

func (c *Client) postJSON(endpoint string, body interface{}) ([]byte, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(c.serverURL+endpoint, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) getJSON(endpoint string) ([]byte, error) {
	resp, err := http.Get(c.serverURL + endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (c *Client) Add(objects []api.Object) error {
	body, err := c.postJSON("/api/objects/add", api.AddRequest{Objects: objects})
	if err != nil {
		return err
	}

	var resp api.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Printf("Successfully added objects\n")
	return nil
}

func (c *Client) Delete(id string) error {
	body, err := c.postJSON("/api/objects/delete", api.DeleteRequest{ID: id})
	if err != nil {
		return err
	}

	var resp api.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	fmt.Printf("Successfully deleted object: %s\n", id)
	return nil
}

func (c *Client) Search(minX, minY, maxX, maxY float64) error {
	url := fmt.Sprintf("/api/objects/search?x1=%.6f&y1=%.6f&x2=%.6f&y2=%.6f",
		minX, minY, maxX, maxY)
	body, err := c.getJSON(url)
	if err != nil {
		return err
	}

	var resp api.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var objects []api.Object
	json.Unmarshal(dataBytes, &objects)

	fmt.Printf("Found %d objects:\n", len(objects))
	for _, obj := range objects {
		fmt.Printf("  ID: %s, Bounds: [%.2f, %.2f, %.2f, %.2f]\n",
			obj.ID, obj.MinX, obj.MinY, obj.MaxX, obj.MaxY)
	}

	return nil
}

func (c *Client) KNN(x, y float64, k int) error {
	url := fmt.Sprintf("/api/objects/knn?x=%.6f&y=%.6f&k=%d", x, y, k)
	body, err := c.getJSON(url)
	if err != nil {
		return err
	}

	var resp api.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var objects []api.Object
	json.Unmarshal(dataBytes, &objects)

	fmt.Printf("Found %d nearest neighbors:\n", len(objects))
	for i, obj := range objects {
		fmt.Printf("  %d. ID: %s, Bounds: [%.2f, %.2f, %.2f, %.2f]\n",
			i+1, obj.ID, obj.MinX, obj.MinY, obj.MaxX, obj.MaxY)
	}

	return nil
}

func (c *Client) Info() error {
	body, err := c.getJSON("/api/tree/info")
	if err != nil {
		return err
	}

	var resp api.Response
	if err := json.Unmarshal(body, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("server error: %s", resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var info api.TreeInfo
	json.Unmarshal(dataBytes, &info)

	fmt.Printf("R-tree Info:\n")
	fmt.Printf("  Node Count: %d\n", info.NodeCount)
	fmt.Printf("  Height: %d\n", info.Height)
	fmt.Printf("  Object Count: %d\n", info.ObjectCount)
	fmt.Printf("  Node MBRs: %d\n", len(info.AllNodesMBR))

	return nil
}

func (c *Client) Visualize() error {
	body, err := c.getJSON("/api/tree/visualize")
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}

func readCSV(filename string) ([]api.Object, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var objects []api.Object
	for i, record := range records {
		if i == 0 && strings.ToLower(record[0]) == "id" {
			continue
		}

		if len(record) < 5 {
			continue
		}

		id := record[0]
		minX, _ := strconv.ParseFloat(record[1], 64)
		minY, _ := strconv.ParseFloat(record[2], 64)
		maxX, _ := strconv.ParseFloat(record[3], 64)
		maxY, _ := strconv.ParseFloat(record[4], 64)

		metadata := make(map[string]interface{})
		if len(record) > 5 {
			for j := 5; j < len(record); j++ {
				metadata[fmt.Sprintf("field_%d", j-5)] = record[j]
			}
		}

		objects = append(objects, api.Object{
			ID:       id,
			MinX:     minX,
			MinY:     minY,
			MaxX:     maxX,
			MaxY:     maxY,
			Metadata: metadata,
		})
	}

	return objects, nil
}

func generateRandom(count int) []api.Object {
	rand.Seed(time.Now().UnixNano())
	objects := make([]api.Object, 0, count)

	for i := 0; i < count; i++ {
		minX := rand.Float64() * 100
		minY := rand.Float64() * 100
		width := rand.Float64()*10 + 1
		height := rand.Float64()*10 + 1

		objects = append(objects, api.Object{
			ID:   fmt.Sprintf("obj_%d", i),
			MinX: minX,
			MinY: minY,
			MaxX: minX + width,
			MaxY: minY + height,
			Metadata: map[string]interface{}{
				"random": true,
			},
		})
	}

	return objects
}

func runBenchmark(sizes []int) {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Running benchmark...")
	fmt.Println("=" + strings.Repeat("-", 78) + "=")
	fmt.Printf("%-15s | %-15s | %-15s | %-15s | %-15s\n",
		"Size", "R-tree Insert", "Brute Insert", "R-tree Search", "Brute Search")
	fmt.Println("=" + strings.Repeat("-", 78) + "=")

	for _, size := range sizes {
		objects := generateRandom(size)
		queryX, queryY := 50.0, 50.0
		queryR := 10.0

		tree := rtree.New()
		start := time.Now()
		for _, obj := range objects {
			rObj := rtree.NewSimpleObject(obj.ID, obj.MinX, obj.MinY, obj.MaxX, obj.MaxY, obj.Metadata)
			tree.Insert(rObj)
		}
		treeInsertTime := time.Since(start)

		bruteList := make([]rtree.SpatialObject, 0, len(objects))
		start = time.Now()
		for _, obj := range objects {
			rObj := rtree.NewSimpleObject(obj.ID, obj.MinX, obj.MinY, obj.MaxX, obj.MaxY, obj.Metadata)
			bruteList = append(bruteList, rObj)
		}
		bruteInsertTime := time.Since(start)

		start = time.Now()
		tree.Search(queryX-queryR, queryY-queryR, queryX+queryR, queryY+queryR)
		treeSearchTime := time.Since(start)

		start = time.Now()
		queryMBR := rtree.NewMBR(queryX-queryR, queryY-queryR, queryX+queryR, queryY+queryR)
		bruteResults := 0
		for _, obj := range bruteList {
			mx, my, mX, mY := obj.Bounds()
			objMBR := rtree.NewMBR(mx, my, mX, mY)
			if objMBR.Intersects(queryMBR) {
				bruteResults++
			}
		}
		bruteSearchTime := time.Since(start)

		fmt.Printf("%-15d | %-15v | %-15v | %-15v | %-15v\n",
			size, treeInsertTime, bruteInsertTime, treeSearchTime, bruteSearchTime)
	}

	fmt.Println("=" + strings.Repeat("-", 78) + "=")
}

func usage() {
	fmt.Println("R-tree Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <csv_file>       Batch add objects from CSV file")
	fmt.Println("  delete <id>          Delete object by ID")
	fmt.Println("  search <x1> <y1> <x2> <y2>  Search objects in rectangle")
	fmt.Println("  knn <x> <y> <k>      Find k nearest neighbors")
	fmt.Println("  random <count>       Generate and add random objects")
	fmt.Println("  info                 Get R-tree info")
	fmt.Println("  visualize            Visualize R-tree structure")
	fmt.Println("  bench                Run performance benchmark")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client add objects.csv")
	fmt.Println("  client search 0 0 10 10")
	fmt.Println("  client knn 5 5 3")
	fmt.Println("  client random 100")
	fmt.Println()
}

func main() {
	serverURL := os.Getenv("RTREE_SERVER_URL")
	if serverURL == "" {
		serverURL = DefaultServerURL
	}

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	client := NewClient(serverURL)
	command := flag.Arg(0)

	var err error
	switch command {
	case "add":
		if flag.NArg() < 2 {
			usage()
			os.Exit(1)
		}
		filename := flag.Arg(1)
		objects, e := readCSV(filename)
		if e != nil {
			fmt.Fprintf(os.Stderr, "Error reading CSV: %v\n", e)
			os.Exit(1)
		}
		err = client.Add(objects)

	case "delete":
		if flag.NArg() < 2 {
			usage()
			os.Exit(1)
		}
		id := flag.Arg(1)
		err = client.Delete(id)

	case "search":
		if flag.NArg() < 5 {
			usage()
			os.Exit(1)
		}
		x1, _ := strconv.ParseFloat(flag.Arg(1), 64)
		y1, _ := strconv.ParseFloat(flag.Arg(2), 64)
		x2, _ := strconv.ParseFloat(flag.Arg(3), 64)
		y2, _ := strconv.ParseFloat(flag.Arg(4), 64)
		err = client.Search(x1, y1, x2, y2)

	case "knn":
		if flag.NArg() < 4 {
			usage()
			os.Exit(1)
		}
		x, _ := strconv.ParseFloat(flag.Arg(1), 64)
		y, _ := strconv.ParseFloat(flag.Arg(2), 64)
		k, _ := strconv.Atoi(flag.Arg(3))
		err = client.KNN(x, y, k)

	case "random":
		if flag.NArg() < 2 {
			usage()
			os.Exit(1)
		}
		count, _ := strconv.Atoi(flag.Arg(1))
		objects := generateRandom(count)
		err = client.Add(objects)

	case "info":
		err = client.Info()

	case "visualize":
		err = client.Visualize()

	case "bench":
		runBenchmark([]int{100, 1000, 10000})

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
