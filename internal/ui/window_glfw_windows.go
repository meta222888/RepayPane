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
