package assets

import (
	"bytes"
	"image/png"

	"github.com/lxn/walk"
)

// WalkLogoIcon returns the application logo as a walk icon for window title bars.
func WalkLogoIcon(dpi int) (*walk.Icon, error) {
	im, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		return nil, err
	}
	return walk.NewIconFromImageForDPI(im, dpi)
}
