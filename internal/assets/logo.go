package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed logo.png
var logoPNG []byte

func LogoResource() fyne.Resource {
	return fyne.NewStaticResource("logo.png", logoPNG)
}
