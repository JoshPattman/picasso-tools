package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/JoshPattman/jcode"
)

func main() {
	inputImageName := flag.String("input", "", "Input image path")
	outputName := flag.String("output", "result", "Output directory for various artifacts and result")
	threshold := flag.Uint("threshold", 128, "Threshold value (0–255)")
	bwMode := flag.String("bw-mode", "threshold", "How to convert the image to B&W pixels, either threshold or edge-detect")
	invert := flag.Bool("invert", false, "Invert grayscale before thresholding")
	yStart := flag.Float64("y-start", 8.0, "The Y position of the bottom of the image")
	height := flag.Float64("height", 8.0, "The height of the image in units")
	speed := flag.Float64("speed", 5.0, "The speed of the toolhead in units/s")
	pointDistance := flag.Float64("point-dist", 0.25, "The distance between waypoints")
	pathStartDelay := flag.Int("start-delay", 1000, "The delay in ms to wait at the start of each path")
	pathEndDelay := flag.Int("end-delay", 1000, "The delay in ms to wait at the end of each path")
	flag.Parse()

	if *inputImageName == "" {
		failf("Please specify an input image")
	}

	inputImage, err := loadImage(*inputImageName)
	if err != nil {
		failf("could not load input image: %v", err)
	}

	err = os.MkdirAll(*outputName, os.ModePerm)
	if err != nil {
		failf("could not create output directory: %v", err)
	}

	var thresholdImage *image.Gray
	if *bwMode == "threshold" {
		thresholdImage := Threshold(inputImage, uint8(*threshold), *invert)
		err = saveImage(thresholdImage, filepath.Join(*outputName, "threshold.png"))
		if err != nil {
			failf("could not save threshold image: %v", err)
		}
		fmt.Println("Created threshold")
	} else {
		thresholdImage = EdgeDetection(inputImage, float64(*threshold))
		err = saveImage(thresholdImage, filepath.Join(*outputName, "edge-detection.png"))
		if err != nil {
			failf("could not save edge detection image: %v", err)
		}
		fmt.Println("Created edge detection")
	}

	skelImage := GuoHallThinning(thresholdImage)
	err = saveImage(skelImage, filepath.Join(*outputName, "skeleton.png"))
	if err != nil {
		failf("could not save skeleton image: %v", err)
	}
	fmt.Println("Created skeleton")

	pointCloud := ExtractWhitePixels(skelImage)
	fmt.Println("Created point cloud")

	paths := BuildPaths(pointCloud)
	fmt.Println("Built point path")

	scaleFactor := *height / float64(skelImage.Bounds().Dy())
	xOffset := -scaleFactor * float64(skelImage.Bounds().Dx()) / 2
	yOffset := *yStart
	code := []jcode.Instruction{jcode.Speed{Speed: *speed}}
	lastPoint := jcode.Waypoint{XPos: 1000000}
	for _, path := range paths {
		for i, pt := range path {
			thisPoint := jcode.Waypoint{
				XPos: float64(pt.X)*scaleFactor + xOffset,
				YPos: float64(skelImage.Bounds().Dy()-pt.Y)*scaleFactor + yOffset,
			}
			if (i == 0) || (i == len(path)-1) || (jcode.Dist(lastPoint, thisPoint) >= *pointDistance) {
				code = append(code, thisPoint)
				lastPoint = thisPoint
			}
			if i == 0 {
				code = append(code, jcode.Delay{Duration: time.Millisecond * time.Duration(*pathStartDelay)}, jcode.Pen{Mode: jcode.PenDown})
			}
		}
		code = append(code, jcode.Delay{Duration: time.Millisecond * time.Duration(*pathEndDelay)}, jcode.Pen{Mode: jcode.PenUp})
	}
	err = saveJCode(code, filepath.Join(*outputName, "path.jcode"))
	if err != nil {
		failf("could not save JCode: %v", err)
	}
	fmt.Println("Exported JCode")
}

func failf(f string, args ...any) {
	fmt.Printf(f+"\n", args...)
	os.Exit(1)
}

func saveImage(img image.Image, to string) error {
	outFile, err := os.Create(to)
	if err != nil {
		return errors.Join(errors.New("could not create output file"), err)
	}
	defer outFile.Close()
	return png.Encode(outFile, img)
}

func saveJCode(jc []jcode.Instruction, to string) error {
	outFile, err := os.Create(to)
	if err != nil {
		return errors.Join(errors.New("could not create output file"), err)
	}
	defer outFile.Close()
	enc := jcode.NewEncoder(outFile)
	return enc.Write(jc...)
}

func loadImage(from string) (image.Image, error) {
	file, err := os.Open(from)
	if err != nil {
		return nil, errors.Join(errors.New("could not open image file"), err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, errors.Join(errors.New("could not decode image"), err)
	}

	return img, nil
}
