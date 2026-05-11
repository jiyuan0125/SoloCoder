package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"currency-converter/api"
)

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, endpoint string, body io.Reader, result interface{}) error {
	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("server error: %s", resp.Status)
		}
		return fmt.Errorf(errResp.Error)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

func (c *Client) Convert(amount float64, from, to string) (*api.ConvertResponse, error) {
	req := api.ConvertRequest{
		Amount: amount,
		From:   api.Currency(strings.ToUpper(from)),
		To:     api.Currency(strings.ToUpper(to)),
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	var resp api.ConvertResponse
	if err := c.doRequest(http.MethodPost, "convert", bytes.NewReader(body), &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) GetCurrencies() ([]api.Currency, error) {
	var resp api.CurrenciesResponse
	if err := c.doRequest(http.MethodGet, "currencies", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Currencies, nil
}

func (c *Client) GetRates() ([]api.Rate, error) {
	var resp api.RatesResponse
	if err := c.doRequest(http.MethodGet, "rates", nil, &resp); err != nil {
		return nil, err
	}
	return resp.Rates, nil
}

func (c *Client) SetRates(rates []api.Rate) error {
	req := api.SetRatesRequest{Rates: rates}
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.doRequest(http.MethodPost, "rates/set", bytes.NewReader(body), nil)
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  convert <amount> <from> <to>     - Convert currency")
	fmt.Println("  convert currencies               - List supported currencies")
	fmt.Println("  convert rates                    - List current exchange rates")
	fmt.Println("  convert set <from> <to> <rate>   - Set a specific exchange rate")
	fmt.Println("\nOptions:")
	fmt.Println("  -server <url>                    - Server URL (default: localhost:8100)")
	fmt.Println("\nExamples:")
	fmt.Println("  convert 100 USD CNY")
	fmt.Println("  convert -server localhost:9090 1000 EUR GBP")
}

func main() {
	var serverURL string
	flag.StringVar(&serverURL, "server", "localhost:8100", "Server URL")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(serverURL)

	command := strings.ToLower(args[0])

	switch command {
	case "currencies":
		currencies, err := client.GetCurrencies()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Supported currencies:")
		for _, c := range currencies {
			fmt.Printf("  %s\n", c)
		}

	case "rates":
		rates, err := client.GetRates()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Current exchange rates:")
		for _, r := range rates {
			fmt.Printf("  %s -> %s: %f\n", r.From, r.To, r.Rate)
		}

	case "set":
		if len(args) < 4 {
			fmt.Println("Usage: convert set <from> <to> <rate>")
			os.Exit(1)
		}
		from := strings.ToUpper(args[1])
		to := strings.ToUpper(args[2])
		rate, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			fmt.Printf("Invalid rate: %v\n", err)
			os.Exit(1)
		}

		rates := []api.Rate{{
			From: api.Currency(from),
			To:   api.Currency(to),
			Rate: rate,
		}}
		if err := client.SetRates(rates); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully set rate: %s -> %s = %f\n", from, to, rate)

	default:
		if len(args) < 3 {
			printUsage()
			os.Exit(1)
		}

		amount, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			fmt.Printf("Invalid amount: %v\n", err)
			os.Exit(1)
		}

		from := args[1]
		to := args[2]

		resp, err := client.Convert(amount, from, to)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%.2f %s = %.2f %s\n", resp.Amount, resp.From, resp.Result, resp.To)
	}
}
