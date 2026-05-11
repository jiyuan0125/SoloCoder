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

	"ruleengine/api"
)

var serverURL = "http://localhost:8212"

func getServerURL() string {
	if u := os.Getenv("RULE_ENGINE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return serverURL
}

func httpGet(url string, out interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func httpPost(url string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(respData, out)
	}
	return nil
}

func httpPut(url string, body interface{}, out interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(respData, out)
	}
	return nil
}

func httpDelete(url string, out interface{}) error {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if out != nil {
		return json.Unmarshal(respData, out)
	}
	return nil
}

func readJSONFile(path string, v interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func cmdAdd(args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	filePath := fs.String("f", "", "JSON file path containing the rule")
	fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "usage: client add -f <rule.json>")
		os.Exit(1)
	}

	var rule api.Rule
	if err := readJSONFile(*filePath, &rule); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read rule file: %v\n", err)
		os.Exit(1)
	}

	url := getServerURL() + "/rules"
	var resp api.AddRuleResponse
	if err := httpPost(url, api.AddRuleRequest{Rule: &rule}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "add rule failed: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printJSON(resp)
}

func cmdList(_ []string) {
	url := getServerURL() + "/rules"
	var resp api.ListRulesResponse
	if err := httpGet(url, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "list rules failed: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp)
}

func cmdEval(args []string) {
	fs := flag.NewFlagSet("eval", flag.ExitOnError)
	filePath := fs.String("f", "", "JSON file path containing environment variables")
	fs.Parse(args)

	if *filePath == "" {
		fmt.Fprintln(os.Stderr, "usage: client eval -f <env.json>")
		os.Exit(1)
	}

	var env map[string]interface{}
	if err := readJSONFile(*filePath, &env); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read env file: %v\n", err)
		os.Exit(1)
	}

	url := getServerURL() + "/evaluate"
	var resp api.EvaluateResponse
	if err := httpPost(url, api.EvaluateRequest{Environment: env}, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "evaluate failed: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printJSON(resp)
}

func cmdDelete(args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	ruleID := fs.String("id", "", "rule id to delete")
	fs.Parse(args)

	if *ruleID == "" {
		fmt.Fprintln(os.Stderr, "usage: client delete -id <rule_id>")
		os.Exit(1)
	}

	url := getServerURL() + "/rules/" + *ruleID
	var resp api.DeleteRuleResponse
	if err := httpDelete(url, &resp); err != nil {
		fmt.Fprintf(os.Stderr, "delete rule failed: %v\n", err)
		os.Exit(1)
	}

	if resp.Error != "" {
		fmt.Fprintf(os.Stderr, "server error: %s\n", resp.Error)
		os.Exit(1)
	}

	printJSON(resp)
}

func usage() {
	fmt.Println("usage: client <command> [options]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  add    - add a rule from JSON file")
	fmt.Println("  list   - list all rules")
	fmt.Println("  eval   - evaluate with environment variables")
	fmt.Println("  delete - delete a rule by id")
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  RULE_ENGINE_URL: override server URL (default http://localhost:8212)")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "add":
		cmdAdd(args)
	case "list":
		cmdList(args)
	case "eval":
		cmdEval(args)
	case "delete":
		cmdDelete(args)
	default:
		usage()
		os.Exit(1)
	}
}
