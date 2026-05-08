package main

import "net/http"

var httpClient = &http.Client{}

type stringArray []string

func (s *stringArray) String() string {
	return ""
}

func (s *stringArray) Set(value string) error {
	*s = append(*s, value)
	return nil
}
