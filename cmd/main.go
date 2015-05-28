package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"path/filepath"

	"github.com/darkliquid/glitch"
)

// Custom usage info func for flags package
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: glitch [-gbls] input_image output_image")
	flag.PrintDefaults()
	os.Exit(2)
}

// Just die with an error message
func bail(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

// Generates a random int64 seed value from the seed string
func randomseed(seed string) (seedInt int64) {
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

// Main
func main() {
	var seed string
	var glitchFactor float64
	var brightnessFactor float64
	var useScanLines bool
	var inputImage string
	var outputImage string

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
		fmt.Fprintln(os.Stderr, "No input image specified")
		usage()
	case len(outputImage) == 0:
		fmt.Fprintln(os.Stderr, "No output image specified")
		usage()
	case glitchFactor > 100.0 || glitchFactor < 0.0:
		fmt.Fprintln(os.Stderr, "Glitch factor must be between 0 and 100")
		usage()
	case brightnessFactor > 100.0 || brightnessFactor < 0.0:
		fmt.Fprintln(os.Stderr, "Brightness factor must be between 0 and 100")
		usage()
	}

	// Seed the random number generator
	rand.Seed(randomseed(seed))

	// Prep writing the output file
	writer, err := os.Create(outputImage)
	if err != nil {
		bail("Couldn't create output file!")
	}
	defer writer.Close()

	// Onto the main event!
	reader, err := os.Open(inputImage)
	if err != nil {
		bail("Couldn't open input file!")
	}
	defer reader.Close()

	inputImg, _, err := image.Decode(reader)
	if err != nil {
		bail("Couldn't decode input file!")
	}

	outputImg := glitch.Glitchify(inputImg, glitchFactor, brightnessFactor, useScanLines)
	if err != nil {
		bail("Couldn't glitch image!")
	}

	// Pass off image writing to appropriate encoder
	switch filepath.Ext(outputImage) {
	case ".jpg", ".jpeg":
		err = jpeg.Encode(writer, outputImg, &jpeg.Options{Quality: jpeg.DefaultQuality})
	case ".gif":
		err = gif.Encode(writer, outputImg, &gif.Options{NumColors: 256, Drawer: nil, Quantizer: nil})
	case ".png":
		err = png.Encode(writer, outputImg)
	default:
		bail("Image format not supported. Please use GIF, JPEG or PNG.")
	}

	if err != nil {
		bail("Couldn't encode image")
	}
}
