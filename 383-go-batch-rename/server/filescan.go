package main

import (
    "bufio"
    "os"
    "path/filepath"
    "strings"

    "batch-rename/protocol"
)

func collectFiles(req *protocol.Request) ([]string, error) {
    var files []string

    if len(req.Files) > 0 {
        files = make([]string, 0, len(req.Files))
        for _, f := range req.Files {
            info, err := os.Stat(f)
            if err != nil {
                continue
            }
            if info.IsDir() {
                dirFiles, err := collectFromDir(f, req.Recursive)
                if err != nil {
                    return nil, err
                }
                files = append(files, dirFiles...)
            } else {
                files = append(files, f)
            }
        }
        return files, nil
    }

    if req.DirPath != "" {
        return collectFromDir(req.DirPath, req.Recursive)
    }

    stdinFiles, err := collectFromStdin()
    if err != nil {
        return nil, err
    }
    if len(stdinFiles) > 0 {
        return stdinFiles, nil
    }

    cwd, err := os.Getwd()
    if err != nil {
        return nil, err
    }
    return collectFromDir(cwd, req.Recursive)
}

func collectFromDir(dir string, recursive bool) ([]string, error) {
    var files []string

    if recursive {
        err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if !info.IsDir() {
                files = append(files, path)
            }
            return nil
        })
        if err != nil {
            return nil, err
        }
    } else {
        entries, err := os.ReadDir(dir)
        if err != nil {
            return nil, err
        }
        for _, entry := range entries {
            if !entry.IsDir() {
                files = append(files, filepath.Join(dir, entry.Name()))
            }
        }
    }

    return files, nil
}

func collectFromStdin() ([]string, error) {
    var files []string
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        info, err := os.Stat(line)
        if err != nil {
            continue
        }
        if info.IsDir() {
            continue
        }
        files = append(files, line)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return files, nil
}
