package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
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

func bail(message string) {
	fmt.Fprintf(os.Stderr, message+"\n")
	os.Exit(1)
}

func random(min, max int) int {
	offset := 0
	input := max - min

	if input <= 0 {
		offset = 1 + input*-1
		input = offset
	}

	return rand.Intn(input) + min - offset
}

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

// Actually does useful stuff
func glitchify() {
	reader, err := os.Open(inputImage)
	if err != nil {
		bail("Couldn't open input image!")
	}

	// Decode the image data from the input file. Don't care about format registration
	inputData, _, err := image.Decode(reader)
	if err != nil {
		bail("Couldn't decode image data!")
	}

	// Close reader since we've got the image data now
	reader.Close()

	// Useful values
	bounds := inputData.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	maxOffset := int(glitchFactor / 100.0 * float64(width))

	// Initialise output as identical to input
	outputData := image.NewRGBA(bounds)
	draw.Draw(outputData, bounds, inputData, bounds.Min, draw.Src)

	// Random image slice offsetting
	for i := 0.0; i < glitchFactor*2; i++ {
		startY := random(0, height)
		chunkHeight := int(math.Min(float64(height-startY), float64(random(1, height/4))))
		offset := random(-maxOffset, maxOffset)

		if offset == 0 {
			continue
		}

		if offset < 0 {
			draw.Draw(outputData, image.Rect(-offset, startY, width+offset, chunkHeight), inputData, image.Pt(0, startY), draw.Src)
			draw.Draw(outputData, image.Rect(0, startY, -offset, chunkHeight), inputData, image.Pt(width+offset, startY), draw.Src)
		} else {
			draw.Draw(outputData, image.Rect(0, startY, width, chunkHeight), inputData, image.Pt(offset, startY), draw.Src)
			draw.Draw(outputData, image.Rect(width-offset, startY, offset, chunkHeight), inputData, image.Pt(0, startY), draw.Src)
		}
	}

	writer, err := os.Create(outputImage)
	if err != nil {
		bail("Couldn't create output file!")
	}
	defer writer.Close()

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
		os.Exit(2)
	case len(outputImage) == 0:
		fmt.Fprintf(os.Stderr, "No output image specified\n")
		usage()
		os.Exit(2)
	case glitchFactor > 100.0 || glitchFactor < 0.0:
		fmt.Fprintf(os.Stderr, "Glitch factor must be between 0 and 100\n")
		usage()
		os.Exit(2)
	case brightnessFactor > 100.0 || brightnessFactor < 0.0:
		fmt.Fprintf(os.Stderr, "Brightness factor must be between 0 and 100\n")
		usage()
		os.Exit(2)
	}

	// Seed the random number generator
	rand.Seed(randomseed())

	// Onto the main event!
	glitchify()
}
