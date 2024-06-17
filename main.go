package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
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

func (c Color) String() string {
	return fmt.Sprintf("Color{Red: %v, Green: %v, Blue: %v}", c.R(), c.G(), c.B())
}

func (c Color) assertValid() {
	if !isValidColor(c.R()) {
		log.Panicf("%v has invalid red value %g. It must be between 0 and 1", c, c.R())
	}
	if !isValidColor(c.G()) {
		log.Panicf("%v has invalid green value %g. It must be between 0 and 1", c, c.G())
	}
	if !isValidColor(c.B()) {
		log.Panicf("%v has invalid blue value %g. It must be between 0 and 1", c, c.B())
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
	focalLength     float64
	Viewport        viewport
	SamplesPerPixel int
}

func (c camera) Render(w io.Writer, world Hittable) {
	fmt.Fprintf(w, "P3\n%d %d\n255\n", c.Width, c.Height)
	for j := 0; j < c.Height; j++ {
		fmt.Fprintf(os.Stderr, "\rScanlines remaining: %d ", c.Height-j)
		for i := 0; i < c.Width; i++ {
			var pixel Color
			for range c.SamplesPerPixel {
				sampleXOffset := rand.Float64() - 0.5
				sampleYOffset := rand.Float64() - 0.5
				yPixelCenter := c.Viewport.FirstPixelCenter.Add(c.Viewport.PixelDeltaY.Scale(float64(j) + sampleYOffset))
				sampleCenter := yPixelCenter.Add(c.Viewport.PixelDeltaX.Scale(float64(i) + sampleXOffset))
				rayDirection := sampleCenter.Subtract(c.Center)
				ray := Ray{c.Center, rayDirection}
				pixel.Vec = pixel.Vec.Add(ray.Color(world, 0, math.Inf(1)).Vec)
			}
			pixel.Vec = pixel.Vec.Divide(float64(c.SamplesPerPixel))
			fmt.Printf(toPPM(pixel))
		}
	}
	fmt.Fprint(os.Stderr, "\rDone.                    \n")
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

func NewCamera(width int, aspectRatio float64, samplesPerPixel int) camera {
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
	viewWidth := viewHeight * float64(width) / float64(height)
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
	return camera{height, width, aspectRatio, center, focalLength, viewport, samplesPerPixel}
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

func (r Ray) Color(h Hittable, tMin float64, tMax float64) Color {
	if hit, hr := h.Hit(r, tMin, tMax); hit {
		// all normals will be unit vectors
		// map the normal vector [-1,1] to valid color space [0,1]
		color := Color{hr.Normal.Add(Vec3{1, 1, 1}).Divide(2)}
		return color
	}

	unitDirection := r.Direction.UnitVector()
	// unit vector's y ranges from [-1, 1], so we transform the range to [0, 1]
	// to do a linear interpolation and get a nice gradient from white to blue
	a := 0.5*unitDirection.Y + 1
	lightBlue := Color{Vec3{0.5, 0.7, 1}}
	colorVec := white.Vec.Scale(1 - a).Add(lightBlue.Vec.Scale(a))
	return Color{colorVec}
}

type HitRecord struct {
	// Factor to scale ray by to get hit point
	T float64
	// Normal vector at the hit point. It points against the ray and is
	// expected to be a unit vector.
	Normal Vec3
	// Exterior is whether the ray hit the geometry from the outside or the
	// inside
	Exterior bool
}

// outwardNormal is a normal pointing out of the hit geometry.
func NewHitRecord(ray Ray, t float64, outwardNormal Vec3) HitRecord {
	// If the ray * outwardNormal was negative, that would mean that the angle
	// between the ray and outward normal is obtuse, meaning that the ray DOES point
	// against the exterior.
	exterior := ray.Direction.Dot(outwardNormal) < 0
	var normal Vec3
	if exterior {
		normal = outwardNormal
	} else {
		normal = outwardNormal.Scale(-1)
	}

	length := normal.Length()
	const acceptableDelta = 0.001
	if math.Abs(length-1) > acceptableDelta {
		log.Panicf(
			"Normal %+v must be a unit vector, but has length %v (acceptable delta is +-%v)",
			normal, length, acceptableDelta,
		)
	}

	return HitRecord{t, normal, exterior}
}

type Hittable interface {
	// Hit returns whether the ray hits the Hittable within the range
	// [tMin,tMax] along the ray. If hit is false, HitRecord is not valid.
	Hit(ray Ray, tMin float64, tMax float64) (hit bool, record HitRecord)
}

type Sphere struct {
	Center Vec3
	Radius float64
}

func (s Sphere) Hit(ray Ray, tMin float64, tMax float64) (bool, HitRecord) {
	if s.Radius < 0 {
		log.Panicf("Sphere radius cannot be negative")
	}

	// We can tell whether a ray hits the sphere by considering the following
	// quadratic equation:
	//
	// (t^2)(d * d) - 2(d * Z)t + (Z * Z - r^2) = 0
	//
	// derived from (C - (Q + td)) * (C - (Q + td)) = r^2
	//
	// Explanation:
	//
	// * is the dot operator
	//
	// Z is (C-Q) where C is the center of the circle and Q is the origin of the
	// ray
	//
	// d is the vector describing the direction of the ray
	//
	// r is the radius of the sphere
	//
	// t is the input of the quadratic. It is used to scale the direction
	// vector of the ray to tell us how far along the ray we are. When t
	// satisfies the above equation, the ray has hit the sphere.
	//
	// We can test how many roots there are to this equation by just calculating the
	// discriminant. If it's less than 0, then there are no real solutions to the
	// equation, which means that the ray does not hit the sphere. Otherwise there are
	// one or two solutions, so the ray DOES hit.
	d := ray.Direction
	Z := s.Center.Subtract(ray.Origin)
	a := d.Dot(d)
	// TODO: There's an optimization we can do by factoring out -2 from b.
	b := d.Scale(-2).Dot(Z)
	c := Z.Dot(Z) - (s.Radius * s.Radius)
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, HitRecord{}
	}

	t := (-b - math.Sqrt(discriminant)) / (2 * a)
	if t < tMin || t > tMax {
		// try the other possible root
		t = (-b + math.Sqrt(discriminant)) / (2 * a)
		if t < tMin || t > tMax {
			// still out of the acceptable range
			return false, HitRecord{}
		}
	}
	hitPoint := ray.Direction.Scale(t)
	outwardNormal := hitPoint.Subtract(s.Center).Divide(s.Radius)
	return true, NewHitRecord(ray, t, outwardNormal)
}

type World []Hittable

func (w World) Hit(ray Ray, tMin float64, tMax float64) (bool, HitRecord) {
	hitAnything := false
	closest := tMax
	var closestRecord HitRecord
	for _, object := range w {
		if object == nil {
			log.Panicf("how hard is it to not add a nil value to world?")
		}
		if hit, record := object.Hit(ray, tMin, closest); hit {
			closest = record.T
			closestRecord = record
			hitAnything = true
		}
	}
	return hitAnything, closestRecord
}

func main() {
	camera := NewCamera(400, 16./9., 100)
	world := make(World, 0, 3)
	world = append(world, Sphere{Vec3{0, 0, -1}, 0.5})
	world = append(world, Sphere{Vec3{0, -100.5, -1}, 100})
	camera.Render(os.Stdout, world)
}
