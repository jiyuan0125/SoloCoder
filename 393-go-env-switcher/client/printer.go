package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"env-switcher/protocol"
)

type ResponsePrinter struct{}

func NewResponsePrinter() *ResponsePrinter {
	return &ResponsePrinter{}
}

func (p *ResponsePrinter) Print(command string, resp *protocol.Response) {
	if !resp.Success {
		if resp.Message != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		}
		os.Exit(1)
	}
	
	switch command {
	case "list":
		p.printList(resp)
	case "show":
		p.printShow(resp)
	case "export":
		p.printExport(resp)
	case "switch":
		p.printSwitch(resp)
	case "add":
		p.printAdd(resp)
	case "del":
		p.printDel(resp)
	}
}

func (p *ResponsePrinter) printList(resp *protocol.Response) {
	if len(resp.Envs) == 0 {
		fmt.Println("No environments available")
		return
	}
	
	maxLen := 0
	for _, name := range resp.Envs {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}
	
	fmt.Printf("%-*s  Status\n", maxLen+2, "Environment")
	fmt.Println(strings.Repeat("-", maxLen+2+10))
	
	for _, name := range resp.Envs {
		status := ""
		if name == resp.Current {
			status = "* (active)"
		}
		fmt.Printf("%-*s  %s\n", maxLen+2, name, status)
	}
}

func (p *ResponsePrinter) printShow(resp *protocol.Response) {
	if len(resp.Vars) == 0 {
		fmt.Printf("Environment '%s' has no variables\n", resp.Current)
		return
	}
	
	var keys []string
	for key := range resp.Vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	maxKeyLen := 0
	for _, key := range keys {
		if len(key) > maxKeyLen {
			maxKeyLen = len(key)
		}
	}
	
	fmt.Printf("Environment: %s\n\n", resp.Current)
	for _, key := range keys {
		fmt.Printf("%-*s = %s\n", maxKeyLen, key, resp.Vars[key])
	}
}

func (p *ResponsePrinter) printExport(resp *protocol.Response) {
	if len(resp.Vars) == 0 {
		return
	}
	
	var keys []string
	for key := range resp.Vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	for _, key := range keys {
		value := resp.Vars[key]
		escapedValue := p.escapeShellValue(value)
		fmt.Printf("export %s=%s\n", key, escapedValue)
	}
}

func (p *ResponsePrinter) escapeShellValue(value string) string {
	if len(value) == 0 {
		return "''"
	}
	
	needsQuoting := false
	specialChars := []rune{' ', '\t', '\n', '"', '\'', '\\', '!', '$', '&', '(', ')', '*', ';', '<', '>', '?', '[', ']', '^', '`', '{', '|', '}', '~'}
	
	for _, char := range value {
		for _, special := range specialChars {
			if char == special {
				needsQuoting = true
				break
			}
		}
		if needsQuoting {
			break
		}
	}
	
	if !needsQuoting {
		return value
	}
	
	escaped := strings.ReplaceAll(value, "'", "'\\''")
	return "'" + escaped + "'"
}

func (p *ResponsePrinter) printSwitch(resp *protocol.Response) {
	if resp.Message != "" {
		fmt.Println(resp.Message)
	}
	if len(resp.Warnings) > 0 {
		for _, warning := range resp.Warnings {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
	}
}

func (p *ResponsePrinter) printAdd(resp *protocol.Response) {
	if resp.Message != "" {
		fmt.Println(resp.Message)
	}
}

func (p *ResponsePrinter) printDel(resp *protocol.Response) {
	if resp.Message != "" {
		fmt.Println(resp.Message)
	}
}
