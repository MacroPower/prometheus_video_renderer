package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/MacroPower/prometheus_video_renderer/pkg/backfiller"
	"github.com/youpy/go-wav"
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
	writeWav        = flag.String("write-wav", "", "Optional .WAV file to write alongside video")
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

			backfiller.WriteHelp(b, *projectName)
			b.WriteString(fmt.Sprintf(`%s{y="%d"} %d %d%s`, *projectName, y, invY, sampleTs, "\n"))
		}
	}
}

func writeGrayscale(b *bytes.Buffer, img image.Image, timestamp time.Time, x, y, invY, sliceSize int) {
	c := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
	for i := 0; i < sliceSize; i += *scrapeInterval {
		sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()

		backfiller.WriteHelp(b, "y", "wav")
		b.WriteString(fmt.Sprintf(`y{y="%d",l="%d"} %d %d%s`, y, c.Y, invY-0, sampleTs, "\n"))
	}
}

func writeRGB(b *bytes.Buffer, img image.Image, timestamp time.Time, x, y, invY, sliceSize int) {
	c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	invY = invY * 3
	for i := 0; i < sliceSize; i += *scrapeInterval {
		sampleTs := timestamp.Add(time.Duration(i) * time.Second).Unix()

		backfiller.WriteHelp(b, "r", "g", "b", "wav")
		b.WriteString(fmt.Sprintf(`r{y="%d",l="%d"} %d %d%s`, y, c.R>>0, invY-0, sampleTs, "\n"))
		b.WriteString(fmt.Sprintf(`g{y="%d",l="%d"} %d %d%s`, y, c.G>>0, invY-1, sampleTs, "\n"))
		b.WriteString(fmt.Sprintf(`b{y="%d",l="%d"} %d %d%s`, y, c.B>>1, invY-2, sampleTs, "\n"))
	}
}

func writeWave(b *bytes.Buffer, timestamp time.Time, reader *wav.Reader, format *wav.WavFormat, samples []wav.Sample) {
	for i, sample := range samples {
		sampleTs := timestamp.Add(time.Duration(i**scrapeInterval) * time.Second)

		backfiller.WriteWaveSample(b, sampleTs, "wav", sample, reader.IntValue(sample, 0), format)
	}
}

func getSamplesInFrame(sliceSize int, img image.Image) int {
	return sliceSize * (img.Bounds().Max.X - img.Bounds().Min.X)
}

func getSliceSize(frameDuration *time.Duration, img image.Image) int {
	return int(frameDuration.Seconds()) / (img.Bounds().Max.X - img.Bounds().Min.X)
}

func main() {
	flag.Parse()

	argErr := false
	if *startTimeMs == 0 {
		fmt.Println("start-time is required")
		argErr = true
	}
	if argErr {
		fmt.Println("Use --help for more information")
		os.Exit(1)
	}

	startTime := backfiller.FromUnixMs(*startTimeMs)

	framesDir := filepath.Join(*framesLocation, *projectName)
	metricsDir := filepath.Join(*metricsLocation, *projectName)
	os.Mkdir(metricsDir, 0664)

	frames, err := ioutil.ReadDir(framesDir)
	if err != nil {
		panic(err)
	}

	var filesWritten int
	var sliceSize int
	var samplesInFrame int
	b := new(bytes.Buffer)

	var reader *wav.Reader
	var format *wav.WavFormat
	var samples []wav.Sample
	if *writeWav != "" {
		file, err := os.Open(*writeWav)
		check(err)

		reader := wav.NewReader(file)
		defer file.Close()

		format, err = reader.Format()
		check(err)

		for {
			s, err := reader.ReadSamples()
			samples = append(samples, s...)
			if err == io.EOF {
				break
			}
			check(err)
		}

		fmt.Printf("Read in %d wave samples\n", len(samples))
	}

	for i, frame := range frames {
		fmt.Printf("Running: %s\n", frame.Name())
		f, err := os.Open(path.Join(framesDir, frame.Name()))
		check(err)
		defer f.Close()

		img, err := png.Decode(f)
		check(err)

		if i == 0 {
			sliceSize = getSliceSize(frameDuration, img)
			samplesInFrame = getSamplesInFrame(sliceSize, img)
			fmt.Printf("Total samples: %d, %d per frame\n", samplesInFrame*len(frames), samplesInFrame)
		} else if sliceSize != getSliceSize(frameDuration, img) {
			fmt.Printf("Inconsistent image size at frame %d\n", i)
			os.Exit(1)
		}

		for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
			invY := img.Bounds().Max.Y - y
			for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
				timestamp := startTime.Add(time.Duration(sliceSize*x) * time.Second)
				writeFuncs[*writeMode](b, img, timestamp, x, y, invY, sliceSize)
			}
		}
		if len(samples) > 0 {
			samplesToWriteStart := samplesInFrame * i
			samplesToWriteEnd := samplesInFrame * (i + 1)
			if samplesToWriteEnd > len(samples) {
				samplesToWriteEnd = len(samples)
			}
			frameSamples := samples[samplesToWriteStart:samplesToWriteEnd]
			writeWave(b, startTime, reader, format, frameSamples)
		}
		startTime = startTime.Add(*frameDuration)

		if b.Len() > 0 && i != 0 && i%*framesPerFile == 0 || i == len(frames)-1 {
			filesWritten++
			backfiller.WriteEnd(b)
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
