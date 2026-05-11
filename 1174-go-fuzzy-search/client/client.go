package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fuzzysearch/api"
)

type FuzzyClient struct {
	baseURL string
}

func NewFuzzyClient(baseURL string) *FuzzyClient {
	return &FuzzyClient{
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

func (c *FuzzyClient) AddWord(word string) error {
	reqBody, err := json.Marshal(api.AddWordRequest{Word: word})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/add", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.AddWordResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Word '%s' added successfully\n", word)
	} else {
		fmt.Printf("Failed to add word '%s': %s\n", word, result.Message)
	}

	return nil
}

func (c *FuzzyClient) RemoveWord(word string) error {
	reqBody, err := json.Marshal(api.RemoveWordRequest{Word: word})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/remove", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.RemoveWordResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Word '%s' removed successfully\n", word)
	} else {
		fmt.Printf("Failed to remove word '%s': %s\n", word, result.Message)
	}

	return nil
}

func (c *FuzzyClient) ImportWords(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	if len(words) == 0 {
		return fmt.Errorf("no words found in file")
	}

	reqBody, err := json.Marshal(api.ImportWordsRequest{Words: words})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/import", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.ImportWordsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Imported %d/%d words successfully\n", result.Imported, result.Total)
	} else {
		fmt.Printf("Failed to import words: %s\n", result.Message)
	}

	return nil
}

func (c *FuzzyClient) Search(query string, threshold int) error {
	reqBody, err := json.Marshal(api.SearchRequest{Query: query, Threshold: threshold})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/search", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.SearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Found %d matches for query '%s' (threshold: %d)\n", len(result.Results), query, threshold)
		for i, r := range result.Results {
			fmt.Printf("%d. '%s' (distance: %d)\n", i+1, r.Word, r.Distance)
		}
	} else {
		fmt.Printf("Search failed: %s\n", result.Message)
	}

	return nil
}

func (c *FuzzyClient) WildcardSearch(pattern string, threshold int) error {
	reqBody, err := json.Marshal(api.WildcardSearchRequest{Pattern: pattern, Threshold: threshold})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/wildcard", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.WildcardSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Found %d matches for pattern '%s' (threshold: %d)\n", len(result.Results), pattern, threshold)
		for i, r := range result.Results {
			fmt.Printf("%d. '%s' (distance: %d)\n", i+1, r.Word, r.Distance)
		}
	} else {
		fmt.Printf("Wildcard search failed: %s\n", result.Message)
	}

	return nil
}

func (c *FuzzyClient) GetDictionary() error {
	resp, err := http.Get(c.baseURL + "/dict")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server error: %s", string(body))
	}

	var result api.GetDictionaryResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if result.Success {
		fmt.Printf("Dictionary contains %d words:\n", result.Size)
		for i, word := range result.Words {
			fmt.Printf("%d. '%s'\n", i+1, word)
		}
	} else {
		fmt.Printf("Failed to get dictionary: %s\n", result.Message)
	}

	return nil
}
