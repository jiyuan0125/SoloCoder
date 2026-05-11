package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"reservoir-sample/common"
)

const (
	defaultServer = "http://localhost:8510"
)

type Client struct {
	baseURL string
	httpCli *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpCli: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AddData(data string) error {
	reqBody, _ := json.Marshal(common.AddDataRequest{Data: data})
	resp, err := c.httpCli.Post(c.baseURL+"/data", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Sample(k *int) ([]string, error) {
	var reqBody io.Reader
	if k != nil {
		data, _ := json.Marshal(common.SampleRequest{K: k})
		reqBody = bytes.NewReader(data)
	}
	resp, err := c.httpCli.Post(c.baseURL+"/sample", "application/json", reqBody)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var result common.SampleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Samples, nil
}

func (c *Client) SetK(k int) error {
	reqBody, _ := json.Marshal(common.SetKRequest{K: k})
	resp, err := c.httpCli.Post(c.baseURL+"/k", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) Count() (int, error) {
	resp, err := c.httpCli.Get(c.baseURL + "/count")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	var result common.CountResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.Count, nil
}

func (c *Client) Clear() error {
	resp, err := c.httpCli.Post(c.baseURL+"/clear", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}

func main() {
	var serverAddr string
	var inputFile string
	var outputFile string
	var k int
	var simulate bool
	var count bool
	var clear bool

	flag.StringVar(&serverAddr, "server", defaultServer, "server address")
	flag.StringVar(&inputFile, "input", "", "input file (read lines from file)")
	flag.StringVar(&outputFile, "output", "", "output file (write sample results)")
	flag.IntVar(&k, "k", 0, "sample size (0 to use server default)")
	flag.BoolVar(&simulate, "simulate", false, "simulate data stream with 1000 test records")
	flag.BoolVar(&count, "count", false, "get current data count")
	flag.BoolVar(&clear, "clear", false, "clear all data")
	flag.Parse()

	client := NewClient(serverAddr)

	if clear {
		if err := client.Clear(); err != nil {
			fmt.Fprintf(os.Stderr, "clear error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("data cleared")
		return
	}

	if count {
		cnt, err := client.Count()
		if err != nil {
			fmt.Fprintf(os.Stderr, "count error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("current count: %d\n", cnt)
		return
	}

	if k > 0 {
		if err := client.SetK(k); err != nil {
			fmt.Fprintf(os.Stderr, "set k error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("k set to %d\n", k)
	}

	if simulate {
		fmt.Println("simulating data stream with 1000 test records...")
		for i := 0; i < 1000; i++ {
			data := fmt.Sprintf("log-entry-%05d", i)
			if err := client.AddData(data); err != nil {
				fmt.Fprintf(os.Stderr, "add data error: %v\n", err)
				os.Exit(1)
			}
			if (i+1)%100 == 0 {
				fmt.Printf("pushed %d records\n", i+1)
			}
		}
		fmt.Println("simulation complete")
	}

	if inputFile != "" {
		fmt.Printf("reading data from file: %s\n", inputFile)
		f, err := os.Open(inputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open input file error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		count := 0
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if err := client.AddData(line); err != nil {
				fmt.Fprintf(os.Stderr, "add data error: %v\n", err)
				os.Exit(1)
			}
			count++
			if count%100 == 0 {
				fmt.Printf("pushed %d lines\n", count)
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "read file error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("file read complete, total %d lines pushed\n", count)
	}

	var sampleK *int
	if k > 0 {
		sampleK = &k
	}
	samples, err := client.Sample(sampleK)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sample error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("sampled %d records\n", len(samples))

	if outputFile != "" {
		f, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create output file error: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		writer := bufio.NewWriter(f)
		for _, s := range samples {
			writer.WriteString(s + "\n")
		}
		writer.Flush()
		fmt.Printf("results written to %s\n", outputFile)
	} else {
		fmt.Println("\nsample results:")
		for i, s := range samples {
			fmt.Printf("[%d] %s\n", i, s)
		}
	}
}
