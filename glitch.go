package glitch

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/darkliquid/glitch/dither"
	"github.com/darkliquid/glitch/effects"
	"github.com/darkliquid/glitch/utils"
)

// Debug enables debugging print outs
var Debug bool

// The imageglitcher algorithm from airtight interactive
func imageglitcher(inputData, outputData *image.RGBA, bounds image.Rectangle, glitchFactor float64) {
	width, height := bounds.Max.X, bounds.Max.Y
	maxOffset := int(glitchFactor / 100.0 * float64(width))
	mask := image.NewUniform(color.Alpha{A: 255})

	// Random image slice offsetting
	for i := 0.0; i < glitchFactor*2; i++ {
		startY := utils.Random(0, height)
		chunkHeight := int(math.Min(float64(height-startY), float64(utils.Random(1, height/4))))
		offset := utils.Random(-maxOffset, maxOffset)

		effects.WrapSlice(outputData, inputData, offset, startY, chunkHeight, mask, draw.Src)
	}

	// Copy a random channel from the pristene original input data onto the slice-offsetted output data
	effects.CopyChannel(outputData, inputData, utils.RandomChannel())
}

func wtfify(inputData, outputData *image.RGBA, bounds image.Rectangle, glitchFactor float64) {
	copyInput := image.NewRGBA(bounds)
	copy(copyInput.Pix, inputData.Pix)

	eightBitted := image.NewRGBA(bounds)
	copy(eightBitted.Pix, inputData.Pix)
	dither.EightBit(eightBitted, utils.Random(0, 255))

	atkinsons := image.NewRGBA(bounds)
	copy(atkinsons.Pix, inputData.Pix)
	dither.Atkinsons(atkinsons, uint8(utils.Random(0, 255)))

	bayer := image.NewRGBA(bounds)
	copy(bayer.Pix, inputData.Pix)
	dither.Bayer(bayer)

	halftone := image.NewRGBA(bounds)
	copy(halftone.Pix, inputData.Pix)
	dither.Halftone(halftone, uint16(utils.Random(0, 255)))

	floydsteinberg := image.NewRGBA(bounds)
	copy(floydsteinberg.Pix, inputData.Pix)
	dither.FloydSteinberg(floydsteinberg, uint8(utils.Random(0, 255)))

	redOnly := image.NewRGBA(bounds)
	effects.CopyChannel(redOnly, inputData, utils.Red)

	greenOnly := image.NewRGBA(bounds)
	effects.CopyChannel(greenOnly, inputData, utils.Green)

	blueOnly := image.NewRGBA(bounds)
	effects.CopyChannel(blueOnly, inputData, utils.Blue)

	alphaMask := image.NewAlpha(bounds)
	for i := range alphaMask.Pix {
		alphaMask.Pix[i] = inputData.Pix[i*4]
	}

	srcs := []*image.RGBA{
		eightBitted,
		halftone,
		redOnly,
		greenOnly,
		blueOnly,
		copyInput,
	}
	srcNames := []string{
		"8bit",
		"halftone",
		"red",
		"green",
		"blue",
		"original",
	}

	wrapSlice := func(in, out *image.RGBA, op draw.Op) {
		width, height := bounds.Max.X, bounds.Max.Y
		maxOffset := int(glitchFactor / 100.0 * float64(width))

		// Random image slice offsetting
		for i := 0.0; i < glitchFactor; i++ {
			startY := utils.Random(0, height)
			chunkHeight := int(math.Min(float64(height-startY), float64(utils.Random(1, int(float64(height/2)*glitchFactor/100.0)))))
			offset := utils.Random(-maxOffset, maxOffset)
			effects.WrapSlice(out, in, offset, startY, chunkHeight, alphaMask, op)
		}
	}

	transforms := []func(in, out *image.RGBA){
		func(in, out *image.RGBA) {
			newIn := image.NewRGBA(bounds)
			copy(newIn.Pix, in.Pix)
			dither.Atkinsons(newIn, uint8(utils.Random(64, 192)))
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = newIn.Pix[i*4]
			}
			wrapSlice(newIn, out, draw.Over)
		},
		func(in, out *image.RGBA) {
			newIn := image.NewRGBA(bounds)
			copy(newIn.Pix, in.Pix)
			dither.EightBit(newIn, utils.Random(64, 192))
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = newIn.Pix[i*4]
			}
			wrapSlice(newIn, out, draw.Over)
		},
		func(in, out *image.RGBA) {
			newIn := image.NewRGBA(bounds)
			copy(newIn.Pix, in.Pix)
			dither.Bayer(newIn)
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = newIn.Pix[i*4]
			}
			wrapSlice(newIn, out, draw.Over)
		},
		func(in, out *image.RGBA) {
			newIn := image.NewRGBA(bounds)
			copy(newIn.Pix, in.Pix)
			dither.Halftone(newIn, uint16(utils.Random(64, 192)))
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = newIn.Pix[i*4]
			}
			wrapSlice(newIn, out, draw.Over)
		},
		func(in, out *image.RGBA) {
			newIn := image.NewRGBA(bounds)
			copy(newIn.Pix, in.Pix)
			dither.FloydSteinberg(newIn, uint8(utils.Random(64, 192)))
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = newIn.Pix[i*4]
			}
			wrapSlice(newIn, out, draw.Over)
		},
		func(in, out *image.RGBA) { wrapSlice(in, out, draw.Over) },
		func(in, out *image.RGBA) { wrapSlice(in, out, draw.Src) },
		func(in, out *image.RGBA) { effects.CopyChannel(out, in, utils.Red) },
		func(in, out *image.RGBA) { effects.CopyChannel(out, in, utils.Green) },
		func(in, out *image.RGBA) { effects.CopyChannel(out, in, utils.Blue) },
		func(in, out *image.RGBA) {
			for i := range alphaMask.Pix {
				alphaMask.Pix[i] = in.Pix[i*4]
			}
		},
	}
	transformNames := []string{
		"atkinsons",
		"8bit",
		"bayer",
		"halftone",
		"floydsteinberg",
		"wrapOver",
		"wrapSrc",
		"copyRed",
		"copyGreen",
		"copyBlue",
		"copyAlpha",
	}

	i := len(transforms)
	for i > 0 {
		destIdx := utils.Random(0, len(srcs))
		srcIdx := utils.Random(0, len(srcs))
		fIdx := utils.Random(0, len(transforms))
		transforms[fIdx](srcs[srcIdx], srcs[destIdx])
		if Debug {
			fmt.Printf("transform[%v] %v -> %v\n", transformNames[fIdx], srcNames[srcIdx], srcNames[destIdx])
		}
		destIdx = utils.Random(0, len(srcs))
		fIdx = utils.Random(0, len(transforms))
		transforms[fIdx](inputData, srcs[destIdx])

		i--
	}

	for i, src := range srcs {
		if Debug {
			fmt.Printf("transform[wrapOver] %v -> output\n", srcNames[i])
		}
		wrapSlice(src, outputData, draw.Over)
	}

	if Debug {
		fmt.Println("reset alpha mask")
	}
	for i := range alphaMask.Pix {
		alphaMask.Pix[i] = 255
	}

	finalOutput := image.NewRGBA(bounds)
	copy(finalOutput.Pix, outputData.Pix)
	if Debug {
		fmt.Println("imageglitcher for final output")
	}
	imageglitcher(finalOutput, outputData, bounds, glitchFactor)
}

// Glitchify returns the glitchified input image
func Glitchify(inputDecode image.Image, glitchFactor, brightnessFactor float64, useScanLines bool) image.Image {
	// Useful values
	bounds := inputDecode.Bounds()

	// Initialise input as RGBA data
	inputData := image.NewRGBA(bounds)
	draw.Draw(inputData, bounds, inputDecode, bounds.Min, draw.Src)

	// Initialise output as identical to input
	outputData := image.NewRGBA(bounds)
	draw.Draw(outputData, bounds, inputDecode, bounds.Min, draw.Src)

	//imageglitcher(inputData, outputData, bounds, glitchFactor)
	wtfify(inputData, outputData, bounds, glitchFactor)

	// Do brightness filter
	effects.ApplyBrightness(outputData, brightnessFactor)

	// Apply scanlines
	if useScanLines {
		effects.ApplyScanlines(outputData)
	}

	return outputData
}
