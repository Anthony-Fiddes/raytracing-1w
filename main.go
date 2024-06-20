package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/Anthony-Fiddes/raytracing-1w/vec"
)

type Vec3 = vec.Vec3

// X, Y, and Z represent red, green, and blue values. They are floats between 0 and 1
type Color struct{ Vec Vec3 }

func newColor(r, g, b float64) Color {
	return Color{Vec3{X: r, Y: g, Z: b}}
}

var (
	white = newColor(1, 1, 1)
	black = newColor(0, 0, 0)
	red   = newColor(1, 0, 0)
	green = newColor(0, 1, 0)
	blue  = newColor(0, 0, 1)
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

func linearToGamma(component float64) float64 {
	if component > 0 {
		return math.Sqrt(component)
	}
	return 0
}

func toPPM(c Color) string {
	c.assertValid()
	gammaR := linearToGamma(c.R())
	gammaG := linearToGamma(c.G())
	gammaB := linearToGamma(c.B())
	scaledR := int(255.999 * gammaR)
	scaledG := int(255.999 * gammaG)
	scaledB := int(255.999 * gammaB)
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
	MaxDepth        int
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
				pixel.Vec = pixel.Vec.Add(ray.Color(world, 0.001, math.Inf(1), c.MaxDepth).Vec)
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

func NewCamera(width int, aspectRatio float64, samplesPerPixel int, maxDepth int) camera {
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
	widthVector := vec.New(viewWidth, 0, 0)
	heightVector := vec.New(0, -viewHeight, 0)
	pixelDeltaX := widthVector.Divide(float64(width))
	pixelDeltaY := heightVector.Divide(float64(height))
	focalLength := 1.0
	upperLeft := center.Subtract(vec.New(0, 0, focalLength)).Subtract(widthVector.Divide(2)).Subtract(heightVector.Divide(2))
	firstPixelCenter := upperLeft.Add(pixelDeltaX.Divide(2)).Add(pixelDeltaY.Divide(2))
	viewport := viewport{
		viewWidth, viewHeight, widthVector,
		heightVector, pixelDeltaX, pixelDeltaY,
		upperLeft, firstPixelCenter,
	}
	return camera{height, width, aspectRatio, center, focalLength, viewport, samplesPerPixel, maxDepth}
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

func (r Ray) Color(h Hittable, tMin float64, tMax float64, depth int) Color {
	if depth <= 0 {
		// no more light is gathered
		return black
	}

	if hit, record := h.Hit(r, tMin, tMax); hit {
		scattered, newRay, attenuation := record.Material.Scatter(record)
		if scattered {
			colorVec := newRay.Color(h, tMin, tMax, depth-1).Vec.Hadamard(attenuation.Vec)
			return Color{colorVec}
		}
		// ray was absorbed
		return black
	}

	unitDirection := r.Direction.UnitVector()
	// unit vector's y ranges from [-1, 1], so we transform the range to [0, 1]
	// to do a linear interpolation and get a nice gradient from white to blue
	a := 0.5*unitDirection.Y + 1
	lightBlue := newColor(0.5, 0.7, 1)
	colorVec := white.Vec.Scale(1 - a).Add(lightBlue.Vec.Scale(a))
	return Color{colorVec}
}

type HitRecord struct {
	Ray Ray
	// Factor to scale ray by to get hit point
	T float64
	// Normal vector at the hit point. It points against the ray and is
	// expected to be a unit vector.
	Normal Vec3
	// Exterior is whether the ray hit the geometry from the outside or the
	// inside
	Exterior bool
	// Where the ray hit the geometry
	HitPoint Vec3
	// Material of the hit geometry
	Material Material
}

// outwardNormal is a normal pointing out of the hit geometry. It must be a unit
// vector.
func NewHitRecord(ray Ray, t float64, outwardNormal Vec3, hitPoint Vec3, mat Material) HitRecord {
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
	const acceptableDelta = 0.02
	if math.Abs(length-1) > acceptableDelta {
		log.Panicf(
			"Normal %+v must be a unit vector, but has length %v (acceptable delta is +-%v)",
			normal, length, acceptableDelta,
		)
	}

	return HitRecord{ray, t, normal, exterior, hitPoint, mat}
}

type Material interface {
	// Scatter returns whether the material scatters the ray and details about
	// the new ray. If scattered is false, the ray was absorbed and scatteredRay and
	// attenuation should be ignored.
	Scatter(record HitRecord) (scattered bool, scatteredRay Ray, attenuation Color)
}

type Lambertian struct {
	Albedo Color
}

func (l Lambertian) Scatter(record HitRecord) (scattered bool, scatteredRay Ray, attenuation Color) {
	scatterDirection := record.Normal.Add(vec.RandomUnit())
	if vec.IsNearZero(scatterDirection) {
		scatterDirection = record.Normal
	}
	newRay := Ray{record.HitPoint, scatterDirection}
	return true, newRay, l.Albedo
}

type Metal struct {
	Albedo Color
	// Fuzz is a proportion that determines how much the direction of reflected
	// rays might vary from a theoretically perfect reflection.
	Fuzz float64
}

func reflect(direction Vec3, normal Vec3) Vec3 {
	b := normal.Scale(direction.Dot(normal))
	return direction.Subtract(b.Scale(2))
}

func (m Metal) Scatter(record HitRecord) (scattered bool, scatteredRay Ray, attenuation Color) {
	if m.Fuzz > 1 || m.Fuzz < 0 {
		log.Panicf("fuzz must be in the range [0,1]")
	}
	scatterDirection := reflect(record.Ray.Direction, record.Normal).UnitVector()
	scatterDirection = scatterDirection.Add(vec.RandomUnit().Scale(m.Fuzz))
	if scatterDirection.Dot(record.Normal) <= 0 {
		return false, Ray{}, Color{}
	}
	newRay := Ray{record.HitPoint, scatterDirection}
	return true, newRay, m.Albedo
}

type Dielectric struct {
	// Refractive index in vacuum or air. To simulate one material in another,
	// use the ratio of the materials' refractive index to that of the
	// surrounding medium.
	RefractionIndex float64
}

func refract(direction Vec3, normal Vec3, refractionIndex float64) Vec3 {
	cosTheta := min(direction.Scale(-1).Dot(normal), 1.0)
	rayOutPerpendicular := normal.Scale(cosTheta).Add(direction).Scale(refractionIndex)
	parallelFactor := -math.Sqrt(math.Abs(1.0 - rayOutPerpendicular.LengthSquared()))
	rayOutParallel := normal.Scale(parallelFactor)
	return rayOutParallel.Add(rayOutPerpendicular)
}

func reflectance(cosine float64, refractionIndex float64) float64 {
	// Schlick's approximation
	r0 := (1 - refractionIndex) / (1 + refractionIndex)
	r0 = r0 * r0
	return r0 + (1-r0)*math.Pow(1-cosine, 5)
}

func (d Dielectric) Scatter(record HitRecord) (scattered bool, scatteredRay Ray, attenuation Color) {
	refractionIndex := d.RefractionIndex
	if record.Exterior {
		refractionIndex = 1. / refractionIndex
	}
	unitDirection := record.Ray.Direction.UnitVector()
	cosTheta := min(unitDirection.Scale(-1).Dot(record.Normal), 1.0)
	sinTheta := math.Sqrt(1. - (cosTheta * cosTheta))
	canRefract := refractionIndex*sinTheta <= 1.
	var scatterDirection Vec3
	if canRefract || reflectance(cosTheta, refractionIndex) <= rand.Float64() {
		scatterDirection = refract(
			unitDirection,
			record.Normal,
			refractionIndex,
		)
	} else {
		scatterDirection = reflect(
			unitDirection,
			record.Normal,
		)
	}
	newRay := Ray{record.HitPoint, scatterDirection}
	return true, newRay, white
}

type Hittable interface {
	// Hit returns whether the ray hits the Hittable within the range
	// [tMin,tMax] along the ray. If hit is false, HitRecord is not valid.
	Hit(ray Ray, tMin float64, tMax float64) (hit bool, record HitRecord)
}

type Sphere struct {
	Center   Vec3
	Radius   float64
	Material Material
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
	// Z is (C-Q) where C is the center of the sphere and Q is the origin of the
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
	b := d.Dot(Z) * -2.
	c := Z.Dot(Z) - (s.Radius * s.Radius)
	discriminant := b*b - 4*a*c
	if discriminant < 0 {
		return false, HitRecord{}
	}

	root := (-b - math.Sqrt(discriminant)) / (2. * a)
	if root <= tMin || tMax <= root {
		// try the other possible root
		root = (-b + math.Sqrt(discriminant)) / (2. * a)
		if root <= tMin || tMax <= root {
			// still out of the acceptable range
			return false, HitRecord{}
		}
	}
	hitPoint := ray.At(root)
	outwardNormal := hitPoint.Subtract(s.Center).Divide(s.Radius)
	return true, NewHitRecord(ray, root, outwardNormal, hitPoint, s.Material)
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
	camera := NewCamera(400, 16./9., 100, 50)
	ground := Sphere{vec.New(0, -100.5, -1), 100, Lambertian{newColor(0.8, 0.8, 0)}}
	middleSphere := Sphere{vec.New(0, 0, -1.2), 0.5, Lambertian{newColor(0.1, 0.2, 0.5)}}
	leftSphere := Sphere{vec.New(-1., 0, -1.), 0.5, Dielectric{1. / 1.33}}
	rightSphere := Sphere{vec.New(1., 0, -1.), 0.5, Metal{newColor(0.8, 0.6, 0.2), 1}}
	world := make(World, 0, 3)
	world = append(world, ground)
	world = append(world, middleSphere)
	world = append(world, leftSphere)
	world = append(world, rightSphere)
	camera.Render(os.Stdout, world)
}
