package main

import (
	"fmt"
	"os"
)

func main() {
	const (
		imageHeight = 256
		imageWidth  = 256
	)

	fmt.Printf("P3\n%d %d\n255\n", imageWidth, imageHeight)
	for j := 0; j < imageHeight; j++ {
		fmt.Fprintf(os.Stderr, "\rScanlines remaining: %d ", imageHeight-j)
		for i := 0; i < imageWidth; i++ {
			// r and g are represented as a value between 0 and 1
			r := float64(i) / (imageWidth - 1)
			g := float64(j) / (imageWidth - 1)
			scaledR := int(255.999 * r)
			scaledG := int(255.999 * g)
			b := 0

			fmt.Printf("%d %d %d\n", scaledR, scaledG, b)
		}
	}
}
