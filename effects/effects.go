package effects

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/darkliquid/glitch/utils"
)

// WrapSlice wraps a slice of the image horizontally either left or right
func WrapSlice(destImage *image.RGBA, sourceImage *image.RGBA, xShift int, yPos int, height int, mask image.Image, op draw.Op) {
	if xShift == 0 {
		return
	}

	width := sourceImage.Bounds().Max.X

	// Wrap slice left
	if xShift < 0 {
		r := image.Rect(-xShift, yPos, width, yPos+height)
		p := image.Pt(0, yPos)
		draw.DrawMask(destImage, r, sourceImage, p, mask, p, op)

		r = image.Rect(0, yPos, -xShift, yPos+height)
		p = image.Pt(width+xShift, yPos)
		draw.DrawMask(destImage, r, sourceImage, p, mask, p, op)
		// Wrap slice right
	} else {
		r := image.Rect(0, yPos, width, yPos+height)
		p := image.Pt(xShift, yPos)
		draw.DrawMask(destImage, r, sourceImage, p, mask, p, op)

		r = image.Rect(width-xShift, yPos, width, yPos+height)
		p = image.Pt(0, yPos)
		draw.DrawMask(destImage, r, sourceImage, p, mask, p, op)
	}
}

// ApplyScanlines applies scanlines
func ApplyScanlines(destImage *image.RGBA) {
	bounds := destImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y = y + 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			destImage.Set(x, y, color.Black)
		}
	}
}

// ApplyBrightness increases brightness of image by brightness factor
func ApplyBrightness(destImage *image.RGBA, brightnessFactor float64) {
	bounds := destImage.Bounds()
	brightnessMultiplier := 1 + (brightnessFactor / 100)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Note type assertion to get a color.RGBA
			sourcePixel := destImage.At(x, y).(color.RGBA)
			destPixel := destImage.At(x, y).(color.RGBA)

			destPixel.R = uint8(math.Min(float64(sourcePixel.R)*brightnessMultiplier, 255))
			destPixel.G = uint8(math.Min(float64(sourcePixel.G)*brightnessMultiplier, 255))
			destPixel.B = uint8(math.Min(float64(sourcePixel.B)*brightnessMultiplier, 255))

			destImage.Set(x, y, destPixel)
		}
	}
}

// CopyChannel copies the channel data for one channel of an image onto the same channel of another image
func CopyChannel(destImage *image.RGBA, sourceImage *image.RGBA, copyChannel utils.Channel) {
	bounds := sourceImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Note type assertion to get a color.RGBA
			sourcePixel := sourceImage.At(x, y).(color.RGBA)
			destPixel := destImage.At(x, y).(color.RGBA)

			switch copyChannel {
			case utils.Red:
				destPixel.R = sourcePixel.R
			case utils.Green:
				destPixel.G = sourcePixel.G
			case utils.Blue:
				destPixel.B = sourcePixel.B
			case utils.Alpha:
				destPixel.A = sourcePixel.A
			}

			destImage.Set(x, y, destPixel)
		}
	}
}
