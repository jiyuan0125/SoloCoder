package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"mac-oui-lookup/pkg/api"
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) doGet(endpoint string, query map[string]string, resp interface{}) error {
	u := c.BaseURL + endpoint
	if len(query) > 0 {
		q := url.Values{}
		for k, v := range query {
			q.Set(k, v)
		}
		u += "?" + q.Encode()
	}

	httpResp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, resp)
}

func (c *Client) doPost(endpoint string, req interface{}, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpResp, err := http.Post(c.BaseURL+endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	respBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, resp)
}

func (c *Client) Parse(mac string) (*api.ParseResponse, error) {
	resp := &api.ParseResponse{}
	err := c.doGet("/parse", map[string]string{"mac": mac}, resp)
	return resp, err
}

func (c *Client) Format(mac, format string) (*api.FormatResponse, error) {
	resp := &api.FormatResponse{}
	err := c.doGet("/format", map[string]string{"mac": mac, "format": format}, resp)
	return resp, err
}

func (c *Client) OUI(mac string) (*api.OUIResponse, error) {
	resp := &api.OUIResponse{}
	err := c.doGet("/oui", map[string]string{"mac": mac}, resp)
	return resp, err
}

func (c *Client) Types(mac string) (*api.TypesResponse, error) {
	resp := &api.TypesResponse{}
	err := c.doGet("/types", map[string]string{"mac": mac}, resp)
	return resp, err
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	format := flag.String("format", "colon", "Output format: colon, dash, dot")
	ouiFlag := flag.Bool("oui", false, "Query OUI vendor")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <MAC-address>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	mac := flag.Arg(0)
	client := NewClient(*serverURL)

	if *ouiFlag {
		resp, err := client.OUI(mac)
		if err != nil {
			printError(err.Error())
		}
		if !resp.Success {
			printError(resp.Error)
		}
		if resp.Vendor == "" {
			fmt.Printf("OUI: %s\nVendor: unknown\n", resp.OUI)
		} else {
			fmt.Printf("OUI: %s\nVendor: %s\n", resp.OUI, resp.Vendor)
		}
		return
	}

	if *format != "colon" {
		resp, err := client.Format(mac, *format)
		if err != nil {
			printError(err.Error())
		}
		if !resp.Success {
			printError(resp.Error)
		}
		fmt.Println(resp.Result)
		return
	}

	resp, err := client.Parse(mac)
	if err != nil {
		printError(err.Error())
	}
	if !resp.Success {
		printError(resp.Error)
	}

	fmt.Printf("Address: %s\n", resp.MAC.Address)
	fmt.Printf("OUI: %s\n", resp.MAC.OUI)
	if resp.MAC.Vendor != "" {
		fmt.Printf("Vendor: %s\n", resp.MAC.Vendor)
	}
	fmt.Printf("Types: %s\n", strings.Join(resp.MAC.Types, ", "))
	fmt.Printf("Multicast: %t\n", resp.MAC.IsMulticast)
	fmt.Printf("Locally Administered: %t\n", resp.MAC.IsLocallyAdmin)
	fmt.Printf("Globally Unique: %t\n", resp.MAC.IsGloballyUnique)
}
