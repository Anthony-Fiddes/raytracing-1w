package main

import (
	"fmt"
	"log"
	"math"
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

func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// X, Y, and Z represent red, green, and blue values. They are floats between 0 and 1
type Color struct{ Vec Vec3 }

var (
	white = Color{Vec3{1, 1, 1}}
	black = Color{Vec3{0, 0, 0}}
	red   = Color{Vec3{1, 0, 0}}
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

// camera is an object in the world
type camera struct {
	// Height is the number of pixels up/down
	Height int
	// Width is the number of pixels left/right
	Width       int
	aspectRatio float64
	Center      Vec3
	// distance between camera and viewport
	focalLength float64
	Viewport    viewport
}

// viewport represents the image that the camera captures.
//
// An increasing x goes to the right and increasing y goes down.
type viewport struct {
	width            float64
	height           float64
	widthVector      Vec3
	heightVector     Vec3
	PixelDeltaX      Vec3
	PixelDeltaY      Vec3
	upperLeft        Vec3
	FirstPixelCenter Vec3
}

func NewCamera(width int, aspectRatio float64) camera {
	if width <= 0 {
		panic("width cannot be <= 0")
	}
	if aspectRatio <= 0 {
		panic("aspect ratio cannot be <= 0")
	}
	height := int(float64(width) / aspectRatio)
	if height == 0 {
		height = 1
	}

	center := Vec3{X: 0, Y: 0, Z: 0}
	viewHeight := 2.0
	viewWidth := 2.0 * float64(width) / float64(height)
	widthVector := Vec3{viewWidth, 0, 0}
	heightVector := Vec3{0, -viewHeight, 0}
	pixelDeltaX := widthVector.Divide(float64(width))
	pixelDeltaY := heightVector.Divide(float64(height))
	focalLength := 1.0
	upperLeft := center.Subtract(Vec3{0, 0, focalLength}).Subtract(widthVector.Divide(2)).Subtract(heightVector.Divide(2))
	firstPixelCenter := upperLeft.Add(pixelDeltaX.Divide(2)).Add(pixelDeltaY.Divide(2))
	viewport := viewport{
		viewWidth, viewHeight, widthVector,
		heightVector, pixelDeltaX, pixelDeltaY,
		upperLeft, firstPixelCenter,
	}
	return camera{height, width, aspectRatio, center, focalLength, viewport}
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

func (r Ray) Color() Color {
	if r.HitSphere(Sphere{Vec3{0, 0, -1}, 0.5}) {
		return red
	}

	unitDirection := r.Direction.UnitVector()
	// unit vector's y ranges from [-1, 1], so we transform the range to [0, 1]
	// to do a linear interpolation and get a nice gradient from white to blue
	a := 0.5*unitDirection.Y + 1
	lightBlue := Color{Vec3{0.5, 0.7, 1}}
	colorVec := white.Vec.Scale(1 - a).Add(lightBlue.Vec.Scale(a))
	return Color{colorVec}
}

func (r Ray) HitSphere(sphere Sphere) bool {
	if sphere.Radius < 0 {
		log.Panicf("Sphere radius cannot be negative")
	}

	/*
		we can tell whether a ray hits the sphere by considering the following
		quadratic equation:

		(t^2)(d * d) - 2(d * Z)t + (Z * Z - r^2) = 0

		derived from (C - (Q + td)) * (C - (Q + td)) = r^2

		Explanation:

		* is the dot operator

		Z is (C-Q) where C is the center of the circle and Q is the origin of the
		ray

		d is the vector describing the direction of the ray

		r is the radius of the sphere

		t is the input of the quadratic. It is used to scale the direction
		vector of the ray to tell us how far along the ray we are. When t
		satisfies the above equation, the ray has hit the sphere.

		We can test how many roots there are to this equation by just calculating the
		discriminant. If it's less than 0, then there are no real solutions to the
		equation, which means that the ray does not hit the sphere. Otherwise there are
		one or two solutions, so the ray DOES hit.
	*/
	Z := sphere.Center.Subtract(r.Origin)
	a := r.Direction.Dot(r.Direction)
	b := r.Direction.Scale(-2).Dot(Z)
	c := Z.Dot(Z) - (sphere.Radius * sphere.Radius)
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false
	}
	return true
}

type Sphere struct {
	Center Vec3
	Radius float64
}

func main() {
	camera := NewCamera(400, 16./9.)
	viewport := camera.Viewport

	fmt.Printf("P3\n%d %d\n255\n", camera.Width, camera.Height)
	for j := 0; j < camera.Height; j++ {
		fmt.Fprintf(os.Stderr, "\rScanlines remaining: %d ", camera.Height-j)
		yPixelCenter := camera.Viewport.FirstPixelCenter.Add(viewport.PixelDeltaY.Scale(float64(j)))
		for i := 0; i < camera.Width; i++ {
			pixelCenter := yPixelCenter.Add(viewport.PixelDeltaX.Scale(float64(i)))
			rayDirection := pixelCenter.Subtract(camera.Center)
			ray := Ray{camera.Center, rayDirection}
			pixel := ray.Color()
			fmt.Printf(toPPM(pixel))
		}
	}
	fmt.Fprint(os.Stderr, "\rDone.                    \n")
}
