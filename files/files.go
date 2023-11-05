package files

import (
	"embed"
)

var (
	//go:embed sounds
	Sounds embed.FS

	//go:embed gameicons
	GameIcons embed.FS

	//go:embed emblem.svg
	EmblemSVG []byte

	//go:embed emblem.png
	EmblemPNG []byte
)
