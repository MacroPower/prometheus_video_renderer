{
  new(uid, name, targets=[], overrides=[]):: {
    annotations: {
      list: [],
    },
    editable: true,
    gnetId: null,
    graphTooltip: 0,
    id: null,
    links: [],
    panels: [
      {
        datasource: 'Prometheus',
        fieldConfig: {
          defaults: {
            color: {
              mode: 'palette-classic',
            },
            custom: {
              axisLabel: '',
              axisPlacement: 'auto',
              barAlignment: 0,
              drawStyle: 'line',
              fillOpacity: 0,
              gradientMode: 'none',
              hideFrom: {
                legend: false,
                tooltip: false,
                viz: false,
              },
              lineInterpolation: 'linear',
              lineWidth: 2,
              pointSize: 5,
              scaleDistribution: {
                type: 'linear',
              },
              showPoints: 'never',
              spanNulls: false,
              stacking: {
                group: 'A',
                mode: 'none',
              },
              thresholdsStyle: {
                mode: 'off',
              },
            },
            mappings: [],
            max: 300,
            min: 0,
            thresholds: {
              mode: 'absolute',
              steps: [
                {
                  color: 'green',
                  value: null,
                },
                {
                  color: 'red',
                  value: 80,
                },
              ],
            },
            unit: 'short',
          },
          overrides: overrides,
        },
        gridPos: {
          h: 25,
          w: 24,
          x: 0,
          y: 0,
        },
        id: 2,
        interval: '1s',
        links: [],
        options: {
          legend: {
            calcs: [],
            displayMode: 'hidden',
            placement: 'bottom',
          },
          tooltip: {
            mode: 'single',
          },
        },
        pluginVersion: '8.0.0',
        repeat: null,
        targets: targets,
        timeFrom: null,
        timeShift: null,
        title: 'Video',
        type: 'timeseries',
      },
    ],
    refresh: false,
    schemaVersion: 30,
    style: 'dark',
    tags: [],
    templating: {
      list: [
        {
          current: {
            selected: false,
            text: 'Prometheus',
            value: 'Prometheus',
          },
          description: null,
          'error': null,
          hide: 1,
          includeAll: false,
          label: null,
          multi: false,
          name: 'PROMETHEUS_DS',
          options: [],
          query: 'prometheus',
          refresh: 1,
          regex: '',
          skipUrlSync: false,
          type: 'datasource',
        },
      ],
    },
    time: {
      from: 'now-6h',
      to: 'now',
    },
    timepicker: {
      refresh_intervals: [
        '5s',
        '10s',
        '30s',
        '1m',
        '5m',
        '15m',
        '30m',
        '1h',
        '2h',
        '1d',
      ],
      time_options: ['5m', '15m', '1h', '6h', '12h', '24h', '2d', '7d', '30d'],
    },
    timezone: 'utc',
    title: name,
    uid: uid,
    version: 1,
  },
}
