package main

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"

	"github.com/Anthony-Fiddes/raytracing-1w/vec"
)

type CameraOpts struct {
	Width              int
	AspectRatio        float64
	VerticalFOVDegrees float64
	SamplesPerPixel    int
	MaxBounces         int
	Position           Vec3
	LookAt             Vec3
	// Up can be used to determine the sideways tilt of the camera
	Up Vec3
}

// camera is an object in the world
type camera struct {
	// height is the number of pixels up/down
	height int
	// Width is the number of pixels left/right
	viewport viewport
	CameraOpts
	// these camera vectors are unit vectors
	cameraUp    Vec3
	cameraRight Vec3
	cameraBack  Vec3
}

// viewport represents the image that the camera captures.
//
// An increasing x goes to the right and increasing y goes down.
type viewport struct {
	width            float64
	height           float64
	center           Vec3
	widthVector      Vec3
	heightVector     Vec3
	pixelDeltaX      Vec3
	pixelDeltaY      Vec3
	upperLeft        Vec3
	firstPixelCenter Vec3
}

func NewCamera(opts CameraOpts) camera {
	const (
		defaultWidth           = 400
		defaultFOV             = 90
		defaultAspectRatio     = 16. / 9.
		defaultSamplesPerPixel = 100
		defaultMaxBounces      = 50
	)

	var (
		defaultUp     = vec.New(0, 1, 0)
		defaultLookAt = vec.New(0, 0, -1)
	)

	if opts.Width < 0 {
		panic("width cannot be <= 0")
	} else if opts.Width == 0 {
		opts.Width = defaultWidth
	}

	if opts.AspectRatio < 0 {
		panic("aspect ratio cannot be <= 0")
	} else if opts.AspectRatio == 0 {
		opts.AspectRatio = defaultAspectRatio
	}

	height := int(float64(opts.Width) / opts.AspectRatio)
	if height == 0 {
		height = 1
	}

	if opts.VerticalFOVDegrees < 0 || 180 < opts.VerticalFOVDegrees {
		panic("Vertical FOV Degrees must be between 1 and 179")
	} else if opts.VerticalFOVDegrees == 0 {
		opts.VerticalFOVDegrees = defaultFOV
	}

	if opts.SamplesPerPixel < 0 {
		panic("samples per pixel cannot be <= 0")
	} else if opts.SamplesPerPixel == 0 {
		opts.SamplesPerPixel = defaultSamplesPerPixel
	}

	if opts.MaxBounces < 0 {
		panic("maxBounces cannot be <= 0")
	} else if opts.MaxBounces == 0 {
		opts.MaxBounces = defaultMaxBounces
	}

	var emptyVec Vec3
	if opts.Position == emptyVec && opts.Position == opts.LookAt {
		opts.LookAt = defaultLookAt
	}
	if opts.Position == opts.LookAt {
		panic("cameraPosition cannot be the same as lookAt")
	}

	if opts.Up == emptyVec {
		opts.Up = defaultUp
	}
	if opts.Up.Dot(opts.LookAt.Subtract(opts.Position)) == 1 {
		panic("the Up vector and the vector between the CameraPosition and the LookAt position cannot be parallel")
	}

	cameraBack := opts.Position.Subtract(opts.LookAt).UnitVector()
	cameraRight := opts.Up.Cross(cameraBack)
	cameraUp := cameraBack.Cross(cameraRight)

	viewport := newViewport(
		opts.Width, height, opts.LookAt,
		opts.Position, opts.VerticalFOVDegrees,
		cameraUp, cameraRight, cameraBack,
	)
	return camera{height, viewport, opts, cameraUp, cameraRight, cameraBack}
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func newViewport(
	width int, height int, center Vec3,
	lookFrom Vec3, verticalFOVDegrees float64,
	cameraUp Vec3, cameraRight Vec3, cameraBack Vec3,
) viewport {
	verticalFOVRads := toRadians(verticalFOVDegrees)
	// I don't think this calculation makes sense at 180 degrees or more, since
	// you can no longer draw a straight line between the two vectors.
	focalLength := center.Subtract(lookFrom).Length()
	viewHeight := math.Tan(verticalFOVRads/2) * 2 * focalLength
	viewWidth := viewHeight * float64(width) / float64(height)
	widthVector := cameraRight.Scale(viewWidth)
	heightVector := cameraUp.Scale(-viewHeight)
	pixelDeltaX := widthVector.Divide(float64(width))
	pixelDeltaY := heightVector.Divide(float64(height))
	upperLeft := lookFrom.Subtract(cameraBack.Scale(focalLength)).Subtract(widthVector.Divide(2)).Subtract(heightVector.Divide(2))
	firstPixelCenter := upperLeft.Add(pixelDeltaX.Divide(2)).Add(pixelDeltaY.Divide(2))

	return viewport{
		viewWidth, viewHeight, center, widthVector,
		heightVector, pixelDeltaX, pixelDeltaY,
		upperLeft, firstPixelCenter,
	}
}

func (c camera) Render(w io.Writer, world Hittable) {
	fmt.Fprintf(w, "P3\n%d %d\n255\n", c.Width, c.height)
	for j := 0; j < c.height; j++ {
		fmt.Fprintf(os.Stderr, "\rScanlines remaining: %d ", c.height-j)
		for i := 0; i < c.Width; i++ {
			var pixel Color
			for range c.SamplesPerPixel {
				sampleXOffset := rand.Float64() - 0.5
				sampleYOffset := rand.Float64() - 0.5
				yPixelCenter := c.viewport.firstPixelCenter.Add(c.viewport.pixelDeltaY.Scale(float64(j) + sampleYOffset))
				sampleCenter := yPixelCenter.Add(c.viewport.pixelDeltaX.Scale(float64(i) + sampleXOffset))
				rayDirection := sampleCenter.Subtract(c.Position)
				ray := Ray{c.Position, rayDirection}
				pixel.Vec = pixel.Vec.Add(ray.Color(world, 0.001, math.Inf(1), c.MaxBounces).Vec)
			}
			pixel.Vec = pixel.Vec.Divide(float64(c.SamplesPerPixel))
			fmt.Printf(toPPM(pixel))
		}
	}
	fmt.Fprint(os.Stderr, "\rDone.                    \n")
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

func linearToGamma(component float64) float64 {
	if component > 0 {
		return math.Sqrt(component)
	}
	return 0
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
