package assets

import "embed"

//go:embed sfx/*.wav
var SFXFS embed.FS

//go:embed icons/icon_final.png
var WindowIconPNG []byte
