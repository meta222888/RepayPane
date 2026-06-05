package assets

import (
	"testing"
)

func TestDecodeEmbeddedDragCursor(t *testing.T) {
	img, hotX, hotY := 0, 0, 0
	got, x, y, err := DecodeCursorImageForTest(copyCUR)
	if err != nil {
		t.Fatalf("decode cursor: %v", err)
	}
	if got.Bounds().Dx() == 0 || got.Bounds().Dy() == 0 {
		t.Fatalf("empty cursor image")
	}
	img, hotX, hotY = got.Bounds().Dx(), x, y
	if img != 32 {
		t.Fatalf("expected 32px cursor, got %d (hotspot %d,%d)", img, hotX, hotY)
	}
	_ = hotY
}
