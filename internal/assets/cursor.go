package assets

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"sync"

	_ "embed"

	"fyne.io/fyne/v2/driver/desktop"
)

//go:embed copy.cur
var copyCUR []byte

// CopyCURBytes returns the embedded drag cursor file bytes.
func CopyCURBytes() []byte {
	out := make([]byte, len(copyCUR))
	copy(out, copyCUR)
	return out
}

type fileDragCursor struct {
	once sync.Once
	img  image.Image
	hotX int
	hotY int
}

var dragCursor fileDragCursor

// FileDragCursor returns the custom file-drag cursor for Fyne desktop builds.
func FileDragCursor() desktop.Cursor {
	dragCursor.once.Do(func() {
		img, x, y, err := decodeCursorImage(copyCUR)
		if err != nil {
			return
		}
		dragCursor.img = img
		dragCursor.hotX = x
		dragCursor.hotY = y
	})
	return &dragCursor
}

func (c *fileDragCursor) Image() (image.Image, int, int) {
	return c.img, c.hotX, c.hotY
}

func decodeCursorImage(data []byte) (image.Image, int, int, error) {
	if len(data) < 22 {
		return nil, 0, 0, fmt.Errorf("cursor file too small")
	}
	typ := binary.LittleEndian.Uint16(data[2:4])
	if typ != 1 && typ != 2 {
		return nil, 0, 0, fmt.Errorf("unsupported cursor icon type %d", typ)
	}
	entry := data[6:22]
	width := int(entry[0])
	height := int(entry[1])
	if width == 0 {
		width = 256
	}
	if height == 0 {
		height = 256
	}
	hotX, hotY := 0, 0
	if typ == 2 {
		hotX = int(binary.LittleEndian.Uint16(entry[4:6]))
		hotY = int(binary.LittleEndian.Uint16(entry[6:8]))
	}
	offset := int(binary.LittleEndian.Uint32(entry[12:16]))
	img, err := decodeICODIB(data, offset, width, height)
	if err != nil {
		return nil, 0, 0, err
	}
	return img, hotX, hotY, nil
}

func decodeICODIB(data []byte, offset, width, height int) (image.Image, error) {
	if offset+40 > len(data) {
		return nil, fmt.Errorf("cursor bitmap header out of range")
	}
	headerSize := int(binary.LittleEndian.Uint32(data[offset : offset+4]))
	if headerSize < 40 || offset+headerSize > len(data) {
		return nil, fmt.Errorf("invalid cursor bitmap header")
	}
	bmpH := int(binary.LittleEndian.Uint32(data[offset+8 : offset+12]))
	bpp := int(binary.LittleEndian.Uint16(data[offset+14 : offset+16]))
	iconH := bmpH / 2
	if iconH <= 0 {
		return nil, fmt.Errorf("invalid cursor height")
	}
	pixelOffset := offset + headerSize
	rowBytes := ((width*bpp + 31) / 32) * 4
	img := image.NewNRGBA(image.Rect(0, 0, width, iconH))
	for y := 0; y < iconH; y++ {
		srcY := iconH - 1 - y
		rowStart := pixelOffset + srcY*rowBytes
		if rowStart+rowBytes > len(data) {
			return nil, fmt.Errorf("cursor pixel data out of range")
		}
		switch bpp {
		case 32:
			for x := 0; x < width; x++ {
				i := rowStart + x*4
				b := data[i]
				g := data[i+1]
				r := data[i+2]
				a := data[i+3]
				if a == 0 && r == 0 && g == 0 && b == 0 {
					a = 0
				}
				img.Set(x, y, color.NRGBA{R: r, G: g, B: b, A: a})
			}
		default:
			return nil, fmt.Errorf("unsupported cursor bpp %d", bpp)
		}
	}
	return img, nil
}

// DecodeCursorImageForTest exposes cursor decoding for tests.
func DecodeCursorImageForTest(data []byte) (image.Image, int, int, error) {
	return decodeCursorImage(data)
}

var _ desktop.Cursor = (*fileDragCursor)(nil)
