package main

import (
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"

    "batch-rename/protocol"
)

func generatePreview(files []string, rules []protocol.Rule) ([]protocol.RenameItem, error) {
    items := make([]protocol.RenameItem, 0, len(files))
    newPathSet := make(map[string]struct{})

    for i, filePath := range files {
        item, err := processFile(filePath, i+1, len(files), rules)
        if err != nil {
            return nil, err
        }

        if item.OldPath == item.NewPath {
            item.Skipped = true
            item.SkipReason = "Old name equals new name"
        } else if _, exists := newPathSet[item.NewPath]; exists {
            item.Skipped = true
            item.SkipReason = "Duplicate new name"
        } else if fileExists(item.NewPath) {
            item.Skipped = true
            item.SkipReason = "Target file already exists"
        }

        items = append(items, item)
        if !item.Skipped {
            newPathSet[item.NewPath] = struct{}{}
        }
    }

    return items, nil
}

func processFile(filePath string, seqNum int, totalFiles int, rules []protocol.Rule) (protocol.RenameItem, error) {
    dir := filepath.Dir(filePath)
    fileName := filepath.Base(filePath)
    ext := filepath.Ext(fileName)
    baseName := strings.TrimSuffix(fileName, ext)

    currentName := baseName

    for _, rule := range rules {
        newName, err := applyRule(currentName, rule, seqNum, totalFiles, filePath)
        if err != nil {
            return protocol.RenameItem{}, err
        }
        currentName = newName
    }

    currentName = sanitizeFileName(currentName)
    newFileName := currentName + ext
    newPath := filepath.Join(dir, newFileName)

    return protocol.RenameItem{
        OldPath: filePath,
        NewPath: newPath,
        OldName: fileName,
        NewName: newFileName,
    }, nil
}

func applyRule(name string, rule protocol.Rule, seqNum int, totalFiles int, filePath string) (string, error) {
    switch rule.Type {
    case protocol.RuleTypePrefix:
        return rule.Prefix + name, nil

    case protocol.RuleTypeSuffix:
        return name + rule.Suffix, nil

    case protocol.RuleTypeSequence:
        return replaceSequence(name, seqNum, totalFiles), nil

    case protocol.RuleTypeDate:
        dateFormat := "20060102"
        if rule.DateFormat != "" {
            dateFormat = rule.DateFormat
        }
        return replaceDate(name, filePath, dateFormat), nil

    case protocol.RuleTypeReplace:
        return strings.ReplaceAll(name, rule.OldString, rule.NewString), nil

    case protocol.RuleTypeRegexReplace:
        return regexReplace(name, rule.Pattern, rule.Replacement)

    default:
        return name, nil
    }
}

func replaceSequence(name string, seqNum int, totalFiles int) string {
    padding := 0
    for totalFiles > 0 {
        padding++
        totalFiles /= 10
    }

    seqStr := fmt.Sprintf("%0*d", padding, seqNum)
    return strings.ReplaceAll(name, "{seq}", seqStr)
}

func replaceDate(name string, filePath string, dateFormat string) string {
    info, err := os.Stat(filePath)
    if err != nil {
        return name
    }

    modTime := info.ModTime()
    dateStr := modTime.Format(dateFormat)
    return strings.ReplaceAll(name, "{date}", dateStr)
}

func regexReplace(name string, pattern string, replacement string) (string, error) {
    re, err := regexp.Compile(pattern)
    if err != nil {
        return "", err
    }
    return re.ReplaceAllString(name, replacement), nil
}

func sanitizeFileName(name string) string {
    invalidChars := []rune{'/', '\\', ':', '*', '?', '"', '<', '>', '|'}
    result := []rune(name)
    for i, r := range result {
        for _, invalid := range invalidChars {
            if r == invalid {
                result[i] = '_'
                break
            }
        }
    }
    return string(result)
}

func fileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil
}

func executeRenames(items []protocol.RenameItem) ([]protocol.RenameItem, int, int, int) {
    resultItems := make([]protocol.RenameItem, 0, len(items))
    success := 0
    failed := 0
    skipped := 0

    for _, item := range items {
        resultItem := item

        if item.Skipped {
            skipped++
            resultItems = append(resultItems, resultItem)
            continue
        }

        err := os.Rename(item.OldPath, item.NewPath)
        if err != nil {
            failed++
            resultItem.Skipped = true
            resultItem.SkipReason = fmt.Sprintf("Rename failed: %v", err)
        } else {
            success++
        }

        resultItems = append(resultItems, resultItem)
    }

    return resultItems, success, failed, skipped
}
