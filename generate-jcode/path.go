package main

import (
	"image"
	"image/color"
	"math"
)

type Point struct{ X, Y int }

// Extract all white pixels (255)
func ExtractWhitePixels(img *image.Gray) []Point {
	points := []Point{}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y > 128 {
				points = append(points, Point{x, y})
			}
		}
	}
	return points
}

// Greedy nearest-neighbor path builder
func BuildPaths(points []Point) [][]Point {
	if len(points) == 0 {
		return nil
	}

	path := []Point{}
	jumps := []int{}
	visited := make([]bool, len(points))
	current := 0

	path = append(path, points[current])
	visited[current] = true

	for len(path) < len(points) {
		bestIdx := -1
		bestDist := math.MaxFloat64

		for i, p := range points {
			if visited[i] {
				continue
			}
			dx := float64(p.X - points[current].X)
			dy := float64(p.Y - points[current].Y)
			dist := math.Sqrt(dx*dx + dy*dy)

			// Prefer local moves (especially within 1 diagonal)
			if dist < bestDist {
				bestDist = dist
				bestIdx = i
			}
		}

		if bestIdx == -1 {
			break
		}

		visited[bestIdx] = true
		current = bestIdx
		path = append(path, points[current])
		if len(path) > 1 {
			this := path[len(path)-1]
			prev := path[len(path)-2]
			if !nextToEachOther(this, prev) {
				jumps = append(jumps, len(path)-1)
			}
		}
	}
	jumps = append(jumps, len(path))

	allPaths := [][]Point{}

	lastPathStart := 0
	for _, nextPathStart := range jumps {
		allPaths = append(allPaths, path[lastPathStart:nextPathStart])
		lastPathStart = nextPathStart
	}

	return allPaths
}

// True if pixels touch or are diagonal
func nextToEachOther(pt1, pt2 Point) bool {
	if math.Abs(float64(pt1.X-pt2.X)) > 1 {
		return false
	}
	if math.Abs(float64(pt1.Y-pt2.Y)) > 1 {
		return false
	}
	return true
}

// Visualize path (optional)
func DrawPath(img *image.Gray, path []Point) *image.Gray {
	out := image.NewGray(img.Bounds())
	for i := 0; i < len(path)-1; i++ {
		p := path[i]
		out.SetGray(p.X, p.Y, color.Gray{255})
	}
	return out
}
