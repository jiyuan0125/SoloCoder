package main

import "embed"

//go:embed all:assets/*
var embedFS embed.FS
