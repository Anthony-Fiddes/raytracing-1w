package main

import (
	"io"
	"testing"

	"github.com/Anthony-Fiddes/raytracing-1w/vec"
)

var simpleSceneCameraOpts = CameraOpts{
	Out:                io.Discard,
	Log:                io.Discard,
	AspectRatio:        16. / 9.,
	Width:              50,
	SamplesPerPixel:    100,
	MaxBounces:         50,
	Position:           vec.New(-2, 2, 1),
	LookAt:             vec.New(0, 0, -1),
	VerticalFOVDegrees: 20,
	DefocusAngle:       10,
	FocusDist:          3.4,
	Parallel:           false,
}

func BenchmarkRenderSimple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		renderSimpleScene(simpleSceneCameraOpts)
	}
}

func BenchmarkRenderSimpleParallel(b *testing.B) {
	opts := simpleSceneCameraOpts
	opts.Parallel = true
	for i := 0; i < b.N; i++ {
		renderSimpleScene(opts)
	}
}

var randomSpheresSceneCameraOpts = CameraOpts{
	Out:                io.Discard,
	Log:                io.Discard,
	AspectRatio:        16. / 9.,
	Width:              50,
	SamplesPerPixel:    100,
	MaxBounces:         50,
	VerticalFOVDegrees: 20,
	Position:           vec.New(13, 2, 3),
	LookAt:             vec.New(0, 0, 0),
	Up:                 vec.New(0, 1, 0),
	DefocusAngle:       0.6,
	FocusDist:          10,
	Parallel:           false,
}

func BenchmarkRenderRandomSpheres(b *testing.B) {
	for i := 0; i < b.N; i++ {
		renderRandomSpheres(randomSpheresSceneCameraOpts)
	}
}

func BenchmarkRenderRandomSpheresParallel(b *testing.B) {
	opts := randomSpheresSceneCameraOpts
	opts.Parallel = true
	for i := 0; i < b.N; i++ {
		renderRandomSpheres(opts)
	}
}
