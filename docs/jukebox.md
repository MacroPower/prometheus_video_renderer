# Jukebox

## Installation

Install with:

```text
go get -u github.com/MacroPower/prometheus_video_renderer/cmd/prometheus_jukebox
```

## How to use

Note that this requires several labels that describe playback behavior, and you need to make sure that the record and playback `name` and `start-time` are aligned (technically you could start playback anywhere in the track).

### Recording

1. Obtain a WAV file, e.g. `cantina.wav`
1. Mix to mono. (Stereo should be possible but I've not implemented it yet.)
1. Downsample as needed, it should always work due to metadata in the labels.
1. Run `prometheus_jukebox record` to generate metrics. (See below for usage.)
1. Loop over generated metrics and send them to `promtool tsdb create-blocks-from openmetrics`
   - Helper script [here](scripts/load.ps1)
1. Run `docker compose up`
1. Proceed to playback if you so choose.

```text
$ prometheus_jukebox record --help

Usage of record:
  -file string
        WAV file to source
  -name string
        Name of the track
  -scrape-interval int
        Frequency in seconds at which samples are written (default 1)
  -start-time int
        The starting timestamp (Unix MS) of the track
```

### Playback

You can stream recorded audio directly from Prometheus. Samples are broken into
small chunks that are queried consecutively. Playback is stopped when the metric
is absent or when you manually stop the program.

```text
$ prometheus_jukebox playback --help

Usage of playback:
  -chunk-size duration
        Amount of time to query at once (default 2h0m0s)
  -name string
        Name of the track
  -prometheus-url string
        Prometheus URL
  -start-time int
        The starting timestamp (Unix MS) of the track
```

You can also visualize the playback in Grafana, by using
[grafana-image-renderer-cli](https://github.com/MacroPower/grafana-image-renderer-sdk-go)
sequencer to create a timelapse of the waveform in Prometheus that matches with
playback. You'll need to use a few specific arguments:

> Where `space = total duration of the metric inside Prometheus in seconds` ...

- `interval` is `space / ( (length of track in seconds) * (desired fps) )`.
- `start-padding` should be a negative duration equal to the desired view size.
- `end-padding` should be `interval * -1`.
- `end-frame` is `space / interval` OR `( space + (start padding in seconds) ) / interval`.

Example with 60s track, 30fps, start: 1420070400000, end: 1421393445495, 24h view:

```shell
$ grafana-image-renderer-cli sequence \
    --api-url=http://localhost:3000 \
    --api-key-or-basic-auth=admin:admin \
    --dashboard=ZpUQVIg7k \
    --start-time=1420070400000 \
    --frame-interval=735s \
    --start-padding=-24h \
    --end-padding=-735s \
    --frames=1-1800
```
