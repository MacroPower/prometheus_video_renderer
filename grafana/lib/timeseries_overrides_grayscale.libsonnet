[
  {
    matcher: {
      id: 'byName',
      options: 'y%s' % l,
    },
    properties: [
      {
        id: 'color',
        value: {
          fixedColor: 'rgba(%s, %s, %s, 1)' % [l, l, l],
          mode: 'fixed',
        },
      },
    ],
  }
  for l in std.range(0, 255)
]
