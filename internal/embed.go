package internal

import (
	"embed"
)

var (
	//go:embed all:resources/*
	res embed.FS
)
