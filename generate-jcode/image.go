package main

import (
	"image"
	"image/color"
	"math"
)

// Threshold applies a binary threshold (with optional inversion)
func Threshold(img image.Image, t uint8, invert bool) *image.Gray {
	bounds := img.Bounds()
	gray := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grayVal := uint8((r*299 + g*587 + b*114 + 500) / 1000 >> 8)
			if invert {
				grayVal = 255 - grayVal
			}
			if grayVal > t {
				gray.SetGray(x, y, color.Gray{255})
			} else {
				gray.SetGray(x, y, color.Gray{0})
			}
		}
	}
	return gray
}

// GuoHallThinning applies the Guo–Hall thinning algorithm
func GuoHallThinning(img *image.Gray) *image.Gray {
	b := img.Bounds()
	w, h := b.Max.X, b.Max.Y
	out := image.NewGray(b)

	// Copy initial binary image (0 or 1)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.GrayAt(x, y).Y > 0 {
				out.SetGray(x, y, color.Gray{255})
			} else {
				out.SetGray(x, y, color.Gray{0})
			}
		}
	}

	changed := true
	for changed {
		changed = false
		for step := 0; step < 2; step++ {
			toRemove := make([][2]int, 0)
			for y := 1; y < h-1; y++ {
				for x := 1; x < w-1; x++ {
					p1 := out.GrayAt(x, y).Y
					if p1 == 0 {
						continue
					}

					// Neighbors p2..p9 clockwise
					p2 := out.GrayAt(x, y-1).Y
					p3 := out.GrayAt(x+1, y-1).Y
					p4 := out.GrayAt(x+1, y).Y
					p5 := out.GrayAt(x+1, y+1).Y
					p6 := out.GrayAt(x, y+1).Y
					p7 := out.GrayAt(x-1, y+1).Y
					p8 := out.GrayAt(x-1, y).Y
					p9 := out.GrayAt(x-1, y-1).Y

					bVals := [9]int{
						boolToInt(p1 > 0),
						boolToInt(p2 > 0),
						boolToInt(p3 > 0),
						boolToInt(p4 > 0),
						boolToInt(p5 > 0),
						boolToInt(p6 > 0),
						boolToInt(p7 > 0),
						boolToInt(p8 > 0),
						boolToInt(p9 > 0),
					}

					// Count nonzero neighbors
					n := bVals[1] + bVals[2] + bVals[3] + bVals[4] + bVals[5] + bVals[6] + bVals[7] + bVals[8]
					if n < 2 || n > 6 {
						continue
					}

					// Count 0->1 transitions in neighborhood
					A := 0
					for i := 1; i <= 8; i++ {
						if bVals[i%8+1] == 1 && bVals[i] == 0 {
							A++
						}
					}
					if A != 1 {
						continue
					}

					// Step 1 or 2 conditions
					if step == 0 {
						if bVals[1]*bVals[3]*bVals[5] != 0 {
							continue
						}
						if bVals[3]*bVals[5]*bVals[7] != 0 {
							continue
						}
					} else {
						if bVals[1]*bVals[3]*bVals[7] != 0 {
							continue
						}
						if bVals[1]*bVals[5]*bVals[7] != 0 {
							continue
						}
					}

					toRemove = append(toRemove, [2]int{x, y})
				}
			}
			if len(toRemove) > 0 {
				changed = true
				for _, xy := range toRemove {
					out.SetGray(xy[0], xy[1], color.Gray{0})
				}
			}
		}
	}

	return out
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func EdgeDetection(img image.Image, threshold float64) *image.Gray {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	gray := image.NewGray(bounds)

	// Sobel operator kernels
	gx := [3][3]int{
		{-1, 0, 1},
		{-2, 0, 2},
		{-1, 0, 1},
	}
	gy := [3][3]int{
		{-1, -2, -1},
		{0, 0, 0},
		{1, 2, 1},
	}

	// Convert input to grayscale for easier processing
	src := image.NewGray(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			grayVal := uint8((r*299 + g*587 + b*114 + 500) / 1000 >> 8)
			src.SetGray(x, y, color.Gray{grayVal})
		}
	}

	// Apply Sobel operator
	for y := 1; y < h-1; y++ {
		for x := 1; x < w-1; x++ {
			var sumX, sumY int
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pix := int(src.GrayAt(x+kx, y+ky).Y)
					sumX += gx[ky+1][kx+1] * pix
					sumY += gy[ky+1][kx+1] * pix
				}
			}
			magnitude := math.Sqrt(float64(sumX*sumX + sumY*sumY))
			if magnitude > threshold {
				magnitude = 255
			} else {
				magnitude = 0
			}
			gray.SetGray(x, y, color.Gray{uint8(magnitude)})
		}
	}

	return gray
}
