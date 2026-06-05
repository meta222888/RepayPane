//go:build windows

package ui

import (
	"reflect"
	"strings"
	"unsafe"

	"fyne.io/fyne/v2"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func fyneGLFWViewport(w fyne.Window) (*glfw.Window, bool) {
	val := reflectWindowStruct(w)
	if !val.IsValid() {
		return nil, false
	}
	if !strings.Contains(val.Type().PkgPath(), "driver/glfw") {
		return nil, false
	}
	field, ok := val.Type().FieldByName("viewport")
	if !ok {
		return nil, false
	}
	base := unsafe.Pointer(val.UnsafeAddr())
	vpVal := reflect.NewAt(field.Type, unsafe.Add(base, field.Offset)).Elem()
	if vpVal.IsNil() {
		return nil, false
	}
	vp, ok := vpVal.Interface().(*glfw.Window)
	return vp, ok
}

func reflectWindowStruct(w fyne.Window) reflect.Value {
	if w == nil {
		return reflect.Value{}
	}
	v := reflect.ValueOf(w)
	for v.Kind() == reflect.Interface {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}

func fyneWindowPos(w fyne.Window) (x, y int, ok bool) {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return 0, 0, false
	}
	x, y = vp.GetPos()
	return x, y, true
}

func fyneWindowSize(w fyne.Window) (width, height int, ok bool) {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return 0, 0, false
	}
	width, height = vp.GetSize()
	return width, height, true
}

func fyneMoveWindowBy(w fyne.Window, dx, dy int) bool {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return false
	}
	x, y := vp.GetPos()
	vp.SetPos(x+dx, y+dy)
	return true
}

func fyneSetWindowPos(w fyne.Window, x, y int) bool {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return false
	}
	vp.SetPos(x, y)
	return true
}

func fyneSetWindowSize(w fyne.Window, width, height int) bool {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return false
	}
	if width < int(minWindowWidth) {
		width = int(minWindowWidth)
	}
	if height < int(minWindowHeight) {
		height = int(minWindowHeight)
	}
	vp.SetSize(width, height)
	return true
}

func fyneSetWindowBounds(w fyne.Window, x, y, width, height int) bool {
	vp, ok := fyneGLFWViewport(w)
	if !ok {
		return false
	}
	if width < int(minWindowWidth) {
		width = int(minWindowWidth)
	}
	if height < int(minWindowHeight) {
		height = int(minWindowHeight)
	}
	vp.SetPos(x, y)
	vp.SetSize(width, height)
	return true
}
