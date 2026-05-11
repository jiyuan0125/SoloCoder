package wiregen

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Generate(pkgPath string) (string, error) {
	pkg, err := ScanPackage(pkgPath)
	if err != nil {
		return "", fmt.Errorf("failed to scan package: %w", err)
	}

	gen := NewGenerator(pkg)
	code, err := gen.Generate()
	if err != nil {
		return "", err
	}

	return code, nil
}

func GenerateAndWrite(pkgPath string) error {
	code, err := Generate(pkgPath)
	if err != nil {
		return err
	}

	outputPath := filepath.Join(pkgPath, "wire_gen.go")
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("failed to write generated code: %w", err)
	}

	return nil
}

func IsCircularDependencyError(err error) bool {
	var target *CircularDependencyError
	return errors.As(err, &target)
}

func IsMultipleProvidersError(err error) bool {
	var target *MultipleProvidersError
	return errors.As(err, &target)
}

func IsMissingProviderError(err error) bool {
	var target *MissingProviderError
	return errors.As(err, &target)
}

func GetCircularDependencyPath(err error) []string {
	var target *CircularDependencyError
	if errors.As(err, &target) {
		return target.Path
	}
	return nil
}
