package main

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var (
	startTime                  = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	frameDurationSeconds       = 60 * 60 * 1
	scrapeInterval             = 5
	framesPerFile              = 120
	framesDir                  = "frames"
	lightThreshold       uint8 = 255 / 2
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	frames, err := ioutil.ReadDir(framesDir)
	if err != nil {
		panic(err)
	}
	var filesWritten int
	b := new(bytes.Buffer)
	for i, frame := range frames {
		fmt.Printf("Running: %s\n", frame.Name())
		f, err := os.Open(path.Join(framesDir, frame.Name()))
		check(err)
		defer f.Close()

		img, err := png.Decode(f)
		check(err)

		sliceSize := frameDurationSeconds / (img.Bounds().Max.X - img.Bounds().Min.X)
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			invY := img.Bounds().Max.Y - y
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				timestamp := startTime.Add(time.Duration(sliceSize*x) * time.Second)
				c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
				if c.Y > lightThreshold {
					for i := 0; i < sliceSize; i += scrapeInterval {
						sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()
						if b.Len() == 0 {
							b.WriteString("# HELP bad_apple The metric.\n# TYPE bad_apple gauge\n")
						}
						b.WriteString(fmt.Sprintf(`bad_apple{y="%d"} %d %d%s`, y, invY, sampleTs, "\n"))
					}
				}
			}
		}
		startTime = startTime.Add(time.Duration(frameDurationSeconds) * time.Second)

		if b.Len() > 0 && i%framesPerFile == 0 || i == len(frames)-1 {
			filesWritten++
			b.WriteString("# EOF")
			filename := fmt.Sprintf("metrics/out%04d", filesWritten)
			fmt.Printf("Writing: %s\n", filename)
			f, err := os.Create(filename)
			check(err)
			_, err = b.WriteTo(f)
			check(err)
			f.Close()
			b.Reset()
		}
	}
}
