package glitch

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"math/rand"
)

// Channel type for colour channels
type Channel int

const (
	// Red is the red channel
	Red Channel = iota
	// Green is the green channel
	Green
	// Blue is the blue channel
	Blue
	// Alpha is the alpha channel
	Alpha
)

// Spits out a random int between min and max
func random(min, max int) int {
	offset := 0
	input := max - min

	// Intn hates 0 or less, so we use this workaround
	if input <= 0 {
		offset = 1 + input*-1
		input = offset
	}

	return rand.Intn(input) + min - offset
}

// Pick a random colour channel (excludes ALPHA, since that's usually boring)
func randomChannel() Channel {
	r := rand.Float32()
	if r < 0.33 {
		return Green
	} else if r < 0.66 {
		return Red
	}
	return Blue
}

// Copy the channel data for one channel of an image onto the same channel of another image
func copyChannel(destImage *image.RGBA, sourceImage *image.RGBA, copyChannel Channel) {
	bounds := sourceImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Note type assertion to get a color.RGBA
			sourcePixel := sourceImage.At(x, y).(color.RGBA)
			destPixel := destImage.At(x, y).(color.RGBA)

			switch copyChannel {
			case Red:
				destPixel.R = sourcePixel.R
			case Green:
				destPixel.G = sourcePixel.G
			case Blue:
				destPixel.B = sourcePixel.B
			case Alpha:
				destPixel.A = sourcePixel.A
			}

			destImage.Set(x, y, destPixel)
		}
	}
}

// Increase brightness of image by brightness factor
func applyBrightness(destImage *image.RGBA, brightnessFactor float64) {
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

// Applies scanlines
func applyScanlines(destImage *image.RGBA) {
	bounds := destImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y = y + 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			destImage.Set(x, y, color.Black)
		}
	}
}

// Wrap a slice of the image horizontally either left or right
func wrapSlice(destImage *image.RGBA, sourceImage *image.RGBA, xShift int, yPos int, height int) {
	if xShift == 0 {
		return
	}

	width := sourceImage.Bounds().Max.X

	// Wrap slice left
	if xShift < 0 {
		r := image.Rect(-xShift, yPos, width, yPos+height)
		p := image.Pt(0, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)

		r = image.Rect(0, yPos, -xShift, yPos+height)
		p = image.Pt(width+xShift, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)
		// Wrap slice right
	} else {
		r := image.Rect(0, yPos, width, yPos+height)
		p := image.Pt(xShift, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)

		r = image.Rect(width-xShift, yPos, width, yPos+height)
		p = image.Pt(0, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)
	}
}

// Glitchify returns the glitchified input image
func Glitchify(inputDecode image.Image, glitchFactor, brightnessFactor float64, useScanLines bool) image.Image {
	// Useful values
	bounds := inputDecode.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	maxOffset := int(glitchFactor / 100.0 * float64(width))

	// Initialise input as RGBA data
	inputData := image.NewRGBA(bounds)
	draw.Draw(inputData, bounds, inputDecode, bounds.Min, draw.Src)

	// Initialise output as identical to input
	outputData := image.NewRGBA(bounds)
	draw.Draw(outputData, bounds, inputDecode, bounds.Min, draw.Src)

	// Random image slice offsetting
	for i := 0.0; i < glitchFactor*2; i++ {
		startY := random(0, height)
		chunkHeight := int(math.Min(float64(height-startY), float64(random(1, height/4))))
		offset := random(-maxOffset, maxOffset)

		wrapSlice(outputData, inputData, offset, startY, chunkHeight)
	}

	// Copy a random channel from the pristene original input data onto the slice-offsetted output data
	copyChannel(outputData, inputData, randomChannel())

	// Do brightness filter
	applyBrightness(outputData, brightnessFactor)

	// Apply scanlines
	if useScanLines {
		applyScanlines(outputData)
	}

	return outputData
}
