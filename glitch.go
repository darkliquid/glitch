package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
)

// Types and constants for Channel enums
type Channel int

const (
	RED Channel = iota
	GREEN
	BLUE
	ALPHA
)

// Global Vars
var seed string
var glitchFactor float64
var brightnessFactor float64
var useScanLines bool
var inputImage string
var outputImage string

// Custom usage info func for flags package
func usage() {
	fmt.Fprintf(os.Stderr, "Usage: glitch [-gbls] input_image output_image\n")
	flag.PrintDefaults()
	os.Exit(2)
}

// Just die with an error message
func bail(message string) {
	fmt.Fprintf(os.Stderr, message+"\n")
	os.Exit(1)
}

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

// Generates a random int64 seed value from the seed string
func randomseed() (seedInt int64) {
	hasher := md5.New()
	io.WriteString(hasher, seed)
	hash := hasher.Sum(nil)

	length := len(hash)
	for i, hashByte := range hash {
		// Get byte shift offset as a uint64
		shift := uint64((length - i - length) * 8)
		// OR the shifted byte onto the return value
		seedInt |= int64(hashByte) << shift
	}

	return
}

func random_channel() Channel {
	r := rand.Float32()
	if r < 0.33 {
		return GREEN
	} else if r < 0.66 {
		return RED
	}
	return BLUE
}

func copy_channel(destImage *image.RGBA, sourceImage *image.RGBA, copyChannel Channel) {
	bounds := sourceImage.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			// Note type assertion to get a color.RGBA
			source_pixel := sourceImage.At(x, y).(color.RGBA)
			dest_pixel := destImage.At(x, y).(color.RGBA)
			
			switch copyChannel {
			case RED:
				dest_pixel.R = source_pixel.R
			case GREEN:
				dest_pixel.G = source_pixel.G
			case BLUE:
				dest_pixel.B = source_pixel.B
			case ALPHA:
				dest_pixel.A = source_pixel.A
			}

			destImage.Set(x, y, dest_pixel)
		}
	}
}

func wrap_slice(destImage *image.RGBA, sourceImage *image.RGBA, xShift int, yPos int, height int) {
	if xShift == 0 {
		return
	}

	width := sourceImage.Bounds().Max.X

	// Wrap slice left
	if xShift < 0 {
		r := image.Rect(-xShift, yPos, width, yPos + height)
		p := image.Pt(0, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)

		r = image.Rect(0, yPos, -xShift, yPos + height)
		p = image.Pt(width+xShift, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)
	// Wrap slice right
	} else {
		r := image.Rect(0, yPos, width, yPos + height)
		p := image.Pt(xShift, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)

		r = image.Rect(width-xShift, yPos, width, yPos + height)
		p = image.Pt(0, yPos)
		draw.Draw(destImage, r, sourceImage, p, draw.Src)
	}
}

// Actually does useful stuff
func glitchify() {
	reader, err := os.Open(inputImage)
	if err != nil {
		bail("Couldn't open input image!")
	}

	// Decode the image data from the input file. Don't care about format registration
	inputDecode, _, err := image.Decode(reader)
	if err != nil {
		bail("Couldn't decode image data!")
	}

	// Close reader since we've got the image data now
	reader.Close()

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

		wrap_slice(outputData, inputData, offset, startY, chunkHeight)
	}

	// Copy a random channel from the pristene original input data onto the slice-offsetted output data
	copy_channel(outputData, inputData, random_channel())

	// Prep writing the output file
	writer, err := os.Create(outputImage)
	if err != nil {
		bail("Couldn't create output file!")
	}
	defer writer.Close()

	// Pass off image writing to appropriate encoder
	switch filepath.Ext(outputImage) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(writer, outputData, &jpeg.Options{jpeg.DefaultQuality})
	case ".gif":
		err = gif.Encode(writer, outputData, &gif.Options{256, nil, nil})
	case ".png":
		err = png.Encode(writer, outputData)
	default:
		bail("Image format not supported. Please use GIF, JPEG or PNG.")
	}

	if err != nil {
		bail("There was an error encoding the image data.")
	}
}

// Main
func main() {
	// Setup usage info
	flag.Usage = usage

	// Get the host name for the default seed
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Get Glitch Factor
	flag.Float64Var(&glitchFactor, "glitch", 5.0, "Defines how much glitching to do (0-100)")
	flag.Float64Var(&glitchFactor, "g", 5.0, "Defines how much glitching to do (0-100) - shorthand syntax")

	// Get Brightness Factor
	flag.Float64Var(&brightnessFactor, "brightness", 5.0, "Defines how much brightening to do (0-100)")
	flag.Float64Var(&brightnessFactor, "b", 5.0, "Defines how much brightening to do (0-100) - shorthand syntax")

	// Should do scan line effect?
	flag.BoolVar(&useScanLines, "scanlines", true, "Apply the scan line filter")
	flag.BoolVar(&useScanLines, "l", true, "Apply the scan line filter - shorthand syntax")

	// A seed to use for the randomiser
	flag.StringVar(&seed, "seed", hostname, "Seed for the randomiser")
	flag.StringVar(&seed, "s", hostname, "Seed for the randomiser - shorthand syntax")

	flag.Parse()

	inputImage = flag.Arg(0)
	outputImage = flag.Arg(1)

	// Sanitise input
	switch {
	case len(inputImage) == 0:
		fmt.Fprintf(os.Stderr, "No input image specified\n")
		usage()
	case len(outputImage) == 0:
		fmt.Fprintf(os.Stderr, "No output image specified\n")
		usage()
	case glitchFactor > 100.0 || glitchFactor < 0.0:
		fmt.Fprintf(os.Stderr, "Glitch factor must be between 0 and 100\n")
		usage()
	case brightnessFactor > 100.0 || brightnessFactor < 0.0:
		fmt.Fprintf(os.Stderr, "Brightness factor must be between 0 and 100\n")
		usage()
	}

	// Seed the random number generator
	rand.Seed(randomseed())

	// Onto the main event!
	glitchify()
}
