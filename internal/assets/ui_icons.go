package assets

import (
	"bytes"
	"image/png"

	_ "embed"

	"github.com/lxn/walk"
)

//go:embed close.png
var closePNG []byte

//go:embed new.png
var newPNG []byte

//go:embed up.png
var upPNG []byte

//go:embed refresh.png
var refreshPNG []byte

//go:embed disk.png
var diskPNG []byte

//go:embed like.png
var likePNG []byte

func buttonBitmap(data []byte, dpi int) (*walk.Bitmap, error) {
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return walk.NewBitmapFromImageForDPI(im, dpi)
}

func CloseBitmap(dpi int) (*walk.Bitmap, error)   { return buttonBitmap(closePNG, dpi) }
func NewTabBitmap(dpi int) (*walk.Bitmap, error)  { return buttonBitmap(newPNG, dpi) }
func UpBitmap(dpi int) (*walk.Bitmap, error)      { return buttonBitmap(upPNG, dpi) }
func RefreshBitmap(dpi int) (*walk.Bitmap, error) { return buttonBitmap(refreshPNG, dpi) }
func DiskBitmap(dpi int) (*walk.Bitmap, error)    { return buttonBitmap(diskPNG, dpi) }
func LikeBitmap(dpi int) (*walk.Bitmap, error)    { return buttonBitmap(likePNG, dpi) }
