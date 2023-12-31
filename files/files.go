package files

import (
	"embed"
)

var (
	//go:embed sounds
	Sounds embed.FS

	//go:embed games
	Games embed.FS

	//go:embed emblem.svg
	EmblemSVG []byte

	//go:embed emblem.png
	EmblemPNG []byte

	//go:embed demo/fpsdoug.jpg
	DemoAvatar []byte

	//go:embed default/avatar.jpg
	DefaultAvatar []byte
)
