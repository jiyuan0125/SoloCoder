package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"go-cross-build/pkg/api"
	"go-cross-build/pkg/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "init":
		cmdInit(args)
	case "validate":
		cmdValidate(args)
	case "script":
		cmdScript(args)
	case "build":
		cmdBuild(args)
	case "commands":
		cmdCommands(args)
	case "remote":
		cmdRemote(args)
	case "help":
		printUsage()
	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`go-cross-build - Go cross-platform build tool

Usage:
  go-cross-build <command> [options]

Commands:
  init                Create a default build configuration
  validate          Validate platform combinations
  script        Generate build script (shell or Makefile)
  build            Execute local build
  commands      Generate build commands without executing
  remote           Commands for remote server

Examples:
  go-cross-build init
  go-cross-build validate --config build.json
  go-cross-build script --config build.json --output build.sh
  go-cross-build script --config build.json --makefile --output Makefile
  go-cross-build build --config build.json
  go-cross-build commands --config build.json
  go-cross-build remote build --server http://localhost:8080 --config build.json
`)
}

func cmdInit(args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	output := fs.String("output", "build.json", "output file path")
	fs.Parse(args)

	config := core.CreateDefaultConfig()

	if err := core.SaveConfig(config, *output); err != nil {
		log.Fatalf("failed to create config: %v", err)
	}

	fmt.Printf("Created default configuration at %s\n", *output)
}

func loadConfig(path string) *api.BuildConfig {
	config, err := core.LoadConfig(path)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	return config
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	configPath := fs.String("config", "build.json", "configuration file path")
	fs.Parse(args)

	config := loadConfig(*configPath)

	fmt.Printf("Validating configuration: %s\n", config.Name)

	hasError := false
	for _, target := range config.Targets {
		valid := core.IsValidCombination(target.Platform.GOOS, target.Platform.GOARCH)
		status := "✓"
		if !valid {
			status = "✗"
			hasError = true
		}
		fmt.Printf("  %s %s/%s\n", status, target.Platform.GOOS, target.Platform.GOARCH)
	}

	if hasError {
		os.Exit(1)
	}

	fmt.Println("All platforms are valid!")
}

func cmdScript(args []string) {
	fs := flag.NewFlagSet("script", flag.ExitOnError)
	configPath := fs.String("config", "build.json", "configuration file path")
	output := fs.String("output", "", "output file path")
	useMakefile := fs.Bool("makefile", false, "generate Makefile instead of shell script")
	fs.Parse(args)

	config := loadConfig(*configPath)

	format := "shell"
	if *useMakefile {
		format = "makefile"
	}

	content, err := core.GenerateScript(*config, format)
	if err != nil {
		log.Fatalf("failed to generate script: %v", err)
	}

	if *output != "" {
		if err := ioutil.WriteFile(*output, []byte(content), 0755); err != nil {
			log.Fatalf("failed to write script: %v", err)
		}
		fmt.Printf("Generated %s script at %s\n", format, *output)
	} else {
		fmt.Println(content)
	}
}

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	configPath := fs.String("config", "build.json", "configuration file path")
	fs.Parse(args)

	config := loadConfig(*configPath)

	fmt.Printf("Building: %s\n", config.Name)
	fmt.Printf("Targets: %d\n", len(config.Targets))

	builder := core.NewBuilder(nil)
	resp, err := builder.BuildAll(*config)
	if err != nil {
		log.Fatalf("build failed: %v", err)
	}

	fmt.Printf("\nResults:\n")
	for _, result := range resp.Results {
		status := "✓"
		if !result.Success {
			status = "✗"
		}
		fmt.Printf("  %s %s/%s", status,
			result.Target.Platform.GOOS, result.Target.Platform.GOARCH)

		if result.Success {
			fmt.Printf(" - %s", result.OutputPath)
			if result.FileSize > 0 {
				fmt.Printf(" (%d bytes)", result.FileSize)
			}
			if result.DebugStripped {
				fmt.Printf(" [stripped]")
			}
		}

		fmt.Println()

		if result.Checksum != "" {
			fmt.Printf("      SHA256: %s\n", result.Checksum)
		}

		for _, warning := range result.Warnings {
			fmt.Printf("      Warning: %s\n", warning)
		}

		if result.ErrorMessage != "" {
			fmt.Printf("      Error: %s\n", result.ErrorMessage)
		}
	}

	fmt.Printf("\nSummary: %d successful, %d failed\n",
		resp.SuccessCount, resp.FailureCount)

	if resp.FailureCount > 0 {
		os.Exit(1)
	}
}

func cmdCommands(args []string) {
	fs := flag.NewFlagSet("commands", flag.ExitOnError)
	configPath := fs.String("config", "build.json", "configuration file path")
	fs.Parse(args)

	config := loadConfig(*configPath)

	for _, target := range config.Targets {
		cmd := core.GenerateBuildCommand(target, *config)
		fmt.Printf("# %s/%s\n", target.Platform.GOOS, target.Platform.GOARCH)
		if len(cmd.EnvVars) > 0 {
			fmt.Print(strings.Join(cmd.EnvVars, " "), " ")
		}
		fmt.Println(cmd.Command)
		fmt.Println()
	}
}

func cmdRemote(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: go-cross-build remote <subcommand> [options]")
		fmt.Println("Subcommands: build, script, validate")
		os.Exit(1)
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "build":
		cmdRemoteBuild(subArgs)
	case "script":
		cmdRemoteScript(subArgs)
	case "validate":
		cmdRemoteValidate(subArgs)
	default:
		fmt.Printf("unknown remote command: %s\n", subCmd)
		os.Exit(1)
	}
}

func cmdRemoteBuild(args []string) {
	fs := flag.NewFlagSet("remote build", flag.ExitOnError)
	serverURL := fs.String("server", "", "remote server URL")
	configPath := fs.String("config", "build.json", "configuration file path")
	fs.Parse(args)

	if *serverURL == "" {
		log.Fatal("server URL is required")
	}

	config := loadConfig(*configPath)

	client := api.NewClient(*serverURL)

	req := &api.BuildRequest{
		Config: *config,
	}

	resp, err := client.Build(req)
	if err != nil {
		log.Fatalf("remote build failed: %v", err)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}

func cmdRemoteScript(args []string) {
	fs := flag.NewFlagSet("remote script", flag.ExitOnError)
	serverURL := fs.String("server", "", "remote server URL")
	configPath := fs.String("config", "build.json", "configuration file path")
	format := fs.String("format", "shell", "script format (shell or makefile)")
	fs.Parse(args)

	if *serverURL == "" {
		log.Fatal("server URL is required")
	}

	config := loadConfig(*configPath)

	client := api.NewClient(*serverURL)

	req := &api.ScriptRequest{
		Config: *config,
		Format: *format,
	}

	resp, err := client.GenerateScript(req)
	if err != nil {
		log.Fatalf("remote script generation failed: %v", err)
	}

	fmt.Println(resp.Content)
}

func cmdRemoteValidate(args []string) {
	fs := flag.NewFlagSet("remote validate", flag.ExitOnError)
	serverURL := fs.String("server", "", "remote server URL")
	configPath := fs.String("config", "build.json", "configuration file path")
	fs.Parse(args)

	if *serverURL == "" {
		log.Fatal("server URL is required")
	}

	config := loadConfig(*configPath)

	client := api.NewClient(*serverURL)

	platforms := make([]api.Platform, 0, len(config.Targets))
	for _, t := range config.Targets {
		platforms = append(platforms, t.Platform)
	}

	req := &api.ValidatePlatformRequest{
		Platforms: platforms,
	}

	resp, err := client.ValidatePlatforms(req)
	if err != nil {
		log.Fatalf("remote validation failed: %v", err)
	}

	data, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(data))
}
