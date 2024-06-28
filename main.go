package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/Anthony-Fiddes/raytracing-1w/vec"
)

type Vec3 = vec.Vec3

// X, Y, and Z represent red, green, and blue values. They are floats between 0 and 1
type Color struct{ Vec Vec3 }

var (
	white = newColor(1, 1, 1)
	black = newColor(0, 0, 0)
	red   = newColor(1, 0, 0)
	green = newColor(0, 1, 0)
	blue  = newColor(0, 0, 1)
)

func newColor(r, g, b float64) Color {
	return Color{Vec3{X: r, Y: g, Z: b}}
}

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

func reflectanceProbability(cosine float64, refractionIndex float64) float64 {
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
	if canRefract && rand.Float64() > reflectanceProbability(cosTheta, refractionIndex) {
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

func renderRandomSpheres(opts CameraOpts) {
	world := make(World, 0)
	boundary := vec.New(4, 0.2, 0)
	glassMat := &Dielectric{1.5}
	for a := -11; a < 11; a++ {
		for b := -11; b < 11; b++ {
			chooseMat := rand.Float64()
			center := vec.New(float64(a)+0.9*rand.Float64(), 0.2, float64(b)+0.9*rand.Float64())

			if center.Subtract(boundary).Length() <= 0.9 {
				continue
			}

			if chooseMat < 0.8 {
				albedo := Color{vec.Random().Hadamard(vec.Random())}
				material := Lambertian{albedo}
				world = append(world, Sphere{center, 0.2, material})
			} else if chooseMat < 0.95 {
				albedo := Color{vec.RandomRange(0.5, 1)}
				// fuzz in range [0, 0.5)
				fuzz := (rand.Float64() + 1) / 4
				material := Metal{albedo, fuzz}
				world = append(world, Sphere{center, 0.2, material})
			} else {
				world = append(world, Sphere{center, 0.2, glassMat})
			}
		}
	}

	world = append(world, Sphere{vec.New(0, -1000, 0), 1000, Lambertian{newColor(0.5, 0.5, 0.5)}})
	world = append(world, Sphere{vec.New(0, 1, 0), 1, glassMat})
	world = append(world, Sphere{vec.New(-4, 1, 0), 1, Lambertian{newColor(0.4, 0.2, 0.1)}})
	world = append(world, Sphere{vec.New(4, 1, 0), 1, Metal{newColor(0.7, 0.6, 0.5), 0}})

	camera := NewCamera(opts)
	camera.Render(world)
}

func renderSimpleScene(opts CameraOpts) {
	ground := Sphere{vec.New(0, -100.5, -1), 100, Lambertian{newColor(0.8, 0.8, 0)}}
	middleSphere := Sphere{vec.New(0, 0, -1.2), 0.5, Lambertian{newColor(0.1, 0.2, 0.5)}}
	leftSphere := Sphere{vec.New(-1., 0, -1.), 0.5, Dielectric{1.5}}
	leftSphereInside := Sphere{vec.New(-1., 0, -1.), 0.4, Dielectric{1. / 1.5}}
	rightSphere := Sphere{vec.New(1., 0, -1.), 0.5, Metal{newColor(0.8, 0.6, 0.2), 1}}
	world := make(World, 0, 3)
	world = append(world, ground)
	world = append(world, middleSphere)
	world = append(world, leftSphere)
	world = append(world, leftSphereInside)
	world = append(world, rightSphere)

	camera := NewCamera(opts)
	camera.Render(world)
}

func main() {
	scene := flag.String("scene", "simple", "random | simple")
	flag.Parse()

	if *scene != "random" && *scene != "simple" {
		fmt.Fprintln(os.Stderr, "scene must be 'random' or 'simple'")
		fmt.Fprintln(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if *scene == "random" {
		renderRandomSpheres(
			CameraOpts{
				AspectRatio:        16. / 9.,
				Width:              300,
				SamplesPerPixel:    100,
				MaxBounces:         50,
				VerticalFOVDegrees: 20,
				Position:           vec.New(13, 2, 3),
				LookAt:             vec.New(0, 0, 0),
				Up:                 vec.New(0, 1, 0),
				DefocusAngle:       0.6,
				FocusDist:          10,
			},
		)
	} else if *scene == "simple" {
		renderSimpleScene(
			CameraOpts{
				Position:           vec.New(-2, 2, 1),
				LookAt:             vec.New(0, 0, -1),
				VerticalFOVDegrees: 20,
				DefocusAngle:       10,
				FocusDist:          3.4,
			},
		)
	}
}
