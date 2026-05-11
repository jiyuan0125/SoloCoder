package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"go-cross-build/pkg/api"
)

type BuildOptions struct {
	WorkingDir string
}

type Builder struct {
	options *BuildOptions
}

func NewBuilder(options *BuildOptions) *Builder {
	if options == nil {
		options = &BuildOptions{}
	}
	if options.WorkingDir == "" {
		options.WorkingDir = "."
	}
	return &Builder{options: options}
}

func GetOutputName(target api.BuildTarget, baseName string) string {
	name := baseName
	if target.OutputName != "" {
		name = target.OutputName
	}

	if target.Platform.GOOS == "windows" {
		if !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
	}

	return name
}

func HasDebugStripped(ldflags string) bool {
	hasS := strings.Contains(ldflags, "-s")
	hasW := strings.Contains(ldflags, "-w")
	return hasS || hasW
}

func GenerateBuildCommand(target api.BuildTarget, config api.BuildConfig) *api.CommandResult {
	outputDir := config.OutputDir
	if outputDir == "" {
		outputDir = "dist"
	}

	baseName := filepath.Base(config.MainPackage)
	if baseName == "." {
		baseName = filepath.Base(outputDir)
		if baseName == "." || baseName == "" {
			baseName = "app"
		}
	}

	platformSuffix := fmt.Sprintf("%s_%s", target.Platform.GOOS, target.Platform.GOARCH)
	outputName := GetOutputName(target, baseName)
	outputFile := filepath.Join(outputDir, platformSuffix, outputName)

	envVars := []string{
		fmt.Sprintf("GOOS=%s", target.Platform.GOOS),
		fmt.Sprintf("GOARCH=%s", target.Platform.GOARCH),
	}

	if target.Env != nil {
		for k, v := range target.Env {
			envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
		}
	}

	args := []string{"build", "-o", outputFile}

	if target.Ldflags != "" {
		args = append(args, "-ldflags", target.Ldflags)
	}

	args = append(args, config.MainPackage)

	command := "go " + strings.Join(args, " ")

	return &api.CommandResult{
		Target:     target,
		Command:    command,
		EnvVars:    envVars,
		OutputFile: outputFile,
	}
}

func (b *Builder) Build(target api.BuildTarget, config api.BuildConfig) (*api.BuildResult, error) {
	result := &api.BuildResult{
		Target:  target,
		Success: false,
	}

	cmdResult := GenerateBuildCommand(target, config)

	result.DebugStripped = HasDebugStripped(target.Ldflags)

	hasCGO, cgoFiles, err := DetectCGODependency(b.options.WorkingDir)
	if err != nil {
		return nil, err
	}
	result.HasCGODependency = hasCGO

	if hasCGO {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("CGO dependency detected in files: %v", cgoFiles))

		cgoEnabled := false
		for k, v := range target.Env {
			if k == "CGO_ENABLED" && v == "1" {
				cgoEnabled = true
			}
		}

		if !cgoEnabled {
			result.Warnings = append(result.Warnings,
				"Cross-compilation with CGO may fail. Consider setting CGO_ENABLED=1 or using a cross-compiler.")
		}
	}

	files, err := ScanSourceFiles(b.options.WorkingDir)
	if err != nil {
		return nil, err
	}

	warnings := ValidatePlatformFiles(files, target.Platform.GOOS, target.Platform.GOARCH)
	result.Warnings = append(result.Warnings, warnings...)

	if err := os.MkdirAll(filepath.Dir(cmdResult.OutputFile), 0755); err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to create output directory: %v", err)
		return result, nil
	}

	cmd := exec.Command("go", strings.Split(cmdResult.Command, " ")[1:]...)
	cmd.Dir = b.options.WorkingDir
	cmd.Env = append(os.Environ(), cmdResult.EnvVars...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("build failed: %v\n%s", err, string(output))
		return result, nil
	}

	if info, err := os.Stat(cmdResult.OutputFile); err == nil {
		result.FileSize = info.Size()
		result.OutputPath = cmdResult.OutputFile
	}

	if config.ReleaseMode {
		checksum, err := computeSHA256(cmdResult.OutputFile)
		if err != nil {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("failed to compute checksum: %v", err))
		} else {
			result.Checksum = checksum

			checksumFile := cmdResult.OutputFile + ".sha256"
			if err := os.WriteFile(checksumFile,
				[]byte(fmt.Sprintf("%s  %s\n", checksum, filepath.Base(cmdResult.OutputFile))),
				0644); err != nil {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("failed to write checksum file: %v", err))
			}
		}
	}

	result.Success = true
	return result, nil
}

func (b *Builder) BuildAll(config api.BuildConfig) (*api.BuildResponse, error) {
	response := &api.BuildResponse{
		TotalTargets: len(config.Targets),
	}

	for _, target := range config.Targets {
		result, err := b.Build(target, config)
		if err != nil {
			response.Results = append(response.Results, api.BuildResult{
				Target:       target,
				Success:      false,
				ErrorMessage: err.Error(),
			})
			response.FailureCount++
			continue
		}

		response.Results = append(response.Results, *result)
		if result.Success {
			response.SuccessCount++
		} else {
			response.FailureCount++
		}
	}

	return response, nil
}

func computeSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
