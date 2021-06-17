package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/MacroPower/prometheus_video_renderer/pkg/backfiller"
)

type writeFunc func(b *bytes.Buffer, img image.Image, timestamp time.Time, x int, y int, invY int, sliceSize int)

var (
	projectName     = flag.String("project", "", "The name of the project")
	framesLocation  = flag.String("frames-location", "frames", "Location of png frames")
	metricsLocation = flag.String("metrics-location", "metrics", "Location to write metrics")
	framesPerFile   = flag.Int("frames-per-file", 120, "Number of frames to include in each metrics file")
	writeMode       = flag.String("mode", "bitmap", "One of: [bitmap, grayscale, rgb]")
	lightThreshold  = flag.Int("bitmap-light-threshold", 127, "Brightness required to write a sample (1-255)")
	scrapeInterval  = flag.Int("scrape-interval", 1, "The frequency at which new samples are written")
	startTimeMs     = flag.Int64("start-time", 0, "The starting timestamp (Unix MS) of the render")
	frameDuration   = flag.Duration(
		"frame-duration",
		5*time.Minute,
		"The max duration that can be used to write samples."+
			" Must be greater than the horizontal resolution times the scrape interval.",
	)

	writeFuncs = map[string]writeFunc{
		"bitmap":    writeBitmap,
		"grayscale": writeGrayscale,
		"rgb":       writeRGB,
	}
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeBitmap(b *bytes.Buffer, img image.Image, timestamp time.Time, x, y, invY, sliceSize int) {
	c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
	if int(c.Y) > *lightThreshold {
		for i := 0; i < sliceSize; i += *scrapeInterval {
			sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()
			if b.Len() == 0 {
				b.WriteString(backfiller.Help(*projectName))
			}
			b.WriteString(fmt.Sprintf(`%s{y="%d"} %d %d%s`, *projectName, y, invY, sampleTs, "\n"))
		}
	}
}

func writeGrayscale(b *bytes.Buffer, img image.Image, timestamp time.Time, x, y, invY, sliceSize int) {
	c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
	for i := 0; i < sliceSize; i += *scrapeInterval {
		sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()
		if b.Len() == 0 {
			b.WriteString(backfiller.Help(*projectName))
		}
		b.WriteString(fmt.Sprintf(`r{y="%d",l="%d"} %d %d%s`, y, c.Y, invY-0, sampleTs, "\n"))
	}
}

func writeRGB(b *bytes.Buffer, img image.Image, timestamp time.Time, x, y, invY, sliceSize int) {
	c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	invY = invY * 3
	for i := 0; i < sliceSize; i += *scrapeInterval {
		sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()
		if b.Len() == 0 {
			b.WriteString(backfiller.Help("r"))
			b.WriteString(backfiller.Help("g"))
			b.WriteString(backfiller.Help("b"))
		}
		b.WriteString(fmt.Sprintf(`r{y="%d",l="%d"} %d %d%s`, y, c.R>>0, invY-0, sampleTs, "\n"))
		b.WriteString(fmt.Sprintf(`g{y="%d",l="%d"} %d %d%s`, y, c.G>>0, invY-1, sampleTs, "\n"))
		b.WriteString(fmt.Sprintf(`b{y="%d",l="%d"} %d %d%s`, y, c.B>>1, invY-2, sampleTs, "\n"))
	}
}

func main() {
	flag.Parse()

	startTime := backfiller.FromUnixMs(*startTimeMs)

	framesDir := filepath.Join(*framesLocation, *projectName)
	metricsDir := filepath.Join(*metricsLocation, *projectName)
	os.Mkdir(metricsDir, 0664)

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

		sliceSize := int(frameDuration.Seconds()) / (img.Bounds().Max.X - img.Bounds().Min.X)
		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			invY := img.Bounds().Max.Y - y
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				timestamp := startTime.Add(time.Duration(sliceSize*x) * time.Second)
				writeFuncs[*writeMode](b, img, timestamp, x, y, invY, sliceSize)
			}
		}
		startTime = startTime.Add(*frameDuration)

		if b.Len() > 0 && i != 0 && i%*framesPerFile == 0 || i == len(frames)-1 {
			filesWritten++
			b.WriteString("# EOF")
			filename := fmt.Sprintf("out%04d", filesWritten)
			fmt.Printf("Writing: %s\n", filename)
			f, err := os.Create(filepath.Join(metricsDir, filename))
			check(err)
			_, err = b.WriteTo(f)
			check(err)
			f.Close()
			b.Reset()
		}
	}
}
