package utils

import "math/rand"

// Random spits out a random int between min and max
func Random(min, max int) int {
	offset := 0
	input := max - min

	// Intn hates 0 or less, so we use this workaround
	if input <= 0 {
		offset = 1 + input*-1
		input = offset
	}

	return rand.Intn(input) + min - offset
}

// RandomChannel picks a random colour channel (excludes ALPHA, since that's usually boring)
func RandomChannel() Channel {
	r := rand.Float32()
	if r < 0.33 {
		return Green
	} else if r < 0.66 {
		return Red
	}
	return Blue
}
