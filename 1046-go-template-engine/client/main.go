package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"template-engine/common"
	"template-engine/template"
)

const defaultServer = "http://localhost:8080"

type config struct {
	server      string
	local       bool
	template    string
	templateFile string
	data        string
	dataFile    string
	name        string
	action      string
}

func main() {
	cfg := parseFlags()

	switch cfg.action {
	case "render":
		if cfg.local {
			localRender(cfg)
		} else {
			remoteRender(cfg)
		}
	case "register":
		registerTemplate(cfg)
	case "render-name":
		renderByName(cfg)
	case "list":
		listTemplates(cfg)
	case "delete":
		deleteTemplate(cfg)
	default:
		fmt.Println("unknown action:", cfg.action)
		os.Exit(1)
	}
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.server, "server", defaultServer, "template server URL")
	flag.BoolVar(&cfg.local, "local", false, "render locally without server")
	flag.StringVar(&cfg.template, "template", "", "template string")
	flag.StringVar(&cfg.templateFile, "template-file", "", "path to template file")
	flag.StringVar(&cfg.data, "data", "", "JSON data string")
	flag.StringVar(&cfg.dataFile, "data-file", "", "path to data file (JSON or YAML)")
	flag.StringVar(&cfg.name, "name", "", "template name for register/render-by-name")

	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		cfg.action = args[0]
	} else {
		cfg.action = "render"
	}

	return cfg
}

func localRender(cfg *config) {
	tpl, err := getTemplate(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	data, err := getData(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	engine := template.New(nil)
	result, err := engine.Render(tpl, data)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func remoteRender(cfg *config) {
	tpl, err := getTemplate(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	data, err := getData(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	req := common.RenderRequest{
		Template: tpl,
		Data:     data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	resp, err := http.Post(cfg.server+"/render", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result common.RenderResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Fprintln(os.Stderr, "error:", result.Error)
		os.Exit(1)
	}

	fmt.Print(result.Result)
}

func registerTemplate(cfg *config) {
	if cfg.name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required for register")
		os.Exit(1)
	}

	tpl, err := getTemplate(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	req := common.TemplateRegisterRequest{
		Name:     cfg.name,
		Template: tpl,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	resp, err := http.Post(cfg.server+"/templates/register", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result common.TemplateRegisterResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Fprintln(os.Stderr, "error:", result.Error)
		os.Exit(1)
	}

	fmt.Println("template registered successfully")
}

func renderByName(cfg *config) {
	if cfg.name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required for render-name")
		os.Exit(1)
	}

	data, err := getData(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	req := common.TemplateRenderByNameRequest{
		Name: cfg.name,
		Data: data,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	resp, err := http.Post(cfg.server+"/templates/render", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result common.RenderResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Fprintln(os.Stderr, "error:", result.Error)
		os.Exit(1)
	}

	fmt.Print(result.Result)
}

func listTemplates(cfg *config) {
	resp, err := http.Get(cfg.server + "/templates/list")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result common.TemplateListResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Fprintln(os.Stderr, "error:", result.Error)
		os.Exit(1)
	}

	for _, t := range result.Templates {
		fmt.Println(t.Name)
	}
}

func deleteTemplate(cfg *config) {
	if cfg.name == "" {
		fmt.Fprintln(os.Stderr, "error: -name is required for delete")
		os.Exit(1)
	}

	req, err := http.NewRequest("DELETE", cfg.server+"/templates/delete?name="+cfg.name, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	var result common.TemplateRegisterResponse
	json.Unmarshal(respBody, &result)

	if !result.Success {
		fmt.Fprintln(os.Stderr, "error:", result.Error)
		os.Exit(1)
	}

	fmt.Println("template deleted successfully")
}

func getTemplate(cfg *config) (string, error) {
	if cfg.template != "" {
		return cfg.template, nil
	}
	if cfg.templateFile != "" {
		return readFile(cfg.templateFile)
	}

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		body, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	return "", fmt.Errorf("template not provided: use -template, -template-file, or pipe from stdin")
}

func getData(cfg *config) (map[string]interface{}, error) {
	if cfg.data != "" {
		return parseDataString(cfg.data)
	}
	if cfg.dataFile != "" {
		return readDataFile(cfg.dataFile)
	}
	return map[string]interface{}{}, nil
}

func parseDataString(data string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}
	if err := yaml.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}
	return nil, fmt.Errorf("failed to parse data as JSON or YAML")
}

func readDataFile(path string) (map[string]interface{}, error) {
	data, err := readFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var result map[string]interface{}

	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal([]byte(data), &result); err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		return result, nil
	}

	if err := json.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}
	if err := yaml.Unmarshal([]byte(data), &result); err == nil {
		return result, nil
	}

	return nil, fmt.Errorf("failed to parse data file as JSON or YAML")
}

func readFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	body, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
