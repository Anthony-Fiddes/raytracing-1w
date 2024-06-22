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

func (v Vec3) Hadamard(other Vec3) Vec3 {
	v.X *= other.X
	v.Y *= other.Y
	v.Z *= other.Z
	return v
}

func (v Vec3) Cross(other Vec3) Vec3 {
	x := v.Y*other.Z - v.Z*other.Y
	y := v.Z*other.X - v.X*other.Z
	z := v.X*other.Y - v.Y*other.X
	return New(x, y, z)
}

func (v Vec3) Divide(factor float64) Vec3 {
	return v.Scale(1. / factor)
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vec3) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) UnitVector() Vec3 {
	length := v.Length()
	if length == 1 {
		return v
	}
	return v.Divide(length)
}

func (v Vec3) Dot(other Vec3) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

func Random() Vec3 {
	return New(rand.Float64(), rand.Float64(), rand.Float64())
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
	for {
		p := RandomRange(-1, 1)
		// We care about length < 1, but length^2 < 1^2 also holds and we can
		// avoid a square root.
		if p.LengthSquared() < 1 {
			return p.UnitVector()
		}
	}
}

func RandomUnitHemisphere(normal Vec3) Vec3 {
	result := RandomUnit()
	if result.Dot(normal) > 0. {
		return result
	}
	return result.Scale(-1)
}

func IsNearZero(v Vec3) bool {
	reallySmall := 1e-8
	xOk := math.Abs(v.X) < reallySmall
	yOk := math.Abs(v.Y) < reallySmall
	zOk := math.Abs(v.Z) < reallySmall
	return xOk && yOk && zOk
}
