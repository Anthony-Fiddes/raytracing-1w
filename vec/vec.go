package vec

import (
	"math"
	"math/rand"
)

type Vec3 struct {
	X float64
	Y float64
	Z float64
}

func New(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
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

func Random() Vec3 {
	return Vec3{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64()}
}

func randFloatRange(min, max float64) float64 {
	return (max-min)*rand.Float64() + min
}

func RandomRange(min, max float64) Vec3 {
	return Vec3{
		randFloatRange(min, max),
		randFloatRange(min, max),
		randFloatRange(min, max),
	}
}

func RandomUnit() Vec3 {
	p := RandomRange(-1, 1)
	return p.UnitVector()
}

func RandomUnitHemisphere(normal Vec3) Vec3 {
	result := RandomUnit()
	if result.Dot(normal) > 0 {
		return result
	}
	return result.Scale(-1)
}
