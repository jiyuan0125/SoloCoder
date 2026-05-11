package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"geohash/pkg/api"
)

type client struct {
	serverURL string
}

func newClient(serverURL string) *client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &client{serverURL: serverURL}
}

func (c *client) post(endpoint string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(c.serverURL+endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		var errResp api.ErrorResponse
		if err := json.NewDecoder(httpResp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("服务器返回错误: %s", httpResp.Status)
		}
		return fmt.Errorf("错误: %s", errResp.Error)
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (c *client) encode(lat, lng float64, precision int) error {
	req := api.EncodeRequest{
		Latitude:  lat,
		Longitude: lng,
		Precision: precision,
	}

	var resp api.EncodeResponse
	if err := c.post("/encode", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("Geohash: %s\n", resp.Geohash)
	return nil
}

func (c *client) decode(geohash string) error {
	req := api.DecodeRequest{Geohash: geohash}

	var resp api.DecodeResponse
	if err := c.post("/decode", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("纬度: %f\n", resp.Latitude)
	fmt.Printf("经度: %f\n", resp.Longitude)
	return nil
}

func (c *client) bbox(geohash string) error {
	req := api.BoundingBoxRequest{Geohash: geohash}

	var resp api.BoundingBoxResponse
	if err := c.post("/bbox", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("最小纬度: %f\n", resp.MinLat)
	fmt.Printf("最大纬度: %f\n", resp.MaxLat)
	fmt.Printf("最小经度: %f\n", resp.MinLng)
	fmt.Printf("最大经度: %f\n", resp.MaxLng)
	return nil
}

func (c *client) neighbors(geohash string) error {
	req := api.NeighborsRequest{Geohash: geohash}

	var resp api.NeighborsResponse
	if err := c.post("/neighbors", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("北:     %s\n", resp.North)
	fmt.Printf("东北:   %s\n", resp.Northeast)
	fmt.Printf("东:     %s\n", resp.East)
	fmt.Printf("东南:   %s\n", resp.Southeast)
	fmt.Printf("南:     %s\n", resp.South)
	fmt.Printf("西南:   %s\n", resp.Southwest)
	fmt.Printf("西:     %s\n", resp.West)
	fmt.Printf("西北:   %s\n", resp.Northwest)
	return nil
}

func (c *client) search(lat, lng, radiusKm float64) error {
	req := api.ProximitySearchRequest{
		Latitude:  lat,
		Longitude: lng,
		RadiusKm:  radiusKm,
	}

	var resp api.ProximitySearchResponse
	if err := c.post("/search", &req, &resp); err != nil {
		return err
	}

	fmt.Printf("找到 %d 个 Geohash:\n", resp.Count)
	for i, gh := range resp.Geohashes {
		fmt.Printf("  %d. %s\n", i+1, gh)
	}
	return nil
}

func printUsage() {
	fmt.Println("Geohash 命令行客户端")
	fmt.Println()
	fmt.Println("用法: geohash-client [全局选项] <命令> [命令选项]")
	fmt.Println()
	fmt.Println("全局选项:")
	fmt.Println("  --server <地址>    服务端地址 (默认: localhost:8080)")
	fmt.Println("  --help             显示帮助信息")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  encode    经纬度编码为 Geohash")
	fmt.Println("  decode    Geohash 解码为经纬度")
	fmt.Println("  bbox      获取 Geohash 边界矩形")
	fmt.Println("  neighbors 获取 Geohash 的邻居")
	fmt.Println("  search    邻近搜索")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  geohash-client encode --lat 39.9 --lng 116.4 --precision 6")
	fmt.Println("  geohash-client decode --geohash wx4g0e")
	fmt.Println("  geohash-client bbox --geohash wx4g0e")
	fmt.Println("  geohash-client neighbors --geohash wx4g0e")
	fmt.Println("  geohash-client search --lat 39.9 --lng 116.4 --radius 5")
	fmt.Println("  geohash-client --server 192.168.1.100:8080 encode --lat 39.9 --lng 116.4")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	globalFlagSet := flag.NewFlagSet("global", flag.ContinueOnError)
	serverURL := globalFlagSet.String("server", "localhost:8080", "服务端地址")
	help := globalFlagSet.Bool("help", false, "显示帮助信息")

	var args []string
	for i, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "-") {
			if i+1 < len(os.Args[1:]) && !strings.HasPrefix(os.Args[1:][i+1], "-") {
				continue
			}
		} else {
			args = os.Args[1 : i+1]
			os.Args = append(os.Args[:1], os.Args[i+1:]...)
			break
		}
	}

	if err := globalFlagSet.Parse(args); err != nil {
		printUsage()
		os.Exit(1)
	}

	if *help {
		printUsage()
		os.Exit(0)
	}

	cmd := os.Args[1]
	client := newClient(*serverURL)

	cmdFlagSet := flag.NewFlagSet(cmd, flag.ContinueOnError)

	switch cmd {
	case "encode":
		lat := cmdFlagSet.Float64("lat", 0, "纬度")
		lng := cmdFlagSet.Float64("lng", 0, "经度")
		precision := cmdFlagSet.Int("precision", 6, "精度 (1-12)")

		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			printUsage()
			os.Exit(1)
		}

		if err := client.encode(*lat, *lng, *precision); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "decode":
		geohash := cmdFlagSet.String("geohash", "", "Geohash 字符串")

		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			printUsage()
			os.Exit(1)
		}

		if *geohash == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --geohash 参数")
			os.Exit(1)
		}

		if err := client.decode(*geohash); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "bbox":
		geohash := cmdFlagSet.String("geohash", "", "Geohash 字符串")

		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			printUsage()
			os.Exit(1)
		}

		if *geohash == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --geohash 参数")
			os.Exit(1)
		}

		if err := client.bbox(*geohash); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "neighbors":
		geohash := cmdFlagSet.String("geohash", "", "Geohash 字符串")

		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			printUsage()
			os.Exit(1)
		}

		if *geohash == "" {
			fmt.Fprintln(os.Stderr, "错误: 必须指定 --geohash 参数")
			os.Exit(1)
		}

		if err := client.neighbors(*geohash); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "search":
		lat := cmdFlagSet.Float64("lat", 0, "中心点纬度")
		lng := cmdFlagSet.Float64("lng", 0, "中心点经度")
		radius := cmdFlagSet.Float64("radius", 1, "搜索半径 (公里)")

		if err := cmdFlagSet.Parse(os.Args[2:]); err != nil {
			printUsage()
			os.Exit(1)
		}

		if err := client.search(*lat, *lng, *radius); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

	case "--help", "-h":
		printUsage()
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}
