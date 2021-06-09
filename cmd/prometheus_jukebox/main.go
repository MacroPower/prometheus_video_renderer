package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/MacroPower/prometheus_video_renderer/pkg/backfiller"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	bwav "github.com/faiface/beep/wav"
	prometheus "github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/youpy/go-wav"
)

var (
	trackName      = flag.String("name", "", "Name of the track")
	filePath       = flag.String("file", "", "WAV file to source")
	scrapeInterval = flag.Int("scrape-interval", 1, "Frequency at which samples are written")
	mode           = flag.String("mode", "", "One of: [read, write]")
	chunkSize      = flag.Duration("chunk-size", 2*time.Hour, "Amount of time to query at once")
	startLoc       = flag.String("start-location", "", "Start location of the track (RFC3339)")
)

func main() {
	flag.Parse()

	st, err := time.Parse(time.RFC3339, *startLoc)
	check(err)

	switch *mode {
	case "read":
		read(st)
	case "write":
		write(st)
	default:
		panic("unsupported mode")
	}
}

func read(startTime time.Time) {
	sr := beep.SampleRate(44100)
	speaker.Init(sr, sr.N(time.Second/10))

	var queue Queue
	done := make(chan bool)
	speaker.Play(&queue)

	client, err := prometheus.NewClient(prometheus.Config{Address: "http://localhost:9090"})
	check(err)

	q := v1.NewAPI(client)

	i := 0
	for {
		data, wavFormat, ok := getAudioChunk(q, startTime, i)
		if !ok {
			break
		}
		i++

		b := new(bytes.Buffer)
		writer := wav.NewWriter(b, uint32(len(data)), 2, wavFormat.SampleRate, wavFormat.BitsPerSample)
		writer.WriteSamples(data)

		streamer, format, err := bwav.Decode(b)
		check(err)

		resampled := beep.Resample(4, format.SampleRate, sr, streamer)

		speaker.Lock()
		queue.Add(resampled)
		speaker.Unlock()
	}
	queue.Add(beep.Callback(func() {
		done <- true
	}))

	<-done
}

func write(startTime time.Time) {
	file, _ := os.Open(*filePath)

	reader := wav.NewReader(file)
	defer file.Close()

	b := new(bytes.Buffer)
	for {
		samples, err := reader.ReadSamples()
		if err == io.EOF {
			break
		}
		check(err)

		format, err := reader.Format()
		check(err)

		for _, sample := range samples {
			if b.Len() == 0 {
				b.WriteString(backfiller.Help(*trackName))
			}

			b.WriteString(
				fmt.Sprintf(
					`%s{audio_format="%d",bits_per_sample="%d",block_align="%d",byte_rate="%d",sample_rate="%d"} %d %d%s`,
					*trackName,
					format.AudioFormat,
					format.BitsPerSample,
					format.BlockAlign,
					format.ByteRate,
					format.SampleRate,
					reader.IntValue(sample, 0),
					startTime.Unix(),
					"\n",
				),
			)
			startTime = startTime.Add(time.Duration(*scrapeInterval) * time.Second)
		}
	}
	b.WriteString("# EOF")
	f, err := os.Create(*trackName)
	check(err)
	_, err = b.WriteTo(f)
	check(err)
}

func getAudioChunk(q v1.API, startTime time.Time, n int) (samples []wav.Sample, format wav.WavFormat, ok bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	value, _, err := q.QueryRange(ctx, *trackName, v1.Range{
		Start: startTime.Add(*chunkSize * time.Duration(n+0)),
		End:   startTime.Add(*chunkSize * time.Duration(n+1)),
		Step:  time.Duration(*scrapeInterval) * time.Second * 1,
	})
	check(err)
	queryType := value.Type()

	if queryType == model.ValMatrix {
		matrixVal := value.(model.Matrix)
		for i, series := range matrixVal {
			if i == 0 {
				format = metricToFormat(series.Metric)
			}
			for _, elem := range series.Values {
				samples = append(samples, wav.Sample{Values: [2]int{int(elem.Value), int(elem.Value)}})
			}
		}

		if len(samples) == 0 {
			return []wav.Sample{}, format, false
		}

		fmt.Printf("Read %d samples, chunk %d\n", len(samples), n)
		return samples, format, true
	}

	return []wav.Sample{}, format, false
}

func metricToFormat(metric model.Metric) wav.WavFormat {
	audioFormat, err := strconv.ParseUint(string(metric["audio_format"]), 10, 16)
	check(err)

	sampleRate, err := strconv.ParseUint(string(metric["sample_rate"]), 10, 32)
	check(err)

	byteRate, err := strconv.ParseUint(string(metric["byte_rate"]), 10, 32)
	check(err)

	blockAlign, err := strconv.ParseUint(string(metric["block_align"]), 10, 16)
	check(err)

	bitsPerSample, err := strconv.ParseUint(string(metric["bits_per_sample"]), 10, 16)
	check(err)

	return wav.WavFormat{
		AudioFormat:   uint16(audioFormat),
		NumChannels:   uint16(1),
		SampleRate:    uint32(sampleRate),
		ByteRate:      uint32(byteRate),
		BlockAlign:    uint16(blockAlign),
		BitsPerSample: uint16(bitsPerSample),
	}
}

type Queue struct {
	streamers []beep.Streamer
	count     int
}

func (q *Queue) Add(streamers ...beep.Streamer) {
	q.streamers = append(q.streamers, streamers...)
}

func (q *Queue) Stream(samples [][2]float64) (n int, ok bool) {
	filled := 0
	for filled < len(samples) {
		if len(q.streamers) == 0 {
			for i := range samples[filled:] {
				samples[i][0] = 0
				samples[i][1] = 0
			}
			break
		}

		n, ok := q.streamers[0].Stream(samples[filled:])
		if !ok {
			q.streamers = q.streamers[1:]

			fmt.Printf("Finished playing chunk %d\n", q.count)
			q.count++
		}
		filled += n
	}
	return len(samples), true
}

func (q *Queue) Err() error {
	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
