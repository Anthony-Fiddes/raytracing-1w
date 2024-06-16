package main

import (
	"fmt"
	"log"
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

func (v Vec3) Divide(factor float64) Vec3 {
	return v.Scale(1. / factor)
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) UnitVector() Vec3 {
	return v.Divide(v.Length())
}

// X, Y, and Z represent red, green, and blue values. They are floats between 0 and 1
type Color struct{ Vec Vec3 }

var (
	white = Color{Vec3{1, 1, 1}}
	black = Color{Vec3{0, 0, 0}}
)

func isValidColor(f float64) bool {
	if f < 0 || f > 1 {
		return false
	}
	return true
}

func (c Color) assertValid() {
	if !isValidColor(c.R()) {
		log.Panicf("%+v has invalid red value %g. It must be between 0 and 1", c, c.R())
	}
	if !isValidColor(c.G()) {
		log.Panicf("%+v has invalid green value %g. It must be between 0 and 1", c, c.G())
	}
	if !isValidColor(c.B()) {
		log.Panicf("%+v has invalid blue value %g. It must be between 0 and 1", c, c.B())
	}
}

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
	c.assertValid()
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
			v := Vec3{
				float64(i) / (imageWidth - 1),
				float64(j) / (imageWidth - 1),
				0,
			}
			pixel := Color{v}
			fmt.Printf(toPPM(pixel))
		}
	}
	fmt.Fprint(os.Stderr, "\rDone.                    \n")
}
