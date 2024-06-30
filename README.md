This is a simple ray tracer I built following ["Ray Tracing in One
Weekend"](https://raytracing.github.io/books/RayTracingInOneWeekend.html).

Here is an image I generated using the code snippet that the authors used to
produce their book cover:

![1080p render with 500 samples per pixel of a bunch of random spheres](result.png)


And here are some benchmarks showing the difference in performance between the
parallel and non-parallel renders:
```
goos: darwin
goarch: arm64
pkg: github.com/Anthony-Fiddes/raytracing-1w
BenchmarkRenderSimple-10                   	     30	159664206 ns/op	   2106 B/op	     14 allocs/op
BenchmarkRenderSimpleParallel-10           	     30	104225461 ns/op	  13358 B/op	     42 allocs/op
BenchmarkRenderRandomSpheres-10            	     30	4738820744 ns/op  60009 B/op	    960 allocs/op
BenchmarkRenderRandomSpheresParallel-10    	     30	808697338 ns/op	  69781 B/op	    985 allocs/op
PASS
ok  	github.com/Anthony-Fiddes/raytracing-1w	180.542s
```

So that's ~1.5x faster for the simple scene and ~5.86x faster for a more
complicated scene that looks like the render above. These benchmarks are
basically shrunken down versions of renders from the book so they don't take
forever to complete a run.

I ran this on an M1 Max with 8 performance cores and 2 efficiency cores, so the
peak theoretical increase would have been somewhere around 8x. Realistically the
performance increase is even less than reported above since there's likely to be
throttling if a render goes long enough.

Another note: while writing the benchmarks, I realized that my implementation
will degrade in performance when the number of samples is not greater than the
number of cores on the system used, but that shouldn't be much of an issue since
using a low number of samples gives a very grainy image. In any case, if I were
to revisit the project it could be interesting to consider a different
concurrency plan or investigate making use of GPU acceleration.
