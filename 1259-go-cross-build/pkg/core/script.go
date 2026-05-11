package core

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-cross-build/pkg/api"
)

func GenerateShellScript(config api.BuildConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\n\n")
	sb.WriteString("# Auto-generated cross-platform build script\n")
	sb.WriteString(fmt.Sprintf("# Project: %s\n", config.Name))
	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("# Description: %s\n", config.Description))
	}
	sb.WriteString("\n")

	sb.WriteString("set -e\n\n")

	sb.WriteString(fmt.Sprintf("OUTPUT_DIR=\"%s\"\n", config.OutputDir))
	sb.WriteString(fmt.Sprintf("MAIN_PKG=\"%s\"\n", config.MainPackage))
	sb.WriteString("\n")

	sb.WriteString("mkdir -p \"$OUTPUT_DIR\"\n\n")

	for i, target := range config.Targets {
		cmdResult := GenerateBuildCommand(target, config)
		platformSuffix := fmt.Sprintf("%s_%s", target.Platform.GOOS, target.Platform.GOARCH)

		sb.WriteString(fmt.Sprintf("# Build %s\n", platformSuffix))
		sb.WriteString(fmt.Sprintf("echo \"Building for %s/%s...\"\n",
			target.Platform.GOOS, target.Platform.GOARCH))
		sb.WriteString(fmt.Sprintf("mkdir -p \"$OUTPUT_DIR/%s\"\n", platformSuffix))

		if len(cmdResult.EnvVars) > 0 {
			sb.WriteString(strings.Join(cmdResult.EnvVars, " "))
			sb.WriteString(" ")
		}

		args := strings.Split(cmdResult.Command, " ")
		for j, arg := range args {
			if strings.Contains(arg, " ") {
				args[j] = fmt.Sprintf("\"%s\"", arg)
			}
		}
		sb.WriteString(strings.Join(args, " "))
		sb.WriteString("\n")

		if config.ReleaseMode {
			sb.WriteString(fmt.Sprintf(
				"sha256sum \"$OUTPUT_DIR/%s/%s\" > \"$OUTPUT_DIR/%s/%s.sha256\"\n",
				platformSuffix, filepath.Base(cmdResult.OutputFile),
				platformSuffix, filepath.Base(cmdResult.OutputFile)))
		}

		if i < len(config.Targets)-1 {
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\necho \"Build completed!\"\n")
	sb.WriteString(fmt.Sprintf("echo \"Output: $OUTPUT_DIR\"\n"))

	return sb.String(), nil
}

func GenerateMakefile(config api.BuildConfig) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Auto-generated cross-platform build Makefile\n")
	sb.WriteString(fmt.Sprintf("# Project: %s\n", config.Name))
	if config.Description != "" {
		sb.WriteString(fmt.Sprintf("# Description: %s\n", config.Description))
	}
	sb.WriteString("\n")

	sb.WriteString("OUTPUT_DIR ?= dist\n")
	sb.WriteString(fmt.Sprintf("MAIN_PKG ?= %s\n", config.MainPackage))
	sb.WriteString("\n")

	var allTargets []string

	for _, target := range config.Targets {
		platformSuffix := fmt.Sprintf("%s_%s", target.Platform.GOOS, target.Platform.GOARCH)
		allTargets = append(allTargets, platformSuffix)

		cmdResult := GenerateBuildCommand(target, config)
		outputFile := cmdResult.OutputFile

		sb.WriteString(fmt.Sprintf(".PHONY: %s\n", platformSuffix))
		sb.WriteString(fmt.Sprintf("%s: $(OUTPUT_DIR)/%s/%s\n\n",
			platformSuffix, platformSuffix, filepath.Base(outputFile)))

		sb.WriteString(fmt.Sprintf("$(OUTPUT_DIR)/%s/%s: $(MAIN_PKG)\n",
			platformSuffix, filepath.Base(outputFile)))
		sb.WriteString("\t@mkdir -p $(dir $@)\n")

		envVars := ""
		if len(cmdResult.EnvVars) > 0 {
			envVars = strings.Join(cmdResult.EnvVars, " ") + " "
		}

		args := strings.Split(cmdResult.Command, " ")
		for j, arg := range args {
			if strings.Contains(arg, " ") {
				args[j] = fmt.Sprintf("\"%s\"", arg)
			}
		}
		sb.WriteString(fmt.Sprintf("\t@echo \"Building for %s/%s...\"\n",
			target.Platform.GOOS, target.Platform.GOARCH))
		sb.WriteString(fmt.Sprintf("\t@%s%s\n", envVars, strings.Join(args, " ")))

		if config.ReleaseMode {
			sb.WriteString(fmt.Sprintf(
				"\t@sha256sum \"$@\" > \"$@.sha256\"\n"))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(".PHONY: all clean\n")
	sb.WriteString(fmt.Sprintf("all: %s\n", strings.Join(allTargets, " ")))
	sb.WriteString("\t@echo \"All builds completed!\"\n\n")

	sb.WriteString("clean:\n")
	sb.WriteString("\t@rm -rf $(OUTPUT_DIR)\n")

	return sb.String(), nil
}

func GenerateScript(config api.BuildConfig, format string) (string, error) {
	switch strings.ToLower(format) {
	case "shell", "sh", "bash":
		return GenerateShellScript(config)
	case "makefile", "make":
		return GenerateMakefile(config)
	default:
		return "", fmt.Errorf("unsupported script format: %s", format)
	}
}
