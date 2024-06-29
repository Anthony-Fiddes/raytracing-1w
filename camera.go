package main

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"

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
	// FocusDist is the distance from the camera to a plane of perfect focus
	FocusDist float64
	// DefocusAngle is the degrees
	DefocusAngle float64
	Out          io.Writer
	Log          io.Writer
	// Parallel specifies whether the render uses multiple threads or not
	Parallel bool
}

// camera is an object in the world
type camera struct {
	// height is the number of pixels up/down
	height   int
	viewport viewport
	CameraOpts
	// these camera vectors are unit vectors
	upVec                Vec3
	rightVec             Vec3
	backVec              Vec3
	defocusDiskWidthVec  Vec3
	defocusDiskHeightVec Vec3
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
		defaultOut    = os.Stdout
		defaultLog    = os.Stderr
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

	if opts.VerticalFOVDegrees < 0 || opts.VerticalFOVDegrees >= 180 {
		panic("Vertical FOV Degrees must be between 1 and 180")
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

	if opts.DefocusAngle < 0 || opts.DefocusAngle >= 180 {
		panic("DefocusAngle must be between 0 and 180")
	}

	if opts.FocusDist < 0 {
		panic("FocusDist must be >= 0")
	} else if opts.FocusDist == 0 {
		opts.FocusDist = opts.LookAt.Subtract(opts.Position).Length()
	}

	if opts.Out == nil {
		opts.Out = defaultOut
	}
	if opts.Log == nil {
		opts.Log = defaultLog
	}

	backVec := opts.Position.Subtract(opts.LookAt).UnitVector()
	rightVec := opts.Up.Cross(backVec)
	upVec := backVec.Cross(rightVec)
	defocusRadius := opts.FocusDist * math.Tan(toRadians(opts.DefocusAngle/2))
	defocusDiskWidthVec := rightVec.Scale(defocusRadius)
	defocusDiskHeightVec := upVec.Scale(defocusRadius)

	camera := camera{
		height: height, CameraOpts: opts,
		upVec: upVec, rightVec: rightVec, backVec: backVec,
		defocusDiskWidthVec:  defocusDiskWidthVec,
		defocusDiskHeightVec: defocusDiskHeightVec,
	}

	camera.viewport = calculateViewport(camera)
	return camera
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func calculateViewport(c camera) viewport {
	verticalFOVRads := toRadians(c.VerticalFOVDegrees)
	// I don't think this calculation makes sense at 180 degrees or more, since
	// you can no longer draw a straight line between the two vectors.
	viewHeight := math.Tan(verticalFOVRads/2) * 2 * c.FocusDist
	viewWidth := viewHeight * float64(c.Width) / float64(c.height)

	widthVector := c.rightVec.Scale(viewWidth)
	heightVector := c.upVec.Scale(-viewHeight)
	pixelDeltaX := widthVector.Divide(float64(c.Width))
	pixelDeltaY := heightVector.Divide(float64(c.height))
	upperLeft := c.Position.Subtract(c.backVec.Scale(c.FocusDist)).Subtract(widthVector.Divide(2)).Subtract(heightVector.Divide(2))
	firstPixelCenter := upperLeft.Add(pixelDeltaX.Divide(2)).Add(pixelDeltaY.Divide(2))

	return viewport{
		viewWidth, viewHeight, c.LookAt, widthVector,
		heightVector, pixelDeltaX, pixelDeltaY,
		upperLeft, firstPixelCenter,
	}
}

func (c camera) Render(world Hittable) {
	if c.Parallel {
		c.renderParallel(world)
		return
	}
	c.render(world)
}

func (c camera) render(world Hittable) {
	fmt.Fprintf(c.Out, "P3\n%d %d\n255\n", c.Width, c.height)
	for j := 0; j < c.height; j++ {
		fmt.Fprintf(c.Log, "\rScanlines remaining: %d ", c.height-j)
		for i := 0; i < c.Width; i++ {
			var pixel Color
			for range c.SamplesPerPixel {
				rayOrigin := c.Position
				if c.DefocusAngle > 0 {
					nudge := vec.RandomDisk()
					rayOrigin = rayOrigin.Add(c.defocusDiskWidthVec.Scale(nudge.X))
					rayOrigin = rayOrigin.Add(c.defocusDiskHeightVec.Scale(nudge.Y))
				}

				sampleXOffset := rand.Float64() - 0.5
				sampleYOffset := rand.Float64() - 0.5
				yPixelCenter := c.viewport.firstPixelCenter.Add(c.viewport.pixelDeltaY.Scale(float64(j) + sampleYOffset))
				sampleCenter := yPixelCenter.Add(c.viewport.pixelDeltaX.Scale(float64(i) + sampleXOffset))
				rayDirection := sampleCenter.Subtract(rayOrigin)
				ray := Ray{rayOrigin, rayDirection}
				pixel.Vec = pixel.Vec.Add(ray.Color(world, 0.001, math.Inf(1), c.MaxBounces).Vec)
			}
			pixel.Vec = pixel.Vec.Divide(float64(c.SamplesPerPixel))
			writePPM(pixel, c.Out)
		}
	}
	fmt.Fprint(c.Log, "\rDone.                    \n")
}

func (c camera) renderParallel(world Hittable) {
	// using a worker pool here because starting a goroutine for every sample
	// was actually slower than the single-threaded version.
	numWorkers := runtime.GOMAXPROCS(0)
	// pixelPositions must be buffered as large as the number of samples per
	// pixel or we'll deadlock when the main routine sends on it.
	pixelPositions := make(chan pos, c.SamplesPerPixel)
	samples := make(chan Vec3, c.SamplesPerPixel)
	for i := 0; i < numWorkers; i++ {
		go sampleWorker(c, world, pixelPositions, samples)
	}

	fmt.Fprintf(c.Out, "P3\n%d %d\n255\n", c.Width, c.height)
	for j := 0; j < c.height; j++ {
		fmt.Fprintf(c.Log, "\rScanlines remaining: %d ", c.height-j)
		for i := 0; i < c.Width; i++ {
			for range c.SamplesPerPixel {
				pixelPositions <- pos{i, j}
			}

			var pixel Color
			for range c.SamplesPerPixel {
				next := <-samples
				pixel.Vec = pixel.Vec.Add(next)
			}
			pixel.Vec = pixel.Vec.Divide(float64(c.SamplesPerPixel))
			writePPM(pixel, c.Out)
		}
	}
	close(pixelPositions)
	close(samples)
	fmt.Fprint(c.Log, "\rDone.                    \n")
}

func sampleWorker(c camera, world Hittable, pixelPositions <-chan pos, samples chan<- Vec3) {
	for pos := range pixelPositions {
		rayOrigin := c.Position
		if c.DefocusAngle > 0 {
			nudge := vec.RandomDisk()
			rayOrigin = rayOrigin.Add(c.defocusDiskWidthVec.Scale(nudge.X))
			rayOrigin = rayOrigin.Add(c.defocusDiskHeightVec.Scale(nudge.Y))
		}

		sampleXOffset := rand.Float64() - 0.5
		sampleYOffset := rand.Float64() - 0.5
		yPixelCenter := c.viewport.firstPixelCenter.Add(c.viewport.pixelDeltaY.Scale(float64(pos.j) + sampleYOffset))
		sampleCenter := yPixelCenter.Add(c.viewport.pixelDeltaX.Scale(float64(pos.i) + sampleXOffset))
		rayDirection := sampleCenter.Subtract(rayOrigin)
		ray := Ray{rayOrigin, rayDirection}
		samples <- ray.Color(world, 0.001, math.Inf(1), c.MaxBounces).Vec
	}
}

type pos struct {
	i, j int
}

func writePPM(c Color, w io.Writer) {
	c.assertValid()
	gammaR := linearToGamma(c.R())
	gammaG := linearToGamma(c.G())
	gammaB := linearToGamma(c.B())
	scaledR := int(255.999 * gammaR)
	scaledG := int(255.999 * gammaG)
	scaledB := int(255.999 * gammaB)
	fmt.Fprintf(w, "%d %d %d\n", scaledR, scaledG, scaledB)
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

type Hittable interface {
	// Hit returns whether the ray hits the Hittable within the range
	// [tMin,tMax] along the ray. If hit is false, HitRecord is not valid.
	Hit(ray Ray, tMin float64, tMax float64) (hit bool, record HitRecord)
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
