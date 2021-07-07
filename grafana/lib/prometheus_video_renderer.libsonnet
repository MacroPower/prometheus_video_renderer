local grafana = import 'github.com/grafana/grafonnet-lib/grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;
local videoHeight = 100;

local videoPanel =
  graphPanel.new(
    'Video',
    min=0,
    max=videoHeight * 3,
    fill=0,
    linewidth=2,
    interval='1s',
    transparent=false,
    legend_show=false,
    shared_tooltip=false,
    datasource='Prometheus',
  )
  .addTarget(
    prometheus.target(
      'r',
      legendFormat='r{{l}}',
      intervalFactor=1,
    )
  )
  .addTarget(
    prometheus.target(
      'g',
      legendFormat='g{{l}}',
      intervalFactor=1,
    )
  )
  .addTarget(
    prometheus.target(
      'b',
      legendFormat='b{{l}}',
      intervalFactor=1,
    )
  ) + {
    seriesOverrides+: [
      {
        alias: 'r%s' % l,
        color: 'rgb(%s, 0, 0)' % l,
      }
      for l in std.range(0, 255)
    ] + [
      {
        alias: 'g%s' % l,
        color: 'rgb(0, %s, 0)' % l,
      }
      for l in std.range(0, 255)
    ] + [
      {
        alias: 'b%s' % l,
        color: 'rgb(0, 0, %s)' % (l * 2),
      }
      for l in std.range(0, 127)
    ],
  };

dashboard.new(
  'Prometheus Video Renderer',
  uid='pvr-dash',
  timezone='utc',
  schemaVersion=16,
)
.addTemplate(
  grafana.template.datasource(
    'PROMETHEUS_DS',
    'prometheus',
    'Prometheus',
    hide='label',
  )
)
.addPanel(
  videoPanel, gridPos={
    x: 0,
    y: 0,
    w: 24,
    h: 25,
  }
)
