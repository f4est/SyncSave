//go:build windows

package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"unsafe"
)

var procCreateIconFromResourceEx = user32.NewProc("CreateIconFromResourceEx")

// buildAppIcon creates a crisp 64x64 rounded-square blue-emerald SyncSave icon
// and returns a Win32 HICON handle using PNG resource loading (Windows Vista+).
func buildAppIcon() uintptr {
	const size = 64
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	// Rounded rectangle mask + diagonal gradient (#3b82f6 -> #10b981)
	const radius = 14.0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			alpha := roundedRectAlpha(float64(x)+0.5, float64(y)+0.5, size, size, radius)
			if alpha <= 0 {
				continue
			}
			t := (float64(x) + float64(y)) / float64(2*size)
			r := uint8(59*(1-t) + 16*t)
			g := uint8(130*(1-t) + 185*t)
			b := uint8(246*(1-t) + 129*t)
			a := uint8(alpha * 255)
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: a})
		}
	}

	// Draw crisp white sync/upload glyph (tray base + up/down arrows)
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	// Bottom tray U-shape
	fillRect(img, 16, 42, 48, 47, white)
	fillRect(img, 16, 33, 21, 47, white)
	fillRect(img, 43, 33, 48, 47, white)
	// Center vertical stem
	fillRect(img, 29, 17, 35, 39, white)
	// Arrowhead pointing up
	for row := 0; row < 12; row++ {
		y := 13 + row
		half := row + 2
		if half > 13 {
			half = 13
		}
		fillRect(img, 32-half, y, 32+half, y+1, white)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return 0
	}
	data := buf.Bytes()
	hIcon, _, _ := procCreateIconFromResourceEx.Call(
		uintptr(unsafe.Pointer(&data[0])),
		uintptr(len(data)),
		1,          // TRUE = icon
		0x00030000, // version
		0, 0,
		0x00000040, // LR_DEFAULTSIZE
	)
	return hIcon
}

func fillRect(img *image.NRGBA, x0, y0, x1, y1 int, c color.NRGBA) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.SetNRGBA(x, y, c)
			}
		}
	}
}

func roundedRectAlpha(x, y, w, h, r float64) float64 {
	cx := math.Max(r, math.Min(x, w-r))
	cy := math.Max(r, math.Min(y, h-r))
	dx := x - cx
	dy := y - cy
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist <= r-0.5 {
		return 1
	}
	if dist >= r+0.5 {
		return 0
	}
	return r + 0.5 - dist
}
