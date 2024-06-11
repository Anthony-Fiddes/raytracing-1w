package main

import (
	"fmt"
	"os"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func (v Vec3) Add(other Vec3) Vec3 {
	v.X += other.X
	v.Y += other.Y
	v.Z += other.Z
	return v
}

func (v Vec3) Subtract(other Vec3) Vec3 {
	v.X -= other.X
	v.Y -= other.Y
	v.Z -= other.Z
	return v
}

func (v Vec3) Scale(factor float64) Vec3 {
	v.X *= factor
	v.Y *= factor
	v.Z *= factor
	return v
}

type Color struct{ Vec Vec3 }

func (c Color) R() float64 {
	return c.Vec.X
}

func (c Color) G() float64 {
	return c.Vec.Y
}

func (c Color) B() float64 {
	return c.Vec.Z
}

func toPPM(c Color) string {
	scaledR := int(255.999 * c.R())
	scaledG := int(255.999 * c.G())
	scaledB := int(255.999 * c.B())
	return fmt.Sprintf("%d %d %d\n", scaledR, scaledG, scaledB)
}

type Ray struct {
	Origin    Vec3
	Direction Vec3
}

func (r Ray) At(t float64) Vec3 {
	distance := r.Direction.Scale(t)
	result := r.Origin.Add(distance)
	return result
}

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
	fmt.Fprint(os.Stderr, "\rDone.                    \n")
}
