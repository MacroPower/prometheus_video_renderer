local prometheus_video_renderer = import 'lib/prometheus_video_renderer.libsonnet';
local prometheus_video_renderer_8 = import 'lib/prometheus_video_renderer_8.libsonnet';
local prometheus_video_renderer_8_audio = import 'lib/prometheus_video_renderer_8_audio.libsonnet';

local timeseries_overrides_grayscale = import 'lib/timeseries_overrides_grayscale.libsonnet';
local timeseries_overrides_rgb = import 'lib/timeseries_overrides_rgb.libsonnet';

local targets_grayscale = import 'lib/targets_grayscale.libsonnet';
local targets_rgb = import 'lib/targets_rgb.libsonnet';

{
  'prometheus-video-renderer.json':
    prometheus_video_renderer,

  'prometheus-video-renderer-8.json':
    prometheus_video_renderer_8.new(
      'pvr-dash-8-rgb',
      'Prometheus Video Renderer (8.0)',
      targets_rgb,
      timeseries_overrides_rgb,
    ),
  'prometheus-video-renderer-8-grayscale.json':
    prometheus_video_renderer_8.new(
      'pvr-dash-8-grayscale',
      'Prometheus Video Renderer (8.0) Grayscale',
      targets_grayscale,
      timeseries_overrides_grayscale,
    ),

  'prometheus-video-renderer-8-audio.json':
    prometheus_video_renderer_8_audio.new(
      'pvr-dash-8-rgb-audio',
      'Prometheus Video Renderer (8.0) +Audio',
      targets_rgb,
      timeseries_overrides_rgb,
    ),
  'prometheus-video-renderer-8-audio-grayscale.json':
    prometheus_video_renderer_8_audio.new(
      'pvr-dash-8-grayscale-audio',
      'Prometheus Video Renderer (8.0) Grayscale +Audio',
      targets_grayscale,
      timeseries_overrides_grayscale,
    ),
}
