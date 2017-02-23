package dither

import (
	"image"
	"math"
)

// EightBit does an 8bit dither of the given image
func EightBit(destImage *image.RGBA, threshold int) {
	bounds := destImage.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	size := 4
	sizeSq := uint16(size * size)
	for y := 0; y < height; y += size {
		for x := 0; x < width; x += size {
			var sumR, sumG, sumB uint16
			for sY := 0; sY < size; sY++ {
				for sX := 0; sX < size; sX++ {
					i := 4 * (width*(y+sY) + (x + sX))
					if i >= len(destImage.Pix) {
						continue
					}
					sumR += uint16(destImage.Pix[i])
					sumG += uint16(destImage.Pix[i+1])
					sumB += uint16(destImage.Pix[i+2])
				}
			}

			var avgR, avgG, avgB uint8
			if sumR/sizeSq > uint16(threshold) {
				avgR = 0xff
			}
			if sumG/sizeSq > uint16(threshold) {
				avgG = 0xff
			}
			if sumB/sizeSq > uint16(threshold) {
				avgB = 0xff
			}

			for sY := 0; sY < size; sY++ {
				for sX := 0; sX < size; sX++ {
					i := 4 * (width*(y+sY) + (x + sX))
					if i >= len(destImage.Pix) {
						continue
					}
					destImage.Pix[i] = avgR
					destImage.Pix[i+1] = avgG
					destImage.Pix[i+2] = avgB
				}
			}
		}
	}
}

// Bayer does a Bayer dither of the given image
func Bayer(destImage *image.RGBA) {
	bounds := destImage.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	thresholdMap := [][]float64{
		{1, 9, 3, 11},
		{13, 5, 15, 7},
		{4, 12, 2, 10},
		{16, 8, 14, 6},
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := 4 * (y*width + x)
			gray := .3*float64(destImage.Pix[i]) + .59*float64(destImage.Pix[i+1]) + .11*float64(destImage.Pix[i+2])
			scaled := (gray * 17) / 255
			var val uint8
			if scaled > thresholdMap[x%4][y%4] {
				val = 0xff
			}
			destImage.Pix[i] = val
			destImage.Pix[i+1] = val
			destImage.Pix[i+2] = val
		}
	}
}

// Halftone does a halftone dither of the given image
func Halftone(destImage *image.RGBA, threshold uint16) {
	bounds := destImage.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	for y := 0; y <= height-2; y += 3 {
		for x := 0; x <= width-2; x += 3 {
			var sumR, sumG, sumB uint16
			var indexed []int
			count := 0
			for sY := 0; sY < 3; sY++ {
				for sX := 0; sX < 3; sX++ {
					i := 4 * (width*(y+sY) + (x + sX))
					sumR += uint16(destImage.Pix[i])
					sumG += uint16(destImage.Pix[i+1])
					sumB += uint16(destImage.Pix[i+2])
					destImage.Pix[i] = 0xff
					destImage.Pix[i+1] = 0xff
					destImage.Pix[i+2] = 0xff
					indexed = append(indexed, i)
					count++
				}
			}

			var avgR, avgG, avgB uint8
			if (sumR / 9) > threshold {
				avgR = 0xff
			}
			if (sumG / 9) > threshold {
				avgG = 0xff
			}
			if (sumB / 9) > threshold {
				avgB = 0xff
			}
			avgLum := float64(avgR+avgG+avgB) / 3
			scaled := math.Floor(((avgLum * 9) / 255) + .5)
			if scaled < 9 {
				destImage.Pix[indexed[4]] = avgR
				destImage.Pix[indexed[4]+1] = avgG
				destImage.Pix[indexed[4]+2] = avgB
			}
			if scaled < 8 {
				destImage.Pix[indexed[5]] = avgR
				destImage.Pix[indexed[5]+1] = avgG
				destImage.Pix[indexed[5]+2] = avgB
			}
			if scaled < 7 {
				destImage.Pix[indexed[1]] = avgR
				destImage.Pix[indexed[1]+1] = avgG
				destImage.Pix[indexed[1]+2] = avgB
			}
			if scaled < 6 {
				destImage.Pix[indexed[6]] = avgR
				destImage.Pix[indexed[6]+1] = avgG
				destImage.Pix[indexed[6]+2] = avgB
			}
			if scaled < 5 {
				destImage.Pix[indexed[3]] = avgR
				destImage.Pix[indexed[3]+1] = avgG
				destImage.Pix[indexed[3]+2] = avgB
			}
			if scaled < 4 {
				destImage.Pix[indexed[8]] = avgR
				destImage.Pix[indexed[8]+1] = avgG
				destImage.Pix[indexed[8]+2] = avgB
			}
			if scaled < 3 {
				destImage.Pix[indexed[2]] = avgR
				destImage.Pix[indexed[2]+1] = avgG
				destImage.Pix[indexed[2]+2] = avgB
			}
			if scaled < 2 {
				destImage.Pix[indexed[0]] = avgR
				destImage.Pix[indexed[0]+1] = avgG
				destImage.Pix[indexed[0]+2] = avgB
			}

			if scaled < 1 {
				destImage.Pix[indexed[7]] = avgR
				destImage.Pix[indexed[7]+1] = avgG
				destImage.Pix[indexed[7]+2] = avgB
			}
		}
	}
}

func adjustPixelError(data []uint8, i int, r, g, b uint8, multiplier float64) {
	if i >= len(data) {
		return
	}
	data[i] = data[i] + uint8(multiplier*float64(r))
	data[i+1] = data[i+1] + uint8(multiplier*float64(g))
	data[i+2] = data[i+2] + uint8(multiplier*float64(b))
}

// Atkinsons does an Atkinsons dither of the given image
func Atkinsons(destImage *image.RGBA, threshold uint8) {
	bounds := destImage.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := 4 * (y*width + x)

			oldR := destImage.Pix[i]
			oldG := destImage.Pix[i+1]
			oldB := destImage.Pix[i+2]

			var newR, newG, newB uint8
			if oldR > threshold {
				newR = 0xff
			}
			if oldG > threshold {
				newG = 0xff
			}
			if oldB > threshold {
				newB = 0xff
			}

			destImage.Pix[i] = newR
			destImage.Pix[i+1] = newG
			destImage.Pix[i+2] = newB

			errR := oldR - newR
			errG := oldG - newG
			errB := oldB - newB

			// Redistribute the pixel's error like this:
			//       *  1/8 1/8
			//  1/8 1/8 1/8
			//      1/8
			// The ones to the right...
			if x < width-1 {
				adjI := i + 4
				adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				// The pixel that's down and to the right
				if y < height-1 {
					adjI = adjI + (width * 4) + 4
					adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				}
				// The pixel two over
				if x < width-2 {
					adjI = i + 8
					adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				}
			}
			if y < height-1 {
				// The one right below
				adjI := i + (width * 4)
				adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				if x > 0 {
					// The one to the left
					adjI = adjI - 4
					adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				}
				if y < height-2 {
					// The one two down
					adjI = i + (2 * width * 4)
					adjustPixelError(destImage.Pix, adjI, errR, errG, errB, 1.0/8.0)
				}
			}
		}
	}
}

// FloydSteinberg does a Floyd-Steinberg dither of the given image
func FloydSteinberg(destImage *image.RGBA, threshold uint8) {
	bounds := destImage.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := 4 * (y*width + x)

			oldR := destImage.Pix[i]
			oldG := destImage.Pix[i+1]
			oldB := destImage.Pix[i+2]

			var newR, newG, newB uint8
			if oldR > threshold {
				newR = 0xff
			}
			if oldG > threshold {
				newG = 0xff
			}
			if oldB > threshold {
				newB = 0xff
			}

			destImage.Pix[i] = newR
			destImage.Pix[i+1] = newG
			destImage.Pix[i+2] = newB

			errR := oldR - newR
			errG := oldG - newG
			errB := oldB - newB

			// Redistribute the pixel's error like this:
			//   * 7
			// 3 5 1
			// The ones to the right...
			if x < width-1 {
				rightI := i + 4
				adjustPixelError(destImage.Pix, rightI, errR, errG, errB, 7.0/16.0)
				// The pixel that's down and to the right
				if y < height-1 {
					nextRightI := rightI + (width * 4)
					adjustPixelError(destImage.Pix, nextRightI, errR, errG, errB, 1.0/16.0)
				}
			}

			if y < height-1 {
				// The one right below
				downI := i + (width * 4)
				adjustPixelError(destImage.Pix, downI, errR, errG, errB, 5.0/16.0)
				if x > 0 {
					// The one down and to the left...
					leftI := downI - 4
					adjustPixelError(destImage.Pix, leftI, errR, errG, errB, 3.0/16.0)
				}
			}
		}
	}
}
