# Renderer

## Installation

Install with:

```text
go get -u github.com/MacroPower/prometheus_video_renderer/cmd/prometheus_video_renderer
```

```text
$ prometheus_video_renderer --help

Usage of prometheus_video_renderer:
  -bitmap-light-threshold int
        Brightness required to write a sample (1-255) (default 127)
  -frame-duration duration
        The max duration that can be used to write samples. Must be greater than the horizontal resolution times the scrape interval. (default 5m0s)
  -frames-location string
        Location of png frames (default "frames")
  -frames-per-file int
        Number of frames to include in each metrics file (default 120)
  -metrics-location string
        Location to write metrics (default "metrics")
  -mode string
        One of: [bitmap, grayscale, rgb] (default "bitmap")
  -project string
        The name of the project
  -scrape-interval int
        The frequency at which new samples are written (default 1)
  -start-time int
        The starting timestamp (Unix MS) of the render
  -write-wav string
        Optional .WAV file to write alongside video
```

## How to use

### Setup

Obtain a video, e.g. `bad_apple.mkv`.

Figure out the output resolution, you should scale 1:1 with height
being the number of time series you want. If you're using Grafana, 100
height works well. On Prometheus, 135 height works well.

Get an ordered PNG sequence from your video, scaled to the resolution
you want. You can also use handbrake to scale the video (just remove
the `-vf 'scale=...'` argument).

```bash
ffmpeg -i 'bad_apple.mkv' -vf 'scale=180:135' -vsync 0 'out%06d.png'
```

Consider adding 1px white bars to each frame to avoid y axis changes.
_You do not have to do this if you're using Grafana_ since you can
easily set a min/max.

```bash
ffmpeg -i 'out%06d.png' -vf 'crop=in_w:in_h-1:0:1,pad=iw+0:ih+1:0:1:#FFFFFF@1,format=rgb24' -y 'out%06d.png'
ffmpeg -i 'out%06d.png' -vf 'crop=in_w:in_h-1:0:-1,pad=iw+0:ih+1:0:-1:#FFFFFF@1,format=rgb24' -y 'out%06d.png'
```

Store all frames in one folder under `frames/your_project_name`.

### Backfill Metrics

Run `prometheus_video_renderer` to generate metrics. You'll need to
pass `--mode` (see the [readme](../README.md) for an explanation of
each), `--project` which is the same as `your_project_name` from the
previous section, and `--start-time` which you can easily grab from a
Grafana dashboard using UTC.

To also include audio, you must pass a WAV file with a matching sample
rate to `--write-wav`. You can match the sample rate with
`sox in.wav -r 99 out.wav dither`, where 99 is the number of samples
per frame multiplied by fps. Note that if you want to perfectly match
the video you'll need to match the total audio and video samples
exactly, so the actual sample rate of the audio may need to be
fractional. All this will reduce audio quality by a good amount but it
should still be listenable through the [jukebox](jukebox.md). There
also may be some clicking if you use a blanking interval due, I
believe this is due to Prometheus adding an extra sample every frame.

Loop over generated metrics and send them to
`promtool tsdb create-blocks-from openmetrics`. Helper script
[here](../scripts/load.ps1).

### Record Frame Data

Run `docker compose up`. If you want to improve Grafana render speed
(at the cost of additional cpu/memory), increase
`RENDERING_CLUSTERING_MAX_CONCURRENCY`. Optionally, wait for
Prometheus to compact. This can help prevent OOMs, timeouts and such.

Now you can begin recording frames from the Prometheus or Grafana UI.
You will want to size the graph such that the aspect ratio is the same
as the original video (or is at least somewhat close).

If you're using Prometheus, it's fast enough that you can use AHK.
(Mediocre example in the scripts directory.)

If you intend to use Grafana, I highly recommend getting my
[grafana-image-renderer-cli](https://github.com/MacroPower/grafana-image-renderer-sdk-go)
as the sequencer package makes automating the rendering process a lot
faster and easier.

- `frame-interval` should be set to the same as `frame-duration`.
- `end-padding` removes any excess time that could not be filled with
  your `frame-duration`. For example, a video with 128px width, with a
  5m duration and 1s scrape interval, will use 256 seconds which is 44
  seconds less than the duration. Thus, the `end-padding` should be
  -44s. (In general using multiples of 5 is better than giving an
  exact `frame-duration` because it will keep the Grafana UI more
  static.)

```text
$ grafana-image-renderer-cli sequence \
    --api-url=http://localhost:3000 \
    --api-key-or-basic-auth=admin:admin \
    --dashboard=pvr-dash-8 \
    --start-time=1256428800000 \
    --frame-interval=5m \
    --end-padding=-44s \
    --frames=1-5
```

### Generate Video

`cd` to wherever you stored your frames and generate the video file
with `ffmpeg -framerate 30 -i '%06d.png'` + any of the following:

- Fast output: `-c:v libx264 -pix_fmt yuv420p out.mp4`
- High quality output: `-c:v libx264rgb -pix_fmt rgb24 -preset veryslow -crf 0 -qp 0 out.mp4`
- Merge with wav: `-i input.wav -c:v copy -c:a aac -c:v libx264rgb -pix_fmt rgb24 out.mp4`
- Merge with audio from original video: `-i input.mp4 -c copy -map 0:0 -map 1:1 -c:v libx264rgb -pix_fmt rgb24 out.mp4`
- Upscale: `-c:v libx264rgb -pix_fmt rgb24 -vf scale=2560:2200:flags=neighbor out.mp4`

I highly recommend upscaling to ~4k for YouTube videos, because it
results in YouTube using much less lossy compression on your video.
The result will be much less muddy and the colors will be
significantly better.

## Issues

If you run into any problems, please
[create an issue](https://github.com/MacroPower/prometheus_video_renderer/issues/new)
and I will try my best to help.
