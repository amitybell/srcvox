package main

import (
	"embed"
)

//go:embed frontend/dist
var assetsFS embed.FS
